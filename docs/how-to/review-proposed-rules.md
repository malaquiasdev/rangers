# How to review the rules a learning round proposed

A rules round leaves a local branch `learn/<stamp>` in a worktree and a reminder titled `Revisar as regras propostas` whose body names the branch and the worktree. This guide shows how to read the branch, keep it or drop it. Nothing is pushed until you do it.

1. Find the branch. Use the one in the reminder, or list them all from the root of the clone:

   ```sh
   git branch --list 'learn/*'
   ```

   If a reminder is missing, look in `~/.local/state/rangers/learn/pending-reminders.txt`.

2. Read what the round changed:

   ```sh
   git log --oneline main..learn/<stamp>
   git diff main...learn/<stamp>
   ```

   Each rule cites its origin: the session and the words the human typed.

3. Decide:

   - To keep every rule, merge the branch into `main`: `git merge learn/<stamp>`.
   - To keep some rules, edit the files in the worktree at `<clone>-worktrees/learn-<stamp>`, commit there, then merge.
   - To drop the round, skip the merge.

4. Remove the worktree and the branch:

   ```sh
   git worktree remove ../rangers-worktrees/learn-<stamp>
   git branch -D learn/<stamp>
   ```

   The path assumes the clone is a folder named `rangers`. The exact path is in the reminder and in the output of the round.
