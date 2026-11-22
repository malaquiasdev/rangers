import tokenize
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


class SourceTest(unittest.TestCase):
    def test_no_python_source_carries_a_comment(self):
        found = []
        for path in sorted(ROOT.rglob("*.py")):
            with path.open("rb") as handle:
                for token in tokenize.tokenize(handle.readline):
                    if token.type == tokenize.COMMENT:
                        found.append(f"{path.name}:{token.start[0]}")
        self.assertEqual(found, [])


if __name__ == "__main__":
    unittest.main()
