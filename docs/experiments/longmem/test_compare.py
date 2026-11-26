import json
from pathlib import Path
import tempfile
import unittest
from types import SimpleNamespace
from unittest.mock import patch

import compare


class PublicComparisonTest(unittest.TestCase):
    def setUp(self):
        self.q = {
            "question_id": "case",
            "question": "Where?",
            "question_date": "today",
            "haystack_dates": ["yesterday"],
            "haystack_sessions": [
                [
                    {"role": "user", "content": "Maple Library. " + "A" * 4900},
                    {"role": "user", "content": "Harbor Room. " + "B" * 4900},
                ]
            ],
            "answer": "HIDDEN_GOLD",
        }
        hits = [
            {
                "ID": f"t{i + 1:06}#a1",
                "Session": 0,
                "Message": i,
                "Start": 0,
                "End": len(t["content"].encode()),
            }
            for i, t in enumerate(self.q["haystack_sessions"][0])
        ]
        self.ranks = {"QID": "case", "BM25": hits, "Contextless": hits}

    def answer(self, arm, ranks=None):
        seen = []

        def reader(model, instruction, content, calls):
            self.assertNotIn("HIDDEN_GOLD", content)
            seen.append(content)
            if "JSON array" in instruction:
                return '["t000002#a1"]'
            return "Maple Library."

        return compare.answer(
            self.q, ranks or self.ranks, arm, "fixture", call=reader
        ), seen

    def test_overflow(self):
        records = {}
        for arm in ("verbatim", "summary", "overview"):
            record, calls = self.answer(arm)
            records[arm] = record
            self.assertTrue(record["overflow"])
            self.assertLessEqual(record["reader_bytes"], compare.BUDGET)
            self.assertEqual(len(calls), 1 if arm == "verbatim" else 2)
            if arm == "overview":
                self.assertIn("B" * 100, calls[-1])
                self.assertNotIn("A" * 4900, calls[-1])
        self.assertEqual(len({r["candidate_sha256"] for r in records.values()}), 1)

    def test_fit_does_not_pay_for_a_summary(self):
        short = {
            "Contextless": self.ranks["Contextless"][:1],
            "BM25": self.ranks["BM25"][:1],
        }
        first = self.answer("verbatim", short)[1]
        self.assertEqual(first, self.answer("summary", short)[1])
        self.assertEqual(first, self.answer("overview", short)[1])

    def test_complete_run_resumes_and_exports_blind_hypotheses(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d)
            rankfile = root / "ranks"
            rankfile.write_text(json.dumps(self.ranks))
            args = SimpleNamespace(
                arm="full", model="fixture", id=None, out=root, ranks=rankfile
            )
            qs, rs = {"case": self.q}, {"case": self.ranks}
            with patch.object(
                compare, "answer", return_value={"hypothesis": "Maple Library."}
            ) as model:
                compare.batch(qs, rs, args)
                compare.batch(qs, rs, args)
                self.assertEqual(model.call_count, 1)
            exported = json.loads((root / "full/hypotheses.jsonl").read_text())
            self.assertEqual(
                exported, {"question_id": "case", "hypothesis": "Maple Library."}
            )


if __name__ == "__main__":
    unittest.main()
