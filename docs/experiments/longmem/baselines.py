"""Native retrieval adapters for the public, post-freeze answer comparison."""

import asyncio
from datetime import datetime, timezone
from importlib.metadata import distribution
import json
import os
from pathlib import Path
import tempfile
import uuid

PINS = {
    "mem0": ("mem0ai", "fff8a6472512d1bc1a4bccab8c645b65a1df7fe5"),
    "graphiti": ("graphiti-core", "9b8a4630092084734b080f2e6f409c2c23bdcc3a"),
}


def public_sessions(q):
    for date, turns in zip(q["haystack_dates"], q["haystack_sessions"]):
        yield date, [{"role": t["role"], "content": t["content"]} for t in turns]


def mem0(q, group):
    from mem0 import Memory

    with tempfile.TemporaryDirectory() as tmp:
        memory = Memory.from_config(
            {
                "vector_store": {
                    "provider": "qdrant",
                    "config": {"path": str(Path(tmp) / "vectors"), "on_disk": True},
                },
                "history_db_path": str(Path(tmp) / "history.db"),
            }
        )
        try:
            for date, messages in public_sessions(q):
                dated = [
                    dict(role=t["role"], content=f"[{date}] {t['content']}")
                    for t in messages
                ]
                memory.add(dated, user_id=group, infer=True)
            result = memory.search(
                q["question"], filters={"user_id": group}, top_k=24, threshold=0.1
            )
            return [r["memory"] + "\n" for r in result["results"]]
        finally:
            memory.close()


async def graphiti(q, group):
    from graphiti_core import Graphiti
    from graphiti_core.nodes import EpisodeType

    # This endpoint is dedicated to benchmark groups. Never delete existing data or indices.
    graph = Graphiti(
        os.environ["LME_NEO4J_URI"],
        os.environ["LME_NEO4J_USER"],
        os.environ["LME_NEO4J_PASSWORD"],
    )
    try:
        await graph.build_indices_and_constraints()
        for i, (date, messages) in enumerate(public_sessions(q)):
            when = datetime.strptime(date, "%Y/%m/%d (%a) %H:%M").replace(
                tzinfo=timezone.utc
            )
            body = "\n".join(f"{t['role']}: {t['content']}" for t in messages)
            await graph.add_episode(
                name=f"session-{i}",
                episode_body=body,
                source_description="public conversation",
                source=EpisodeType.message,
                reference_time=when,
                group_id=group,
            )
        hits = await graph.search(q["question"], group_ids=[group], num_results=24)
        return [f"[{h.valid_at} to {h.invalid_at}] {h.fact}\n" for h in hits]
    finally:
        await graph.close()


def retrieve(arm, q):
    package, revision = PINS[arm]
    info = json.loads(distribution(package).read_text("direct_url.json") or "{}")
    if info.get("vcs_info", {}).get("commit_id") != revision:
        raise ValueError(f"install {package} at the benchmark revision {revision}")
    group = "lme" + uuid.uuid4().hex
    items = mem0(q, group) if arm == "mem0" else asyncio.run(graphiti(q, group))
    return items, {
        "commit": revision,
        "group": group,
        "ingest_sessions": len(q["haystack_sessions"]),
        "returned_items": len(items),
        "native_model_usage": None,
    }
