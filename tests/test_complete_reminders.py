import json
import os
import stat
import subprocess
import tempfile
import time
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "bin" / "complete-reminders"
STOP_HOOK = ROOT / "bin" / "stop-push-check"

RECORDING_STUB = """#!/bin/bash
for argument in "$@"; do printf '%s\\n' "$argument"; done > "$STUB_DIR/arguments"
cat > "$STUB_DIR/script"
"""


class CompleteRemindersTest(unittest.TestCase):
    def setUp(self):
        self.folder = tempfile.TemporaryDirectory()
        self.addCleanup(self.folder.cleanup)
        self.stub_dir = Path(self.folder.name) / "stub"
        self.stub_dir.mkdir()
        self.bin_dir = Path(self.folder.name) / "bin"
        self.bin_dir.mkdir()
        (self.bin_dir / "bash").symlink_to(subprocess.run(
            ["which", "bash"], capture_output=True, text=True, check=True
        ).stdout.strip())

    def install_stub(self, body):
        path = self.bin_dir / "osascript"
        path.write_text(body)
        path.chmod(path.stat().st_mode | stat.S_IEXEC)

    def run_hook(self, env_extra=None, payload='{"prompt": "go"}'):
        env = {"PATH": str(self.bin_dir), "STUB_DIR": str(self.stub_dir)}
        env.update(env_extra or {})
        return subprocess.run(
            [str(SCRIPT)], input=payload, capture_output=True, text=True, check=False, env=env
        )

    def recorded_arguments(self):
        return (self.stub_dir / "arguments").read_text().splitlines()

    def recorded_script(self):
        return (self.stub_dir / "script").read_text()

    def assertSilent(self, result):
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")

    def test_given_the_human_sends_a_message_when_the_hook_runs_then_it_asks_reminders_about_the_tarefas_claude_list_and_the_prefix(self):
        self.install_stub(RECORDING_STUB)
        self.assertSilent(self.run_hook())
        self.assertEqual(self.recorded_arguments()[-2:], ["Tarefas Claude", "Claude espera: "])

    def test_given_the_human_sends_a_message_when_the_hook_runs_then_the_script_completes_only_open_reminders_that_start_with_the_prefix(self):
        self.install_stub(RECORDING_STUB)
        self.run_hook()
        script = self.recorded_script()
        self.assertIn('application "Reminders"', script)
        self.assertIn("completed is false", script)
        self.assertIn("name starts with", script)
        self.assertIn("set completed of", script)

    def test_given_the_list_does_not_exist_when_the_hook_runs_then_the_script_does_nothing(self):
        self.install_stub(RECORDING_STUB)
        self.run_hook()
        self.assertIn("exists list", self.recorded_script())

    def test_given_osascript_prints_something_when_the_hook_runs_then_standard_output_stays_empty(self):
        self.install_stub('#!/bin/bash\necho "3 reminders"\necho "warning" >&2\n')
        result = self.run_hook()
        self.assertSilent(result)
        self.assertEqual(result.stderr, "")

    def test_given_osascript_fails_when_the_hook_runs_then_exits_silently_with_zero(self):
        self.install_stub("#!/bin/bash\nexit 1\n")
        self.assertSilent(self.run_hook())

    def test_given_osascript_is_not_installed_when_the_hook_runs_then_exits_silently_with_zero(self):
        self.assertSilent(self.run_hook())

    def test_given_osascript_hangs_when_the_hook_runs_then_it_gives_up_after_the_timeout(self):
        self.install_stub("#!/bin/bash\nexec sleep 30\n")
        started = time.monotonic()
        result = self.run_hook({"COMPLETE_REMINDERS_TIMEOUT": "1"})
        self.assertSilent(result)
        self.assertLess(time.monotonic() - started, 10)

    def test_given_input_that_is_not_json_when_the_hook_runs_then_it_still_asks_reminders(self):
        self.install_stub(RECORDING_STUB)
        self.assertSilent(self.run_hook(payload="not json"))
        self.assertEqual(self.recorded_arguments()[-1], "Claude espera: ")

    def test_given_the_prefix_the_hook_closes_when_the_stop_hook_blocks_then_the_reminder_it_asks_for_uses_that_same_prefix(self):
        self.install_stub(RECORDING_STUB)
        self.run_hook()
        prefix = self.recorded_arguments()[-1]
        transcript = Path(self.folder.name) / "session.jsonl"
        entries = [
            {"type": "user", "message": {"role": "user", "content": "go"}},
            {"type": "assistant", "message": {"role": "assistant", "content": [{"type": "text", "text": "Which one?"}]}},
        ]
        transcript.write_text("".join(json.dumps(entry) + "\n" for entry in entries))
        result = subprocess.run(
            [str(STOP_HOOK)],
            input=json.dumps({"transcript_path": str(transcript)}),
            capture_output=True,
            text=True,
            check=False,
            env={**os.environ},
        )
        self.assertIn(prefix, json.loads(result.stdout)["reason"])


if __name__ == "__main__":
    unittest.main()
