# How to use Rangers from a tablet or a phone

This guide runs an activity from a tablet or a phone while the agents work on your computer. The link between the two is Remote Control, a Claude Code feature: the session runs on the computer, and the Claude app or [claude.ai/code](https://claude.ai/code) shows it on the other device. The details of Remote Control are in the [official Remote Control documentation](https://code.claude.com/docs/en/remote-control).

Before you start, you need:

- the computer set up as in [the setup tutorial](../tutorials/prepare-the-machine.md), with the agents and the skill connected (see [how to connect the agents, the rules and the skill](connect-agents.md));
- Claude Code on the computer signed in with a claude.ai account on a paid plan. Remote Control does not work with an API key;
- on the tablet or phone, the Claude app signed in with the same account, or a browser open at claude.ai/code.

## 1. Start the session on the computer

1. On the computer, open a terminal in the folder that holds your repositories, such as `~/Developer`. The coordinator looks for the repository of the activity two levels below it, as in `~/Developer/<group>/<repository>`.
2. Start Remote Control:

   ```sh
   claude rc --permission-mode bypassPermissions --name rangers
   ```

   - `claude rc` is the short form of `claude remote-control`. It starts a server that waits for connections from the app or the browser and keeps running in the terminal.
   - `--name rangers` is the title of the session in the session list of the app and of claude.ai/code. Any name works; pick one you will recognize.
   - `--permission-mode bypassPermissions` starts the sessions with no permission prompts, so the agents run commands, edit files and open pull requests without stopping to ask you while you are away.

   The cost: in this mode nothing asks before a command runs, not even a harmful one or one planted in text an agent read, and the official documentation recommends it only for isolated containers or virtual machines. For a background safety check with few prompts, use `--permission-mode auto` instead. Without the flag, each prompt waits until you answer it on the tablet.

3. Answer the questions the first run asks. The first time, Claude Code asks `Enable Remote Control? (y/n)`; answer `y`. In a folder you have not trusted yet, it asks `Trust <directory>? [y/N]`; answer `y`. If it asks for a spawn mode, `1` (`same-dir`) is fine.

You should see the server waiting, with a link to claude.ai/code. Press the space bar to show a QR code.

If it fails:

- `You must be logged in to use Remote Control`: run `claude auth login` and choose the claude.ai option.
- Another message: look it up under [Troubleshooting in the Remote Control documentation](https://code.claude.com/docs/en/remote-control#troubleshooting). `claude doctor` tells which check failed.

## 2. Keep the computer awake

The agents run on the computer, not on the tablet. While the computer sleeps, nothing runs: no agent works and no message reaches you. When it wakes up and is online again, Claude Code reconnects by itself.

- Keep the computer plugged in and turn off sleep while you work away from it. On macOS you can start the server under `caffeinate`, which holds sleep off for as long as the command runs:

  ```sh
  caffeinate -is claude rc --permission-mode bypassPermissions --name rangers
  ```

  `-s` works only while the computer is on power.
- Leave the terminal open. Closing it stops the server and the session goes offline.
- If the computer stays awake but loses the network for about 10 minutes, the server gives up and exits.

If the server stopped, run `claude rc` again in the same folder, with the same flags, within about four hours. It brings back the sessions it was serving.

## 3. Open the session on the tablet or phone

Use one of these:

- Scan the QR code from the terminal with the phone. It opens the session in the Claude app.
- In the Claude app, tap **Code** and pick the session by its name.
- In a browser, open claude.ai/code and pick the session by its name.

You should see the session with a computer icon and a green dot, which means it is online. What you type on the tablet goes to the session on the computer.

If the session is missing or shows **Can't reach**, check on the computer that the server is still running and that the computer is awake and online.

## 4. Start an activity

1. In the session, type `/rangers` followed by a Linear issue, a GitHub issue or a short description, as in [how to start an activity](start-activity.md):

   ```text
   /rangers ABC-123
   ```

2. Answer the coordinator's questions. It asks one at a time, each with the answer it recommends, in short Brazilian Portuguese. A one-word answer is enough.
3. Wait for the plan. The coordinator writes it, and the courier (the agent that delivers pages and notifications) publishes it as a private page and sends you its link. A question comes with it: the summary of the plan, a short code that identifies this version of the plan, and a request for one yes. You do nothing with the code; it ties your yes to that version.
4. Answer in the session with `sim` (or `yes`) to start the work, or write what to change. Nothing starts without the yes.

After the yes, the agents build, test and open the pull request on the computer. The coordinator stops when the pull request is open and sends its link, with what is left for you to do: read it, merge it and deploy it.

## 5. Read the plan on the tablet

The plan link opens in the tablet browser. The page is private to your Claude account, so open it signed in with the same account. There is no file to find and no folder to open: the courier publishes the whole plan, as it was written, as a page. If the page tool is not available in that session, the plan goes to a Claude Docs page instead, which also opens from a link.

The page opens with a three-line summary in Brazilian Portuguese: what changes, what you will see and the risk. Then come the scenarios, a flowchart when the change has a flow, and a table of the data before and after when it touches data. The technical detail for the author sits in a closed block at the end.

If you get a file path instead of a link, the page was not published. Ask in the session for the link of the plan.

## 6. Get the notifications

When the coordinator needs you, or something important happens (the plan waits for your yes, a pull request opened, an agent stopped or got blocked, something is ready to test), the courier sends two things:

- A push notification with the action and the link. According to the official documentation, a Remote Control session can send a push to the Claude mobile app. It needs the app installed and signed in with the same account, the notification permission allowed on the device, and the setting **Push when Claude decides** turned on. To turn it on, run `claude` in a terminal on the computer, type `/config`, turn the setting on and close that session. A push does not show while you have the app open, and Focus modes on iOS can hold it back.
- A reminder in the Reminders list `Tarefas Claude`, due one minute later, when you must answer. The courier creates it with the Reminders app on the computer, so this part needs macOS and a list with exactly that name. Create the list in the Reminders app if it does not exist. The reminder reaches the tablet only if the tablet syncs the same Reminders list. When you answer, the coordinator marks the reminder as done.

If neither arrives, the messages are still in the session. Open it on the tablet and read the last message.

## 7. Know what the tablet can and cannot do

You can:

- type messages, answer questions and say yes or no;
- open every link the agents send: plans, pull requests, issues;
- see the agents running in the background, and stop one from the session;
- attach a photo or a file to a message. Claude Code copies the file to the computer.

You cannot:

- open folders or files of the computer. Ask for a link instead of a path;
- run a command. A line that starts with `!` goes to Claude as a message, not to a shell. Ask the session to run it;
- use the commands that work only in the terminal, such as `/plugin` or `/resume`. The list is under [Limitations in the Remote Control documentation](https://code.claude.com/docs/en/remote-control#limitations).

Every command, test, commit and pull request runs on the computer, with its tools and logins.

## 8. Know what stays safe

The gate is a program with no model that checks a pull request before a merge (see [how the agents split the work](../explanation/how-agents-split-work.md)). It stops at a release pull request or one labeled `production`, even after a yes, and leaves production to you. A yes to merge is bound to the head commit of the pull request, so a push after the question makes that yes useless and a new question is needed. The checks are in [the `rangers gate` command reference](../reference/gate-command.md).
