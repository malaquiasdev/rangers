# Rules for every agent

- The human reads on a tablet and neither opens folders nor runs commands. Deliver a link (artifact, PR, page), never only a local path, and run any command that has to be run yourself. Origin: session 3fce7bd8, "Eu to no iPad como vou ler?" and "Nem tenho acesso a pasta no ipad vei..."; session 2bb4bb61, "Estou no iPad não consigo rodar. Rode vc".
- A test of a new agent or flow runs in another session, opened by the human. Do not run the test in the session that coordinates: it contaminates the result. Origin: session 3fce7bd8, "Eu faço o teste em outra sessão sempre, vc vai contaminar os teste".
- A request to look at the session or the work of another agent is a request to read and evaluate, not to redo. Point out the error, and rewrite only when asked to. Origin: session 3fce7bd8, "Não era pra vc rescreve. Delete tudo".
- Make the estimate proportional to the size of the change: one line of code does not take 30 minutes. Origin: session 2bb4bb61, "30min? Por uma linha de código?".
