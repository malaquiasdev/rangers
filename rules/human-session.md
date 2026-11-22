# Rules for every agent

- The human reads on a tablet and neither opens folders nor runs commands. Deliver a link (artifact, PR, page), never only a local path, and run any command that has to be run yourself.
- A test of a new agent or flow runs in another session, opened by the human. Do not run the test in the session that coordinates: it contaminates the result.
- A request to look at the session or the work of another agent is a request to read and evaluate, not to redo. Point out the error, and rewrite only when asked to.
- Make the estimate proportional to the size of the change: one line of code does not take 30 minutes.
- In the human's personal repositories, write everything that goes to GitHub (code, identifiers, test names, commits, PR titles and bodies, issues, README, docs) in English: it costs fewer tokens. In the repositories of the human's work, follow the language of the repository, which is Portuguese. Chat with the human stays in pt-BR.
- In the human's personal repositories, run CI checks locally (hooks, scripts) and keep only release-please in GitHub Actions, to save the Actions quota.
- Public docs and READMEs cite nothing the reader cannot find: no private or work project as an example or reference.
- Keep the settings of a project in one place, not spread across files: spread settings are hard to change later.
- Change repository settings that carry no risk (description, topics, merge options) yourself with `gh` or `gh api` instead of asking the human to open the settings page. Visibility, deletion and permissions still need the human's yes. When a setting has no API, say so and send the direct link to the exact page and field.
