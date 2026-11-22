# Test plan

The measures and criteria that decide whether the thin layer works. Why the test exists is in [why a thin layer](../explanation/why-a-thin-layer.md#the-test-decides). The steps are in [how to run the test](../how-to/run-test.md).

## The activity

- Real work, picked by the owner of the repository.
- Small: it fits in one pull request.
- Its context lives in the repository of the activity, not in this one.

## Measures

| Measure | How to record it | Passes if |
| --- | --- | --- |
| How many times the human had to say yes | Count in the session | 2 or fewer |
| Saw the agents running in the app | Yes or no | Yes |
| Read the session text without opening another file | Yes or no, and where it got stuck | Yes |
| Time from the activity to the pull request ready | Minutes | No fixed target; record it |
| Quota spent | Percentage before and after | Record it |
| Something went wrong and nobody said so | Describe it | No case |

## Where to record

In `resultados/YYYY-MM-DD.md`, at the root of this repository, with the date of the test. The folder does not exist yet; the first test creates it.

## Decision criteria

| Result | Condition | What to do |
| --- | --- | --- |
| Works | Passed the first three measures and no error went unreported | Add only what was missing |
| Works in part | One measure failed | Record which one and fix only that, with a new test |
| Does not work | Two or more measures failed | Back to the design |

## After the test

If it worked, decide whether the gate program comes in. It only comes in if the test showed it was missing.
