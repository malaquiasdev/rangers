import json
import subprocess
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).resolve().parent.parent / "bin" / "stop-push-check"


def user(text):
    return {"type": "user", "message": {"role": "user", "content": text}}


def tool_result():
    return {
        "type": "user",
        "message": {
            "role": "user",
            "content": [{"type": "tool_result", "tool_use_id": "t1", "content": "ok"}],
        },
    }


def said(text):
    return {
        "type": "assistant",
        "message": {"role": "assistant", "content": [{"type": "text", "text": text}]},
    }


def called(name):
    return {
        "type": "assistant",
        "message": {
            "role": "assistant",
            "content": [{"type": "tool_use", "id": "t1", "name": name, "input": {}}],
        },
    }


REMINDER_COMMAND = (
    "osascript -e 'tell application \"Reminders\" to make new reminder "
    "at end of reminders of list \"Tarefas Claude\" "
    "with properties {name:\"Claude espera: which one?\"}'"
)


def ran(command):
    return {
        "type": "assistant",
        "message": {
            "role": "assistant",
            "content": [
                {"type": "tool_use", "id": "t2", "name": "Bash", "input": {"command": command}}
            ],
        },
    }


def reminded():
    return ran(REMINDER_COMMAND)


class StopPushCheckTest(unittest.TestCase):
    def run_hook(self, entries, active=False, raw=None):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder) / "session.jsonl"
            path.write_text("".join(json.dumps(entry) + "\n" for entry in entries))
            payload = raw if raw is not None else json.dumps(
                {"transcript_path": str(path), "stop_hook_active": active}
            )
            return subprocess.run(
                [str(SCRIPT)], input=payload, capture_output=True, text=True, check=False
            )

    def assertSilent(self, result):
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")

    def block_reason(self, result):
        self.assertEqual(result.returncode, 0)
        decision = json.loads(result.stdout)
        self.assertEqual(decision["decision"], "block")
        return decision["reason"]

    def assertBlocks(self, result):
        self.assertIn("PushNotification", self.block_reason(result))

    def assertBlocksForReminder(self, result):
        reason = self.block_reason(result)
        self.assertIn("Tarefas Claude", reason)
        self.assertIn("Claude espera: ", reason)

    def test_given_stop_hook_active_when_the_turn_ends_with_a_question_then_exits_silently(self):
        result = self.run_hook([user("go"), said("Which one?")], active=True)
        self.assertSilent(result)

    def test_given_a_final_message_ending_with_a_question_mark_and_no_push_when_the_turn_ends_then_blocks(self):
        result = self.run_hook([user("go"), said("Which one?")])
        self.assertBlocks(result)

    def test_given_a_question_mark_followed_by_whitespace_when_the_turn_ends_then_blocks(self):
        result = self.run_hook([user("go"), said("Which one?  \n")])
        self.assertBlocks(result)

    def test_given_the_word_responda_and_no_push_when_the_turn_ends_then_blocks(self):
        result = self.run_hook([user("go"), said("Responda sim para seguir.")])
        self.assertBlocks(result)

    def test_given_the_words_answer_sim_and_no_push_when_the_turn_ends_then_blocks(self):
        result = self.run_hook([user("go"), said("Answer sim to continue.")])
        self.assertBlocks(result)

    def test_given_a_push_notification_call_in_the_turn_when_the_turn_ends_with_a_question_then_exits_silently(self):
        entries = [user("go"), called("PushNotification"), tool_result(), reminded(), tool_result(), said("Which one?")]
        self.assertSilent(self.run_hook(entries))

    def test_given_a_push_notification_only_in_the_previous_turn_when_the_turn_ends_with_a_question_then_blocks(self):
        entries = [
            user("first"),
            called("PushNotification"),
            tool_result(),
            reminded(),
            tool_result(),
            said("Done."),
            user("second"),
            said("Which one?"),
        ]
        self.assertBlocks(self.run_hook(entries))

    def test_given_a_push_and_no_reminder_when_the_turn_ends_with_a_question_then_blocks_asking_for_the_reminder(self):
        entries = [user("go"), called("PushNotification"), tool_result(), said("Which one?")]
        self.assertBlocksForReminder(self.run_hook(entries))

    def test_given_a_reminder_and_no_push_when_the_turn_ends_with_a_question_then_blocks_asking_for_the_push(self):
        entries = [user("go"), reminded(), tool_result(), said("Which one?")]
        self.assertBlocks(self.run_hook(entries))

    def test_given_a_reminder_and_no_push_when_the_turn_ends_with_a_question_then_the_block_does_not_ask_for_the_reminder(self):
        entries = [user("go"), reminded(), tool_result(), said("Which one?")]
        self.assertNotIn("Tarefas Claude", self.block_reason(self.run_hook(entries)))

    def test_given_no_push_and_no_reminder_when_the_turn_ends_with_a_question_then_the_block_asks_for_both(self):
        reason = self.block_reason(self.run_hook([user("go"), said("Which one?")]))
        self.assertIn("PushNotification", reason)
        self.assertIn("Tarefas Claude", reason)

    def test_given_a_reminder_only_in_the_previous_turn_when_the_turn_ends_with_a_question_then_blocks_asking_for_the_reminder(self):
        entries = [
            user("first"),
            reminded(),
            tool_result(),
            said("Done."),
            user("second"),
            called("PushNotification"),
            tool_result(),
            said("Which one?"),
        ]
        self.assertBlocksForReminder(self.run_hook(entries))

    def test_given_a_reminder_in_another_list_when_the_turn_ends_with_a_question_then_blocks_asking_for_the_reminder(self):
        other = REMINDER_COMMAND.replace("Tarefas Claude", "Groceries")
        entries = [user("go"), called("PushNotification"), tool_result(), ran(other), tool_result(), said("Which one?")]
        self.assertBlocksForReminder(self.run_hook(entries))

    def test_given_a_reminder_without_the_waiting_prefix_when_the_turn_ends_with_a_question_then_blocks_asking_for_the_reminder(self):
        plain = REMINDER_COMMAND.replace("Claude espera: ", "")
        entries = [user("go"), called("PushNotification"), tool_result(), ran(plain), tool_result(), said("Which one?")]
        self.assertBlocksForReminder(self.run_hook(entries))

    def test_given_a_command_that_only_mentions_the_list_when_the_turn_ends_with_a_question_then_blocks_asking_for_the_reminder(self):
        mention = 'echo "Tarefas Claude Reminders Claude espera: "'
        entries = [user("go"), called("PushNotification"), tool_result(), ran(mention), tool_result(), said("Which one?")]
        self.assertBlocksForReminder(self.run_hook(entries))

    def test_given_a_final_message_that_asks_nothing_when_no_reminder_was_created_then_exits_silently(self):
        self.assertSilent(self.run_hook([user("go"), said("Done.")]))

    def test_given_tool_results_inside_the_turn_when_a_push_came_before_them_then_the_turn_still_counts_it(self):
        entries = [
            user("go"),
            called("PushNotification"),
            tool_result(),
            reminded(),
            tool_result(),
            called("Bash"),
            tool_result(),
            said("Which one?"),
        ]
        self.assertSilent(self.run_hook(entries))

    def test_given_a_final_message_that_asks_nothing_when_the_turn_ends_then_exits_silently(self):
        self.assertSilent(self.run_hook([user("go"), said("Done. PR is open.")]))

    def test_given_a_question_in_the_middle_of_the_message_when_the_final_line_is_a_statement_then_exits_silently(self):
        self.assertSilent(self.run_hook([user("go"), said("Why? Because it was broken. Fixed.")]))

    def test_given_a_final_tool_call_after_a_question_when_the_turn_ends_then_uses_the_last_text_message(self):
        entries = [user("go"), said("Which one?"), called("Bash"), tool_result(), said("Fixed it.")]
        self.assertSilent(self.run_hook(entries))

    def test_given_an_empty_turn_when_the_hook_runs_then_exits_silently(self):
        self.assertSilent(self.run_hook([user("go")]))

    def test_given_a_transcript_path_that_does_not_exist_when_the_hook_runs_then_exits_silently(self):
        raw = json.dumps({"transcript_path": "/nonexistent/session.jsonl"})
        self.assertSilent(self.run_hook([], raw=raw))

    def test_given_input_that_is_not_json_when_the_hook_runs_then_exits_silently(self):
        self.assertSilent(self.run_hook([], raw="not json"))

    def test_given_a_transcript_with_a_broken_line_when_the_turn_ends_with_a_question_then_still_blocks(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder) / "session.jsonl"
            lines = [json.dumps(user("go")), "{broken", json.dumps(said("Which one?"))]
            path.write_text("\n".join(lines) + "\n")
            payload = json.dumps({"transcript_path": str(path)})
            result = subprocess.run(
                [str(SCRIPT)], input=payload, capture_output=True, text=True, check=False
            )
        self.assertBlocks(result)


if __name__ == "__main__":
    unittest.main()
