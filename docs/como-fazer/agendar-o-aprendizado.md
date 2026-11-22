# How to schedule the learning job

This guide builds `rangers learn` and schedules it to run every 2 hours on macOS. What the command reads and writes is in the [`rangers learn` reference](../referencia/comando-learn.md).

The scheduled job runs with no flags, so it uses the default folders. Before you start, check that they match your machine: the clone at `~/Developer/me/rangers`, the transcripts at `~/.claude/projects` and the notes folder at `~/Documents/brain`. If one is somewhere else, the job will not find it.

1. From the root of the repository, build the program where the job expects it:

   ```sh
   mkdir -p ~/.local/bin
   go build -o ~/.local/bin/rangers ./cmd/rangers
   ```

2. Run it once by hand to see that it works. `--skip-notes` keeps this first run to the rules pass:

   ```sh
   ~/.local/bin/rangers learn --skip-notes
   ```

   It prints either a gate message, such as too few new lines, or the answer of the model and the branch it created. Both mean the program works.

3. Install the job:

   ```sh
   scripts/learn-schedule.sh install
   ```

   The job runs once right away and then every 2 hours.

4. Check it:

   ```sh
   scripts/learn-schedule.sh status
   tail ~/.local/state/rangers/learn/launchd.log
   ```

To see what the rounds cost, run `~/.local/bin/rangers learn metrics`.

## Update the program

After pulling changes to the Go code, rebuild with the command from step 1. The job runs the new binary on its next run.

`scripts/learn-schedule.sh` has no option for flags. To run a round with other values, such as a lower dollar cap, run `~/.local/bin/rangers learn` by hand with the flags.

## Stop the job

```sh
scripts/learn-schedule.sh uninstall
```

The state folder stays. Delete `~/.local/state/rangers/learn` by hand if you want the history gone.
