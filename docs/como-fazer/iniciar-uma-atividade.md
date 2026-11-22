# How to start an activity

This guide starts a piece of work with the coordinator through the `/rangers` skill. You need the agents and the skill connected (see [how to connect the agents and the skill](ligar-os-agentes.md)).

1. Open a Claude Code session.
2. Type `/rangers` followed by one of these:
   - a Linear issue, as an id such as `ABC-123` or a linear.app link;
   - a GitHub issue, as a github.com link or `owner/repo#12`;
   - a short description of the work.
3. Answer the coordinator's questions. It asks one at a time, each with the answer it recommends.
4. Open the plan link it sends and answer the question with one yes, or say what to change. Nothing starts without the yes.
5. Wait for the pull request link. The coordinator stops there and tells you what is left: read the pull request, merge it and deploy it.

If the coordinator cannot tell which repository the work belongs to, it asks. It looks for the repository two levels below `~/Developer/`, such as `~/Developer/<group>/<repository>`.

What each agent does along the way is in [how the agents split the work](../explicacao/como-os-agentes-dividem-o-trabalho.md).
