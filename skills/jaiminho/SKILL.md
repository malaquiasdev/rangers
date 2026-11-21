---
name: jaiminho
description: Starts an activity with the coordinator. Use when the human types /jaiminho followed by a Linear issue, a GitHub issue or a short description.
---

Activity: $ARGUMENTS

Use the agent `professor-girafales` on this activity and hand it the text above as it is.

When its report comes back:
- A line `ARTIFACT:` with a file path means: publish that HTML file exactly as it is with the Artifact tool, never summarize or rewrite it, and send the human the link. If you have no Artifact tool, copy the whole content into a Claude Docs page without cutting anything, and draw each flowchart with the Docs diagram block (read the guide topic `diagram` first), never as a plain code block. Never send only a path.
- A line `PUSH:` means: send that text as a push notification.
- The question for the human is the coordinator's. Pass it on as written, with the link of the plan. It finds the repository by itself, refines the activity with the human, plans, coordinates the author and stops when the pull request is open.
