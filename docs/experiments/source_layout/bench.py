"""Compare lossless source layouts in disposable directories. Requires POSIX and Python 3.9+."""
import argparse
from collections import OrderedDict
import hashlib
import json
import os
from pathlib import Path
import random
import sqlite3
import statistics
import struct
import tempfile
import time


REF = struct.Struct("<QI")
PER_SEGMENT = 1024
HASH_BYTES = 32
APPEND_BATCH = 32


def body(seq):
    line = f"Record {seq:08d} records the original café venue, reservation time, and decision. ".encode()
    repeats = 250 if seq % 101 == 0 else 5 + seq % 19
    return line * repeats + b"End."


def frame(seq):
    raw = body(seq)
    return hashlib.sha256(raw).digest() + raw


def elapsed_ms(start):
    return round((time.perf_counter() - start) * 1000, 3)


def sync_dir(path):
    fd = os.open(path, os.O_RDONLY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)


class Files:
    def __init__(self, root):
        self.root = root
        self.n = 0

    def load(self):
        head = self.root / "head"
        self.n = int(head.read_text()) if head.exists() else 0

    def publish(self, size):
        temp = self.root / "head.tmp"
        with temp.open("w") as f:
            f.write(str(size))
            f.flush()
            os.fsync(f.fileno())
        os.replace(temp, self.root / "head")
        sync_dir(self.root)
        self.n = size

    def close(self):
        pass


class Loose(Files):
    """One binary frame per turn, strengthened with a constant-time committed head."""

    def path(self, seq):
        return self.root / f"t{seq:08d}.bin"

    def load(self):
        super().load()
        seq = self.n
        while self.path(seq).exists():
            self.path(seq).unlink()
            seq += 1

    def append(self, frames, crash=False):
        for seq, record in enumerate(frames, self.n):
            with self.path(seq).open("xb") as f:
                f.write(record)
                f.flush()
                os.fsync(f.fileno())
        sync_dir(self.root)
        if crash:
            os._exit(51)
        self.publish(self.n + len(frames))

    def span(self, seq, offset, size):
        with self.path(seq).open("rb") as f:
            return os.pread(f.fileno(), size, HASH_BYTES + offset)

    def records(self):
        for seq in range(self.n):
            yield self.path(seq).read_bytes()


class Segment(Files):
    """Dense positions select one fixed-width offset entry and one source region."""

    def __init__(self, root):
        super().__init__(root)
        self.index = (root / "offsets.idx").open("a+b", buffering=0)
        self.handles = OrderedDict()

    def path(self, seg):
        return self.root / f"{seg:08d}.seg"

    def load(self):
        super().load()
        self.index.truncate(self.n * REF.size)
        if self.n:
            offset, size = self.ref(self.n - 1)
            with self.path((self.n - 1) // PER_SEGMENT).open("r+b") as f:
                f.truncate(offset + size)
        seg = (self.n + PER_SEGMENT - 1) // PER_SEGMENT
        while self.path(seg).exists():
            self.path(seg).unlink()
            seg += 1

    def ref(self, seq):
        return REF.unpack(os.pread(self.index.fileno(), REF.size, seq * REF.size))

    def append(self, frames, crash=False):
        done = 0
        refs = bytearray()
        while done < len(frames):
            seg, used = divmod(self.n + done, PER_SEGMENT)
            batch = frames[done : done + min(PER_SEGMENT - used, len(frames) - done)]
            with self.path(seg).open("ab") as f:
                for record in batch:
                    refs.extend(REF.pack(f.tell(), len(record)))
                    f.write(record)
                f.flush()
                os.fsync(f.fileno())
            done += len(batch)
        self.index.write(refs)
        os.fsync(self.index.fileno())
        sync_dir(self.root)
        if crash:
            os._exit(51)
        self.publish(self.n + len(frames))

    def span(self, seq, offset, size):
        seg = seq // PER_SEGMENT
        if seg not in self.handles:
            if len(self.handles) == 16:
                self.handles.popitem(last=False)[1].close()
            self.handles[seg] = self.path(seg).open("rb", buffering=0)
        self.handles.move_to_end(seg)
        start, length = self.ref(seq)
        if HASH_BYTES + offset + size > length:
            raise ValueError("read past record")
        return os.pread(self.handles[seg].fileno(), size, start + HASH_BYTES + offset)

    def records(self):
        for lo in range(0, self.n, PER_SEGMENT):
            count = min(PER_SEGMENT, self.n - lo)
            entries = os.pread(self.index.fileno(), REF.size * count, lo * REF.size)
            with self.path(lo // PER_SEGMENT).open("rb") as f:
                for offset, size in REF.iter_unpack(entries):
                    if offset != f.tell():
                        raise ValueError("noncontiguous offsets")
                    yield f.read(size)

    def close(self):
        self.index.close()
        for f in self.handles.values():
            f.close()


class Sql:
    def __init__(self, root):
        self.db = sqlite3.connect(root / "records.db")
        self.db.execute("PRAGMA journal_mode=WAL")
        self.db.execute("PRAGMA synchronous=FULL")
        self.db.execute("CREATE TABLE IF NOT EXISTS frames(id INTEGER PRIMARY KEY, raw BLOB NOT NULL)")

    def load(self):
        self.n = self.db.execute("SELECT coalesce(max(id)+1,0) FROM frames").fetchone()[0]

    def append(self, frames, crash=False):
        with self.db:
            self.db.executemany("INSERT INTO frames VALUES(?,?)", enumerate(frames, self.n))
            if crash:
                os._exit(51)
        self.n += len(frames)

    def span(self, seq, offset, size):
        sql = "SELECT substr(raw,?,?) FROM frames WHERE id=?"
        return self.db.execute(sql, (1 + HASH_BYTES + offset, size, seq)).fetchone()[0]

    def records(self):
        for (raw,) in self.db.execute("SELECT raw FROM frames ORDER BY id"):
            yield raw

    def close(self):
        self.db.close()


def probes(n):
    rng = random.Random(1234)
    out = []
    for _ in range(2000):
        seq = rng.randrange(n)
        text = body(seq)
        offset = rng.randrange(len(text) - 96 + 1)
        out.append((seq, offset, text[offset : offset + 96]))
    return out


def verify_replay(db, n, expected):
    digests = hashlib.sha256()
    count = 0
    for rec in db.records():
        digest = hashlib.sha256(rec[HASH_BYTES:]).digest()
        if rec[:HASH_BYTES] != digest:
            raise ValueError("damaged frame")
        digests.update(digest)
        count += 1
    if count != n or digests.digest() != expected.digest():
        raise ValueError("unexpected replay")


def cycle(backend, root, n, reads):
    store = backend(root)
    store.load()
    expected = hashlib.sha256()
    raw_bytes = 0
    for start in range(0, n, 512):
        batch = [frame(j) for j in range(start, min(n, start + 512))]
        for rec in batch:
            raw_bytes += len(rec) - HASH_BYTES
            expected.update(rec[:HASH_BYTES])
        store.append(batch)
    store.close()

    # Each open+append starts with newly opened handles. OS caches remain warm.
    append_times = []
    for start in range(n, n + 256, APPEND_BATCH):
        batch = [frame(seq) for seq in range(start, start + APPEND_BATCH)]
        t = time.perf_counter()
        store = backend(root)
        store.load()
        if store.n != start:
            raise ValueError("wrong committed head")
        store.append(batch)
        store.close()
        append_times.append(elapsed_ms(t))
        for rec in batch:
            expected.update(rec[:HASH_BYTES])

    store = backend(root)
    store.load()
    t = time.perf_counter()
    for seq, offset, want in reads:
        if store.span(seq, offset, len(want)) != want:
            raise ValueError("span differs from synthetic source")
    random_ms = elapsed_ms(t)
    t = time.perf_counter()
    verify_replay(store, n + 256, expected)
    replay_ms = elapsed_ms(t)
    store.close()

    # Terminate a process immediately before publishing a cross-segment append.
    # The SQL equivalent terminates inside the write transaction.
    pid = os.fork()
    if pid == 0:
        child = backend(root)
        child.load()
        child.append([b"poison"] * (PER_SEGMENT + 1), crash=True)
        os._exit(9)
    _, status = os.waitpid(pid, 0)
    if os.waitstatus_to_exitcode(status) != 51:
        raise RuntimeError("crash injection failed")
    store = backend(root)
    store.load()
    if store.n != n + 256:
        raise ValueError("unpublished records became visible")
    survivors = [frame(seq) for seq in range(n + 256, n + 259)]
    store.append(survivors)
    for rec in survivors:
        expected.update(rec[:HASH_BYTES])
    store.close()
    store = backend(root)
    store.load()
    verify_replay(store, n + 259, expected)
    store.close()
    nodes = [file.stat() for file in root.iterdir() if file.is_file()]
    return {
        "layout": backend.__name__, "preseed": n, "raw_bytes_preseed": raw_bytes,
        "median_open_append_close_32_ms": round(statistics.median(append_times), 3),
        "open_append_close_32_ms": append_times, "random_2000_spans_ms": random_ms,
        "verified_replay_ms": replay_ms, "replayed_frames": n + 256,
        "post_recovery_files": len(nodes), "post_recovery_allocated_bytes": sum(s.st_blocks * 512 for s in nodes),
        "process_crash_recovery": "pass", "sqlite": sqlite3.sqlite_version,
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--records", type=int, default=12000)
    parser.add_argument("--trials", type=int, default=3)
    parser.add_argument("--layouts", nargs="+", choices=("loose", "sqlite", "segment"), default=["loose", "sqlite", "segment"])
    args = parser.parse_args()
    if args.records <= 0 or args.trials <= 0:
        parser.error("records and trials must be positive")
    classes = {"loose": Loose, "sqlite": Sql, "segment": Segment}
    reads = probes(args.records)
    for trial in range(args.trials):
        offset = trial % len(args.layouts)
        for layout in args.layouts[offset:] + args.layouts[:offset]:
            with tempfile.TemporaryDirectory() as tmp:
                report = cycle(classes[layout], Path(tmp), args.records, reads)
                print(json.dumps(dict(trial=trial + 1, **report)), flush=True)


if __name__ == "__main__":
    main()
