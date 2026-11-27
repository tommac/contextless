"""Regression tests for the actual two-stage protocol and evidence boundaries."""
from contextlib import redirect_stdout
import io
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch
import study


class ProtocolTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.work = Path(self.tmp.name)
        study.prepare(self.work, "r", "a", "j", "b")
        self.m, self.cards = study.manifest_cards(self.work)

    @staticmethod
    def design(case):
        return {"policy": {"value": 2, "frequency": 1},
                "test": {"workload": "scan", "baseline": "lru" if case.startswith("C") else "fifo",
                         "prediction": "win"}, "explanation": "Protect a valuable passage when visits repeat."}

    def fill(self, stage):
        for job, (packet, mapping) in study.jobs_for(self.work, stage).items():
            if stage == "judge":
                answers = {k: [2, 3, 1] for k in mapping}
            else:
                answers = {k: self.design(v[2]) for k, v in mapping.items()}
            response = {"model": packet["model"], "family": packet["family"], "answers": answers,
                        "usage": {"input_tokens": None, "output_tokens": 10}}
            study.store_response(self.work, job, packet, response, .1)

    def test_identical_drafts_labels_caps_and_blindness(self):
        self.fill("draft")
        seen = {}
        for job, (packet, mapping) in study.jobs_for(self.work, "revise").items():
            self.assertNotIn("holdout", str(packet))
            self.assertLessEqual(len(study.encode(packet)), self.m["protocol"]["input_bytes_per_request"])
            for key,(rep,arm,case) in mapping.items():
                task = packet["tasks"][key]
                seen.setdefault((case, rep), []).append(task["draft"])
                if arm == "free":
                    self.assertNotIn("material",task)
        self.assertEqual(len(seen),32)
        self.assertTrue(all(len(v)==5 and all(d==v[0] for d in v) for v in seen.values()))
        for case, cards in self.cards.items():
            self.assertEqual({x["title"] for x in cards["labeled"]} & {x["title"] for x in cards["random"]},set())
            for n,l in zip(cards["neutral"],cards["labeled"]):
                self.assertEqual((n["title"],n["text"]),(l["title"],l["text"]))
                self.assertEqual(len(n["relationship"].encode()),len(l["relationship"].encode()))
                self.assertEqual(len(n["relationship"].split()),len(l["relationship"].split()))
        self.fill("revise")
        jobs = study.jobs_for(self.work,"judge")
        self.assertEqual(len(jobs),48)
        for packet, key in jobs.values():
            self.assertEqual(len(key),11)
            self.assertEqual(sum(rep=="seed" for rep,_,_ in key.values()),1)
            self.assertEqual(set(packet["tasks"]),set(key))
            self.assertTrue(all(set(x)=={"brief","design"} for x in packet["tasks"].values()))

    def test_score_keeps_invalid_and_rejects_holdout_drift(self):
        self.fill("draft")
        self.fill("revise")
        job,(packet,key)=next(iter(study.jobs_for(self.work,"revise").items()))
        path=study.response_file(self.work,job)
        resp=study.read(path)
        resp["response"]["answers"][next(iter(key))]["explanation"]="long " * 111
        study.save(path, resp)
        self.fill("judge")
        fake = self.work/"test-evaluator"
        fake.write_text("#!/usr/bin/env python3\n")
        result={"returncode":0,"stdout":study.encode({metric:True for metric in study.METRICS})}
        with patch.object(study.subprocess,"run",return_value=type("Exit",(),result)()),redirect_stdout(io.StringIO()):
            study.score(self.work,str(fake))
        report=study.read(self.work/"results.json")
        self.assertEqual(len(report["rows"]),160)
        self.assertEqual(sum(x["invalid"] for x in report["arms"].values()),1)
        self.assertEqual(sum(x["success"] for x in report["arms"].values()),159)
        self.assertEqual(report["review"]["jobs_complete"],48)
        self.assertEqual(report["usage"]["completed"]["revise"]["missing_input_tokens"],40)
        study.save(self.work/"private"/"holdouts.json",{})
        with self.assertRaisesRegex(ValueError,"commitment"):
            study.score(self.work,str(fake))

    def test_reject_model_substitution_and_request_drift(self):
        self.fill("draft")
        job,(packet,key)=next(iter(study.jobs_for(self.work,"draft").items()))
        path=study.response_file(self.work,job)
        resp=study.read(path)
        changed=dict(packet,instructions="changed")
        with self.assertRaisesRegex(ValueError,"drift"):
            study.validated_response(self.work,job,changed)
        resp["response"]["model"]="wrong"
        study.save(path,resp)
        with self.assertRaisesRegex(ValueError,"pinned model"):
            study.validated_response(self.work,job,packet)


if __name__=="__main__":
    unittest.main()
