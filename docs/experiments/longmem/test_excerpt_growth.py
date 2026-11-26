import unittest

from compare import BUDGET
from excerpt_growth import Source, boundaries, grow, reader_input, render


class ExactExcerptTest(unittest.TestCase):
    def test_multibyte_anchor_and_rendered_budget(self):
        # Two copies force partial spans. Non-ASCII characters make byte
        # offsets differ from character offsets, including at the anchor.
        text = ("Répétition à côté. " * 270 + "Peacock beacon. " + "Épilogue. " * 300).encode()
        pool = [Source((0, i), f"t{i + 1:06}#a1", "yesterday", "user", 55, text) for i in range(2)]
        chosen = grow(pool, "Where is the peacock beacon?", True)
        self.assertEqual(len(chosen), 2)
        self.assertLessEqual(len(render(chosen)), BUDGET)
        for src, lo, hi in chosen:
            self.assertIn("Peacock beacon.", src.text[lo:hi].decode())
            self.assertIn(f"@{55 + lo}:{55 + hi}".encode(), src.header(lo, hi))
        ends = boundaries(text)
        self.assertEqual(b"".join(text[a:b] for a, b in zip(ends, ends[1:])), text)

    def test_reader_input_discards_benchmark_labels(self):
        row = {"answer": "private-label", "question": "Where?", "haystack_dates": ["today"],
               "haystack_sessions": [[{"role": "user", "content": "Hello", "has_answer": True}]]}
        clean = reader_input(row)
        self.assertNotIn("answer", clean)
        self.assertEqual(clean["haystack_sessions"], [[{"role": "user", "content": "Hello"}]])


if __name__ == "__main__":
    unittest.main()
