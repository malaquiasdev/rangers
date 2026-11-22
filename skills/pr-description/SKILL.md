---
name: pr-description
description: Writes or rewrites a PR description in the repository's template, in the language of that template, so a junior can understand it: an analogy in "Why it exists", a validated Mermaid diagram in "How it works inside" and a Situation, What we expect, Why it matters table in the test checklist. Use when opening a PR (before gh pr create), when the human asks to rewrite or improve a PR description, or when /pr-description runs with a PR number.
---

# pr-description

The description is read by people who did not write the code. It explains the behavior and does not list function names. A list of identifiers ("`processX`, `processY` normalize...") and a checklist that only dumps numbers ("10/10, 342/342") are what reviewers call impossible to understand.

Write in the language of the repository's PR template.

Read the project's notes in the notes folder for local traps before you start.

## 1. Gather the material

- Template: `.github/pull_request_template.md` or `.github/PULL_REQUEST_TEMPLATE.md` (look for both spellings, in `docs/` and at the root). Use its own headings and keep their order. Without a template: What this PR does / Why it exists / How it works inside / What did not change / Test checklist / Points of attention for the reviewer / Closes.
- The diff (`gh pr diff <n>` or `git diff origin/main...HEAD`), the card in the tracker and the new tests: each test becomes one row of the test table.
- Rewriting an existing PR: `gh pr view <n> --json body --jq .body`, and keep what is already right (links, Closes, ticked checkboxes).

## 2. Write section by section

**What this PR does** — one or two tweets. The behavior before and after, in product words. Example: "the reference sent to the provider goes from `<uuid>` to `<region>-<uuid>`, and on the way back we strip the region before looking up the record".

**Why it exists** — starts with a concrete analogy (a luggage tag, a bank queue, a mail carrier). Then the risk if it is not done, or if it is done by half.

**How it works inside** — a Mermaid diagram first, text after.
- Flow between systems (front end, service, database, third party): `sequenceDiagram` with `autonumber`, a `rect` per phase and `alt`/`else` for the detours (error, old case, different format).
- Decision inside one service: `flowchart TD`.
- After the diagram, a short paragraph per phase ("**Going.**", "**Coming back.**"). A file name only when it helps to find the code. No list of functions.

**What did not change** — short bullets, each with the reason in half a sentence.

**Test checklist** — one sentence saying what is real and what is fake (count the fakes). Then the table:

| Situation | What we expect | Why it matters |
|---|---|---|
| a real production situation, in people's words | the observable result | what would break without it |

One row per test scenario. Below it, only the checkboxes: integration, unit, and pending manual validation (with the link of what blocks it). Numbers (10 tests, 342 unit) stay in the checkboxes, not in the table.

**Points of attention for the reviewer** — each item with a bold title that states the decision ("**Two checks on the bank, on purpose.**") and one sentence with the reason. Blockers and dependencies between PRs come first.

**Closes** — the tracker link of the card, or `Closes #N`.

Remove the sections the template says to remove when they do not apply. Tick the checkboxes only with what was really validated.

## 3. Validate before publishing

1. Each Mermaid block: save it to `<scratchpad>/x.mmd` and run `npx -y -p @mermaid-js/mermaid-cli@11 mmdc -q -i x.mmd -o x.svg`. No error and the `.svg` generated means it passed. A `;` inside a message breaks the parse: use a comma.
2. Zero inline comments in the diff: `git diff origin/main | grep -cE "^\+\s*(//|\*|/\*)"` must give 0.
3. Reread looking for a list of identifiers or a loose number without explanation. If there is one, rewrite it as a sentence.

## 4. Publish

- New PR: `gh pr create --title ... --body-file <file>` (`--body` does not apply the template by itself).
- Existing PR: `gh pr edit <n> --body-file <file>`.
- End with the attribution lines the session asks for.
- Send the human the PR link.
