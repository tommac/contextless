"""Offline tests for the creativity-study toolkit. No model calls, no network."""
import json
import tempfile
import unittest
from pathlib import Path

import seeds
import studytool

PROTO = {"problems": {"A": "Problem A text.", "B": "Problem B text."},
         "conditions": ["question", "labeled"], "repeats": 1}


def make_workdir():
    work = Path(tempfile.mkdtemp())
    (work / "out0.json").write_text(json.dumps({"A": "alpha answer", "B": "bravo answer " * 60}))
    (work / "out1.json").write_text(json.dumps({"A": "charlie answer", "B": "delta answer"}))
    return work


class SeedsTest(unittest.TestCase):
    def test_twelve_seeds_three_defects_four_problems(self):
        made = seeds.make_seeds()
        self.assertEqual(len(made), 12)
        self.assertEqual({(s["problem"], s["defect"]) for s in made},
                         {(p, d) for p in "ABCD"
                          for d in ("generic", "breaks-constraint", "weak-test")})

    def test_seeds_fit_word_cap(self):
        for s in seeds.make_seeds():
            self.assertLessEqual(len(s["answer"].split()), 110, s)


class BlindTest(unittest.TestCase):
    def test_blinding_integrity_and_cap(self):
        work = make_workdir()
        studytool.blind(PROTO, work)
        rows = json.loads((work / "answers.json").read_text())
        self.assertEqual(len(rows), 4)
        ids = [r["id"] for r in rows]
        self.assertEqual(len(set(ids)), 4)
        overlong = next(r for r in rows if r["problem"] == "B" and r["original_words"] > 110)
        self.assertEqual(len(overlong["answer"].split()), 110)
        for j in range(3):
            packet = json.loads((work / f"review{j}.json").read_text())
            self.assertEqual(len(packet), 2)
            for problem in packet:
                self.assertEqual(set(problem), {"problem", "designs"})
                for design in problem["designs"]:
                    self.assertEqual(set(design), {"id", "answer"})

    def test_seeds_get_separate_ids_and_stay_out_of_answers(self):
        work = make_workdir()
        studytool.blind(PROTO, work, seeds=[s for s in seeds.make_seeds() if s["problem"] in "AB"])
        rows = json.loads((work / "answers.json").read_text())
        self.assertEqual(len(rows), 4)
        seed_meta = json.loads((work / "seeds.json").read_text())
        self.assertEqual(len(seed_meta), 6)
        self.assertTrue(all(int(i) > studytool.SEED_BASE for i in seed_meta))
        packet = json.loads((work / "review0.json").read_text())
        total = sum(len(p["designs"]) for p in packet)
        self.assertEqual(total, 10)  # 4 real + 6 seeds


class ScoreTest(unittest.TestCase):
    def _scored_workdir(self):
        work = make_workdir()
        studytool.blind(PROTO, work, seeds=[s for s in seeds.make_seeds() if s["problem"] in "AB"])
        rows = json.loads((work / "answers.json").read_text())
        seed_meta = json.loads((work / "seeds.json").read_text())
        ratings = {}
        for r in rows:
            ratings[r["id"]] = [4, 4, 4] if r["condition"] == "question" else [2, 2, 2]
        for sid, meta in seed_meta.items():
            ratings[sid] = [1, 1, 1]  # every defect caught
        for j in range(3):
            (work / f"ratings{j}.json").write_text(json.dumps(ratings))
        return work

    def test_means_and_calibration(self):
        out = studytool.summarize(PROTO, self._scored_workdir())
        self.assertEqual(out["means"]["question"]["originality"], 4)
        self.assertEqual(out["means"]["labeled"]["originality"], 2)
        cal = out["calibration"]
        self.assertEqual(cal["mean_catch_rate"], 1.0)
        self.assertEqual(out["word_cap_applied"], 1)

    def test_calibration_flags_lenient_judge(self):
        work = self._scored_workdir()
        ratings = json.loads((work / "ratings0.json").read_text())
        seed_meta = json.loads((work / "seeds.json").read_text())
        for sid in seed_meta:
            ratings[sid] = [5, 5, 5]  # judge zero misses everything
        (work / "ratings0.json").write_text(json.dumps(ratings))
        cal = studytool.calibrate(PROTO, work)
        self.assertEqual(cal["per_judge"][0]["catch_rate"], 0.0)
        self.assertEqual(cal["per_judge"][1]["catch_rate"], 1.0)
        self.assertAlmostEqual(cal["mean_catch_rate"], 2 / 3, places=3)


if __name__ == "__main__":
    unittest.main()
