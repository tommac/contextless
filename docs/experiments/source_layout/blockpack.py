"""Verify seekable lossless packing of the pinned, public LongMemEval-S message bodies."""
import argparse
import hashlib
import json
import mmap
import os
from pathlib import Path
import random
import struct
import tempfile
import time
import zlib


PIN = "d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442"
BODY = struct.Struct("<QI")   # uncompressed starting address and UTF-8 length
REF = struct.Struct("<I")    # occurrence -> body ordinal for the deduplicated arm
BLOCK = struct.Struct("<QI32s")  # packed start, packed length and SHA-256
# Header encodes block size, codec, dedup mode, number of bodies and occurrences.
HEADER = struct.Struct("<8sIIQQQ")


def cases(path):
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for data in iter(lambda: source.read(1 << 20), b""):
            digest.update(data)
    if digest.hexdigest() != PIN:
        raise ValueError("unexpected dataset revision")
    decoder = json.JSONDecoder()
    with path.open(encoding="utf-8") as f:
        buf = f.read(1 << 20).lstrip()
        if buf[0] != "[":
            raise ValueError("expected a JSON array")
        buf = buf[1:]
        while True:
            buf = buf.lstrip(" \t\r\n,")
            if buf.startswith("]"):
                return
            try:
                record, end = decoder.raw_decode(buf)
            except json.JSONDecodeError:
                extra = f.read(1 << 20)
                if not extra:
                    raise ValueError("incomplete JSON array")
                buf += extra
                continue
            yield record
            buf = buf[end:]


def archive(path):
    unique = []
    seen = {}
    occurrences = []
    question_count = 0
    for case in cases(path):
        question_count += 1
        for session in case["haystack_sessions"]:
            for msg in session:
                raw = msg["content"].encode()
                index = seen.get(raw)
                if index is None:
                    index = len(unique)
                    seen[raw] = index
                    unique.append(raw)
                occurrences.append(index)
    return unique, occurrences, question_count


def packed_blocks(bodies, block_size):
    pending = bytearray()
    for raw in bodies:
        pending.extend(raw)
        start = 0
        while len(pending) - start >= block_size:
            yield bytes(pending[start : start + block_size])
            start += block_size
        if start:
            del pending[:start]
    if pending:
        yield bytes(pending)


def write_archive(root, unique, refs, block_size, level, dedup):
    def bodies():
        return iter(unique) if dedup else (unique[i] for i in refs)

    nbody = len(unique) if dedup else len(refs)
    index = root / "index"
    blocks_start = HEADER.size + nbody * BODY.size + int(dedup) * len(refs) * REF.size
    expected = hashlib.sha256()
    with index.open("wb") as idx, (root / "payload").open("wb") as out:
        idx.write(HEADER.pack(b"CXPACK01", block_size, level, int(dedup), nbody, len(refs)))
        at = 0
        for raw in bodies():
            idx.write(BODY.pack(at, len(raw)))
            at += len(raw)
            expected.update(raw)
        if dedup:
            for ordinal in refs:
                idx.write(REF.pack(ordinal))
        if idx.tell() != blocks_start:
            raise ValueError("wrong index size")
        for raw in packed_blocks(bodies(), block_size):
            encoded = zlib.compress(raw, level) if level else raw
            idx.write(BLOCK.pack(out.tell(), len(encoded), hashlib.sha256(encoded).digest()))
            out.write(encoded)
    # This 32-byte manifest commitment must be held by a trusted reader.
    (root / "manifest.sha256").write_bytes(hashlib.sha256(index.read_bytes()).digest())
    return expected.digest()


def workload(unique, refs):
    rng = random.Random(4231)
    reads = []
    while len(reads) < 2000:
        pos = rng.randrange(len(refs))
        raw = unique[refs[pos]]
        if len(raw) < 96:
            continue
        off = rng.randrange(len(raw) - 96 + 1)
        reads.append((pos, off, raw[off : off + 96]))
    return reads


def verify_and_read(root, unique, refs, reads, expected_stream):
    with (root / "index").open("rb") as index, (root / "payload").open("rb") as data:
        with mmap.mmap(index.fileno(), 0, access=mmap.ACCESS_READ) as idx:
            if hashlib.sha256(idx).digest() != (root / "manifest.sha256").read_bytes():
                raise ValueError("index digest mismatch")
            tag, bs, codec, dedup, nbody, nrefs = HEADER.unpack_from(idx)
            if tag != b"CXPACK01" or nrefs != len(refs):
                raise ValueError("unexpected header")
            refs_at = HEADER.size + BODY.size * nbody
            blocks_at = refs_at + (REF.size * nrefs if dedup else 0)
            nblocks, extra = divmod(len(idx) - blocks_at, BLOCK.size)
            if extra:
                raise ValueError("invalid block table")
            fd = data.fileno()

            def unpack(block):
                at, size, digest = BLOCK.unpack_from(idx, blocks_at + block * BLOCK.size)
                encoded = os.pread(fd, size, at)
                if hashlib.sha256(encoded).digest() != digest:
                    raise ValueError("damaged block")
                return (zlib.decompress(encoded) if codec else encoded), size

            def locate(pos):
                ordinal = REF.unpack_from(idx, refs_at + REF.size * pos)[0] if dedup else pos
                return BODY.unpack_from(idx, HEADER.size + BODY.size * ordinal)

            # Walk every boundary and map entry as well as the reconstructed stream.
            stream = hashlib.sha256()
            length = 0
            for block in range(nblocks):
                raw, _ = unpack(block)
                if block < nblocks - 1 and len(raw) != bs:
                    raise ValueError("short interior block")
                stream.update(raw)
                length += len(raw)
            if stream.digest() != expected_stream:
                raise ValueError("whole-stream digest mismatch")
            body_lengths = [len(row) for row in unique] if dedup else [len(unique[i]) for i in refs]
            start = 0
            for ordinal, size in enumerate(body_lengths):
                if BODY.unpack_from(idx, HEADER.size + BODY.size * ordinal) != (start, size):
                    raise ValueError("wrong body boundary")
                start += size
            if length != start:
                raise ValueError("length mismatch")
            if dedup and any(REF.unpack_from(idx, refs_at + i * REF.size)[0] != ref for i, ref in enumerate(refs)):
                raise ValueError("wrong occurrence mapping")

            accesses = 0
            touched_raw = 0
            touched_encoded = 0
            t = time.perf_counter()
            for pos, off, want in reads:
                start, size = locate(pos)
                if off + len(want) > size:
                    raise ValueError("span overruns message")
                start += off
                end = start + len(want)
                acc = bytearray()
                for block in range(start // bs, (end - 1) // bs + 1):
                    decoded, packed = unpack(block)
                    low = max(0, start - block * bs)
                    high = min(len(decoded), end - block * bs)
                    acc.extend(decoded[low:high])
                    accesses += 1
                    touched_raw += len(decoded)
                    touched_encoded += packed
                if acc != want:
                    raise ValueError("span differs from original")
            ms = (time.perf_counter() - t) * 1000
            direct = None
            if not codec:
                t = time.perf_counter()
                for pos, off, want in reads:
                    start, _ = locate(pos)
                    if os.pread(fd, len(want), start + off) != want:
                        raise ValueError("direct slice differs from original")
                direct = round((time.perf_counter() - t) * 1000, 3)
    return dict(checked_blocks=nblocks, checked_spans=len(reads), span_ms=round(ms, 3), direct_unhashed_span_ms=direct,
                blocks_opened=accesses, raw_bytes_touched=touched_raw, encoded_bytes_read=touched_encoded)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("data", type=Path)
    p.add_argument("--kib", nargs="+", type=int, default=(4, 16, 64))
    p.add_argument("--levels", type=int, nargs="+", default=(0, 1), choices=range(10))
    args = p.parse_args()
    if min(args.kib) <= 0:
        p.error("block sizes must be positive")
    unique, refs, ncase = archive(args.data)
    reads = workload(unique, refs)
    original_bytes = sum(len(unique[i]) for i in refs)
    unique_bytes = sum(map(len, unique))
    for dedup in (False, True):
        for kib in args.kib:
            for level in args.levels:
                with tempfile.TemporaryDirectory() as tmp:
                    root = Path(tmp)
                    t = time.perf_counter()
                    digest = write_archive(root, unique, refs, kib * 1024, level, dedup)
                    build_ms = (time.perf_counter() - t) * 1000
                    result = verify_and_read(root, unique, refs, reads, digest)
                    payload = (root / "payload").stat().st_size
                    metadata = (root / "index").stat().st_size + (root / "manifest.sha256").stat().st_size
                    print(json.dumps(dict(questions=ncase, messages=len(refs), unique_bodies=len(unique),
                                          bytes_if_unshared=original_bytes, unique_body_bytes=unique_bytes,
                                          dedup=dedup, block_kib=kib, zlib_level=level, packed_bytes=payload,
                                          metadata_bytes=metadata, total_bytes=payload+metadata,
                                          build_ms=round(build_ms, 2), **result)), flush=True)


if __name__ == "__main__":
    main()
