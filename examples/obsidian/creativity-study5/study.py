#!/usr/bin/env python3
"""Build frozen packets, run pinned adapters, and score the executable creativity study."""
import argparse
import concurrent.futures
import fcntl
import hashlib
import json
import os
from pathlib import Path
import random
import re
import secrets
import sqlite3
import subprocess
import sys
import tempfile
import time
from contextlib import closing
from statistics import mean

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE.parent / "studytool"))
import studytool

PROTO = HERE / "protocol.json"
SELECTION = HERE / "selection.json"
FIELDS = ("recency", "frequency", "value", "cost", "age")
METRICS = ("feasible", "beats_baseline", "distinct", "success", "test_confirmed", "test_distinguishes")
GENERATION = """Independently design a policy for each task. Return an object keyed by task ID. Each value is {"policy":{...},"test":{"workload":"burst|scan|reuse","baseline":"...","prediction":"win|tie|loss"},"explanation":"..."}; choose one value in each pipe-separated list. Use only the supplied material. The policy is implemented directly by a fixed simulator. Maximize useful originality. In at most 110 explanation words, describe the mechanism, name a simpler alternative, and propose a test with a specific prediction that could disfavor your design. If source material is present, use it at your discretion. Do not add source identifiers or mention study conditions in the explanation. When a draft is supplied, reconsider and revise it. Only the final submitted design will be run."""
RUBRIC = """Score each anonymous design independently against its brief. Return one object mapping task IDs to [originality, feasibility, discriminating test], each an integer 1 to 5. Originality: 1 familiar default, 3 useful adaptation, 5 surprising distinct mechanism. Feasibility: 1 violates constraints, 3 plausible with gaps, 5 workable under the stated limits. Test: 1 fails to discriminate, 3 measurable comparison, 5 isolates the design choice against a simpler baseline and could disfavor it. No run outcomes are supplied."""
CACHE = """Resident slots have equal size. Score is fetch cost times value saved on hits. Evict the lowest scoring resident. admit (1..3) is the number of observed appearances required for admission, history (0..3) is ghost metadata capacity in multiples of resident capacity. Integer weights recency, frequency, value, cost, age (-4..4) reward recent access, repeated access, estimated benefit, expensive retrieval, or age. Omitted admit defaults to 1; all other fields default to 0. Diagnostic baselines: lru and 2q."""
SCHEDULE = """Each epoch has limited read credits. Spend a passage's cost in credits and earn its value if completed before its deadline. Highest-scoring eligible request executes next. Integer weights recency, frequency, value, cost, age (-4..4) reward recent arrival, repeated requests, estimated benefit, expensive retrieval, or older age. Omitted weights are zero. Diagnostic baselines: fifo and value-cost."""
# Outside-field notes about human practice, statistics, concurrency, or representation.
# Exclude direct resource-allocation, ranking, storage, and scheduling concepts.
DISTANT = ("Maker-checker separation", "Zettelkasten", "Evergreen note titles",
           "Decontextualization", "McNemar test for paired outcomes", "Provenance and lineage",
           "Version control blame and bisect", "Conflict-free replicated data types",
           "Optimistic concurrency control", "Snapshot isolation and write skew")


def encode(obj):
    return json.dumps(obj, sort_keys=True, ensure_ascii=False, separators=(",", ":")).encode()


def digest(obj):
    return hashlib.sha256(encode(obj)).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def save(path, obj):
    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", dir=path.parent, delete=False) as f:
        json.dump(obj, f, indent=2, ensure_ascii=False, allow_nan=False)
        f.write("\n")
    os.replace(f.name, path)


def neutral(label):
    # Equal UTF-8 bytes and word counts; pad with ordinary space after punctuation.
    n = len(label.split())
    phrases = {5:"Read this note as background.", 6:"Read this note as general background.",
               7:"Read this note as general background material.", 8:"Read this note as general background for context."}
    return phrases.get(n, " ".join((["See"] + ["the"] * (n - 2) + ["note."])))


def select(proto):
    curated = read(SELECTION)
    _, notes = studytool.vault_prose()
    rng = random.Random(19117)
    cards, timings = {}, {}
    start = time.perf_counter()
    with closing(sqlite3.connect(":memory:")) as db:
        db.execute("CREATE VIRTUAL TABLE docs USING fts5(path UNINDEXED,title,body,tokenize='porter unicode61')")
        db.executemany("INSERT INTO docs VALUES (?,?,?)", [(p, Path(p).stem, body) for p, body in sorted(notes.items())])
        indexed = time.perf_counter() - start
        for case, problem in proto["briefs"].items():
            start = time.perf_counter()
            terms = [w for w in re.findall(r"[a-z0-9]+", problem.lower()) if w not in studytool.STOP]
            query = " OR ".join(f'"{w}"' for w in terms)
            hits = [p for p, in db.execute("SELECT path FROM docs WHERE docs MATCH ? ORDER BY rank,path LIMIT 3", (query,))]
            if len(hits) != 3:
                raise ValueError(f"fewer than three search hits for {case}")
            search_secs = time.perf_counter() - start
            start = time.perf_counter()
            picks = curated["cards"][case]
            selected = {c["path"] for c in picks}
            pool = [f"fields/{name}.md" for name in DISTANT if f"fields/{name}.md" not in selected]
            distant = rng.sample(pool, 3)
            timings[case] = {"search_seconds":search_secs, "random_seconds":time.perf_counter()-start}
            def card(path):
                return {"title":Path(path).stem,"text":studytool.excerpt(notes[path], proto["source_bytes_per_card"])}
            row = {"free":[], "search":list(map(card,hits)), "random":list(map(card,distant)), "neutral":[], "labeled":[]}
            for item in picks:
                label = item["label"]
                blank = neutral(label)
                width = max(len(blank.encode()), len(label.encode()))
                for arm, role in [("neutral", blank), ("labeled", label)]:
                    row[arm].append(dict(card(item["path"]), relationship=role+" "*(width-len(role.encode()))))
            cards[case] = row
    audit = {"curation":curated, "index_seconds":indexed, "selection_by_case":timings, "size":len(notes)}
    return cards, audit


def prepare(work, reader, reader_family, judge, judge_family):
    if reader_family.strip().casefold() == judge_family.strip().casefold():
        raise ValueError("reviewer must be from a different model family")
    if work.exists() and any(work.iterdir()):
        raise ValueError("new run needs an empty directory")
    protocol = read(PROTO)
    cards, selection = select(protocol)
    holds = {key:secrets.randbits(32) for key in protocol["briefs"]}
    manifest = {"protocol":protocol,"reader":{"model":reader,"family":reader_family},
                "judge":{"model":judge,"family":judge_family},"holdout_commitment":digest(holds),
                "cards_sha256":digest(cards),"script_sha256":hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                "selection_sha256":digest(selection["curation"])}
    save(work/"private"/"holdouts.json", holds)
    save(work/"cards.json", cards)
    save(work/"selection-audit.json", selection)
    save(work/"manifest.json", manifest)


def info(proto, case):
    i = int(case[1:])-1
    return {"brief":proto["briefs"][case], "capacity":proto["capacities"][i],
            "rules":CACHE if case.startswith("C") else SCHEDULE}


def token(key):
    return hashlib.sha256(key.encode()).hexdigest()[:12]


def response_file(work, job):
    return work/"responses"/(job+".json")


def make_packet(identity, instruction, tasks, protocol):
    packet = {**identity, "instructions":instruction, "tasks":tasks, "max_output_tokens":protocol["output_tokens_per_request"]}
    if len(encode(packet)) > protocol["input_bytes_per_request"]:
        raise ValueError("request exceeds frozen byte ceiling")
    return packet


def generation_jobs(work, stage, manifest, cards):
    proto = manifest["protocol"]
    drafts = collect(work, "draft", manifest, cards) if stage == "revise" else {}
    jobs = {}
    for rep in range(proto["repeats"]):
        for arm in (proto["arms"] if stage=="revise" else ["draft"]):
            order = list(proto["briefs"])
            random.Random(f"study5:{rep}:{arm}").shuffle(order)
            size = proto["tasks_per_generation"]
            for offset in range(0,len(order),size):
                group = order[offset:offset+size]
                job = f"{stage}-{rep}-{arm}-{offset//size}"
                tasks, key = {}, {}
                for case in group:
                    id_ = token(job+case)
                    task = info(proto,case)
                    if stage=="revise":
                        # An invalid draft remains the shared starting point for every revision.
                        task["draft"] = drafts[(rep,"draft",case)]
                        if cards[case][arm]:
                            task["material"] = cards[case][arm]
                    tasks[id_] = task
                    key[id_] = (rep,arm,case)
                jobs[job] = (make_packet(manifest["reader"], GENERATION, tasks, proto), key)
    return jobs


def collect(work, stage, manifest, cards):
    combined = {}
    for job, (packet, mapping) in generation_jobs(work, stage, manifest, cards).items():
        result = validated_response(work, job, packet)
        for ident, fields in mapping.items():
            combined[fields] = result["answers"][ident]
    return combined


def checked(run, packet):
    if not isinstance(run,dict) or any(run.get(k)!=packet[k] for k in ["model","family"]):
        raise ValueError("adapter did not confirm the pinned model and family")
    if not isinstance(run.get("answers"),dict) or set(run["answers"])!=set(packet["tasks"]):
        raise ValueError("adapter must return exactly the requested task IDs")
    usage = run.get("usage")
    if not isinstance(usage, dict):
        raise ValueError("provide usage object; use null for quantities provider omits")
    for field in ("input_tokens","output_tokens"):
        if field not in usage or (usage[field] is not None and (type(usage[field]) is not int or usage[field]<0)):
            raise ValueError(f"bad usage {field}")
    if usage["output_tokens"] is not None and usage["output_tokens"]>packet["max_output_tokens"]:
        raise ValueError("adapter exceeded the frozen output token budget")
    return run


def validated_response(work, job, packet):
    path = response_file(work,job)
    if not path.exists():
        raise ValueError(f"missing completed job {job}")
    data = read(path)
    if data["request_sha256"] != digest(packet):
        raise ValueError(f"request drift at {job}")
    return checked(data["response"],packet)


def manifest_cards(work):
    manifest, cards = read(work/"manifest.json"), read(work/"cards.json")
    if digest(cards)!=manifest["cards_sha256"]:
        raise ValueError("frozen cards changed")
    return manifest,cards

def review_jobs(work, manifest, cards):
    proto = manifest["protocol"]
    proposals = collect(work, "revise", manifest, cards)
    jobs = {}
    for j in range(proto["judges"]):
        for case in proto["briefs"]:
            job = f"judge-{j}-{case}"
            entries = [(key,x) for key,x in proposals.items() if key[2]==case]
            entries += [(("seed","test",case), {"policy":{"recency":1},
                       "test":{"workload":"reuse","baseline":"lru" if case[0]=="C" else "fifo","prediction":"win"},
                       "explanation":"The design is attractive. Ask an admirer whether they like the description. A positive answer proves the mechanism."})]
            rng = random.Random(job)
            rng.shuffle(entries)
            ids = rng.sample(range(100000,1000000),len(entries))
            key = {str(i):k for i,(k,_) in zip(ids,entries)}
            tasks = {str(i):{"brief":info(proto,case),"design":design} for i,(_,design) in zip(ids,entries)}
            jobs[job] = (make_packet(manifest["judge"],RUBRIC,tasks,proto),key)
    return jobs


def jobs_for(work,stage):
    manifest,cards = manifest_cards(work)
    if stage=="judge":
        return review_jobs(work,manifest,cards)
    return generation_jobs(work,stage,manifest,cards)


def export_packets(work,stage):
    jobs = jobs_for(work,stage)
    for job,(packet,key) in jobs.items():
        save(work/"requests"/(job+".json"),packet)
        save(work/"private"/"keys"/(job+".json"),key)
    return jobs


def check_judgments(answer):
    return isinstance(answer,list) and len(answer)==3 and all(type(x) is int and 1<=x<=5 for x in answer)


def store_response(work,job,packet,response,seconds):
    checked(response,packet)
    if job.startswith("judge-") and not all(map(check_judgments,response["answers"].values())):
        raise ValueError("judge must supply three integer ratings 1..5 per design")
    save(response_file(work,job), {"request_sha256":digest(packet), "input_bytes":len(encode(packet)),
                                   "elapsed_seconds":seconds,"response":response})


def run_adapter(work, stage, adapter, workers, timeout):
    jobs = export_packets(work,stage)
    role = "judge" if stage == "judge" else "reader"
    adapter_hash = hashlib.sha256(Path(adapter).read_bytes()).hexdigest()
    # Pin both generation stages to the same executable even when resuming.
    pin = work/"private"/(role+"-adapter.sha256")
    with (work/"private"/(role+"-adapter.lock")).open("a") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        if pin.exists() and pin.read_text().strip() != adapter_hash:
            raise ValueError(f"{role} adapter changed since the first call")
        if not pin.exists():
            pin.write_text(adapter_hash+"\n")
    def one(item):
        job,(packet,_) = item
        lock = work/"locks"/(job+".lock")
        lock.parent.mkdir(parents=True,exist_ok=True)
        with lock.open("a") as f:
            fcntl.flock(f,fcntl.LOCK_EX)
            if response_file(work,job).exists():
                validated_response(work,job,packet)
                return
            attempt = work/"attempts"/job/str(time.time_ns())
            attempt.mkdir(parents=True)
            start=time.perf_counter()
            try:
                # A fresh directory has no hidden traces, condition keys or run history.
                with tempfile.TemporaryDirectory() as scratch:
                    proc=subprocess.run([adapter],input=encode(packet),capture_output=True,
                                        cwd=scratch,timeout=timeout)
                seconds=time.perf_counter()-start
                (attempt/"stdout").write_bytes(proc.stdout)
                (attempt/"stderr").write_bytes(proc.stderr)
                save(attempt/"meta.json",{"request_sha256":digest(packet),"returncode":proc.returncode,
                     "elapsed_seconds":seconds,"input_bytes":len(encode(packet))})
                if proc.returncode:
                    raise ValueError(f"{job}: adapter exited {proc.returncode}; inspect {attempt}")
                store_response(work,job,packet,json.loads(proc.stdout),seconds)
                print(f"completed {job}",flush=True)
            except subprocess.TimeoutExpired as e:
                (attempt/"stdout").write_bytes(e.stdout or b"")
                (attempt/"stderr").write_bytes(e.stderr or b"")
                save(attempt/"meta.json",{"request_sha256":digest(packet),"status":"timeout",
                     "elapsed_seconds":time.perf_counter()-start})
                raise ValueError(f"{job}: adapter timed out")
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        list(pool.map(one,jobs.items()))


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    commands=parser.add_subparsers(dest="command",required=True)
    p=commands.add_parser("prepare")
    p.add_argument("work",type=Path)
    for field in ("reader","reader-family","judge","judge-family"):
        p.add_argument("--"+field,required=True)
    for cmd in ("packets","run","ingest"):
        p=commands.add_parser(cmd)
        p.add_argument("work",type=Path)
        p.add_argument("--stage",choices=["draft","revise","judge"],required=True)
        if cmd=="run":
            p.add_argument("--adapter",type=Path,required=True)
            p.add_argument("--jobs",type=int,default=1)
            p.add_argument("--timeout",type=int,default=600)
        if cmd=="ingest":
            p.add_argument("--job",required=True)
            p.add_argument("--response",type=Path,required=True)
            p.add_argument("--elapsed",type=float)
    p=commands.add_parser("score")
    p.add_argument("work",type=Path)
    p.add_argument("--cx",type=Path,required=True)
    args=parser.parse_args()
    work=args.work.resolve()
    if args.command=="prepare":
        prepare(work,args.reader,args.reader_family,args.judge,args.judge_family)
    elif args.command=="packets":
        jobs=export_packets(work,args.stage)
        print(f"{len(jobs)} requests written under {work/'requests'}")
    elif args.command=="run":
        run_adapter(work,args.stage,str(args.adapter.resolve()),args.jobs,args.timeout)
    elif args.command=="ingest":
        jobs=export_packets(work,args.stage)
        packet,_=jobs[args.job]
        if response_file(work,args.job).exists():
            raise ValueError("job already completed")
        store_response(work,args.job,packet,read(args.response),args.elapsed)
    else:
        score(work,str(args.cx.resolve()))

def design_error(design, case, max_words):
    if not isinstance(design,dict) or set(design)!={"explanation","policy","test"}:
        return "invalid design fields"
    text=design["explanation"]
    if not isinstance(text,str) or not 1<=len(text.split())<=max_words:
        return "explanation violates word bound"
    policy=design["policy"]
    if not isinstance(policy,dict) or set(policy)-set(FIELDS)-{"admit","history"}:
        return "invalid policy fields"
    for key,n in policy.items():
        lo,hi={"admit":(1,3),"history":(0,3)}.get(key,(-4,4))
        if type(n) is not int or not lo<=n<=hi:
            return f"invalid value for {key}"
    test=design["test"]
    baselines=("lru","2q") if case[0]=="C" else ("fifo","value-cost")
    if not isinstance(test,dict) or set(test)!={"workload","baseline","prediction"}:
        return "invalid test fields"
    if test["baseline"] not in baselines or test["workload"] not in ("burst","scan","reuse") or test["prediction"] not in ("win","tie","loss"):
        return "invalid test choice"
    return None


def paired_interval(rows, briefs):
    means={}
    for case in briefs:
        group=[r for r in rows if r["case"]==case]
        arm=lambda name: mean(r["evaluation"]["success"] for r in group if r["arm"]==name)
        means[case]=arm("labeled")-arm("free")
    diffs=list(means.values())
    rng=random.Random(33091)
    samples=sorted(mean(rng.choices(diffs,k=len(diffs))) for _ in range(10000))
    return {"difference":mean(diffs),"ci95":[samples[249],samples[9749]],"brief_differences":means,
            "method":"10,000 percentile bootstrap resamples of whole paired briefs"}


def usage_report(work):
    usage={}
    per_arm={}
    for path in sorted((work/"responses").glob("*.json")):
        stage=path.stem.split("-")[0];data=read(path)
        buckets=usage.setdefault(stage,{"calls":0,"input_bytes":0,"known_input_tokens":0,"known_output_tokens":0,
                                        "missing_input_tokens":0,"missing_output_tokens":0,"known_seconds":0.0,
                                        "missing_seconds":0})
        buckets["calls"]+=1
        buckets["input_bytes"]+=data["input_bytes"]
        if stage == "revise":
            arm = path.stem.split("-")[2]
            batch = per_arm.setdefault(arm, {"calls":0,"input_bytes":0,
                "input_tokens":[],"output_tokens":[],"seconds":[]})
            batch["calls"] += 1
            batch["input_bytes"] += data["input_bytes"]
            for meter in ("input_tokens","output_tokens"):
                batch[meter].append(data["response"]["usage"][meter])
            batch["seconds"].append(data["elapsed_seconds"])
        elapsed=data["elapsed_seconds"]
        if elapsed is None:
            buckets["missing_seconds"]+=1
        else:
            buckets["known_seconds"]+=elapsed
        for name in ("input_tokens","output_tokens"):
            n=data["response"]["usage"][name]
            if n is None:
                buckets["missing_"+name]+=1
            else:
                buckets["known_"+name]+=n
    return {"completed":usage,"revision_usage_by_arm":per_arm,"adapter_attempts":len(list((work/"attempts").glob("*/*/meta.json"))),
            "usage_scope":"Completed provider calls; failed attempts may have incurred unreported token use. No reopens allowed."}


def score(work,cx):
    manifest,cards=manifest_cards(work)
    proto=manifest["protocol"]
    proposals=collect(work,"revise",manifest,cards)
    holds=read(work/"private"/"holdouts.json")
    if digest(holds)!=manifest["holdout_commitment"]:
        raise ValueError("holdout commitment mismatch")
    rows=[]
    for (rep,arm,case),answer in sorted(proposals.items()):
        error=design_error(answer,case,proto["explanation_words"])
        if error:
            result={k:False for k in METRICS}
            result["invalid_reason"]=error
        else:
            req={"family":"cache" if case.startswith("C") else "schedule", "scenario":int(case[1:])-1,
                 "seed":holds[case], "policy":answer["policy"], "test":answer["test"]}
            proc=subprocess.run([cx,"demo","creativity","-evaluate"],input=encode(req),capture_output=True,timeout=60)
            if proc.returncode:
                raise RuntimeError(f"simulator failed for {case}/{arm}/{rep}: {proc.stderr.decode(errors='replace')[:500]}")
            result=json.loads(proc.stdout)
            if not all(type(result.get(k)) is bool for k in METRICS):
                raise ValueError(f"simulator missing metric(s): {list(result)}")
        rows.append({"case":case,"repeat":rep,"arm":arm,"evaluation":result})
    judges,lookup=[],{(r["repeat"],r["arm"],r["case"]):r for r in rows}
    judge_jobs=review_jobs(work,manifest,cards)
    completed=[]
    for name,(packet,key) in judge_jobs.items():
        if not response_file(work,name).exists():
            continue
        ratings=validated_response(work,name,packet)["answers"]
        completed.append(name)
        for ident,(rep,arm,case) in key.items():
            values=ratings[ident]
            if not check_judgments(values):
                raise ValueError(f"invalid ratings {name}/{ident}")
            if rep=="seed":
                judges.append({"job":name,"case":case,"caught_weak_test":values[2]<=2})
            else:
                lookup[rep,arm,case].setdefault("ratings",{})[name]=values
    summary={}
    for arm in proto["arms"]:
        group=[r for r in rows if r["arm"]==arm]
        metrics={k:sum(r["evaluation"][k] for r in group) for k in METRICS}
        metrics.update(n=len(group),invalid=sum("invalid_reason" in r["evaluation"] for r in group))
        if len(completed)==len(judge_jobs):
            metrics["judge_means"]={name:mean(grade[i] for row in group for grade in row["ratings"].values())
                                    for i,name in enumerate(("originality","feasibility","test"))}
        summary[arm]=metrics
    primary=paired_interval(rows,proto["briefs"])
    save(work/"results.json",{"arms":summary,"primary":primary,
         "holdout_seeds":holds, "holdout_commitment":manifest["holdout_commitment"],
         "review":{"jobs_complete":len(completed),"jobs_planned":len(judge_jobs),"calibration":judges},
         "evaluator_sha256":hashlib.sha256(Path(cx).read_bytes()).hexdigest(),
         "usage":usage_report(work),"rows":rows})
    print(json.dumps({"arms":summary,"primary":primary},indent=2))


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, KeyError) as exc:
        sys.exit(str(exc))
