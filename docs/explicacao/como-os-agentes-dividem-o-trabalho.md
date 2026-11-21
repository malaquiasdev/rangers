# Como os agentes dividem o trabalho

Cada agente tem um papel só e um limite claro do que não faz. A divisão segue uma ideia simples: quem escreve o código não revisa, quem revisa não edita, e nenhum agente leva nada para produção sem o sim explícito do humano.

Os papéis estão nos arquivos de [`agents/`](../../agents/). O campo `description` de cada um começa pelo nome do papel em inglês (Coordinator, Author, Reviewer e assim por diante). O formato do arquivo está na [referência do arquivo de agente](../referencia/arquivo-de-agente.md).

## O caminho de uma atividade

1. O coordenador (Coordinator) refina a atividade com o humano usando a skill `grilling`, uma pergunta por vez. Escolhe o roteiro do pstack que serve, escreve o plano e pede um único sim, mostrando o resumo e um hash de 12 caracteres do plano. Sem esse sim, nada começa.
2. O autor (Author) implementa uma unidade no próprio worktree. Escreve o teste que falha antes do código, faz commits pequenos e abre a própria PR. Não faz merge nem deploy.
3. O revisor (Reviewer) lê a PR numa sessão nova, sem o contexto do autor. Cada achado precisa de prova e de uma gravidade: BLOQUEIA, MÉDIO ou NIT. Só um BLOQUEIA com prova segura o merge. São no máximo duas rodadas.
4. O portão (Gate) roda o programa `jaiminho gate`, que confere o sim do humano, a CI e produção antes do merge. A decisão é do programa, não do modelo. Esse programa ainda não existe neste repositório (veja [o que ainda não tem dono](por-que-uma-camada-fina.md#o-que-ainda-não-tem-dono-em-open-source)).
5. O QA testa em dev o que foi mergeado, pela porta de entrada do serviço, e registra a evidência completa. Só lê o banco, dentro de uma transação somente leitura.

## Os papéis fora do caminho

O explorador (Explorer) responde perguntas sobre código e dados sem mudar nada, com `caminho:linha` para cada afirmação. Ele tem dois modos: `find`, para achar onde algo está, e `explain`, para explicar como funciona e o que quebra se mudar.

O agente de aprendizado (Learning) lê as transcrições de sessão que mudaram desde a última vez e transforma correções do humano em regras. Ele abre uma PR neste repositório com as propostas e o humano aprova ou recusa cada regra. Quando a PR está aberta, ele avisa o humano em uma linha com o link: por notificação, se a ferramenta estiver disponível na execução agendada, ou por um lembrete na lista "Tarefas Claude". O arquivo dele prevê uma tarefa agendada que o roda duas vezes por dia. Essa tarefa não está neste repositório.

O agente de documentação (Documentation) organiza a pasta `docs/` de um repositório pelo método Diátaxis. Só o humano o chama, e ele nunca abre PR nem muda código.

## Quem avisa o humano

O humano lê num tablet e acompanha os agentes pelo app. Quem manda notificação é o coordenador: quando o plano espera o sim, quando uma PR abre, quando um agente para, trava ou termina sem resultado, quando o revisor acha um BLOQUEIA e quando algo fica pronto para o humano testar. Agentes com ferramentas limitadas, como o revisor e o explorador, não notificam. Eles respondem ao coordenador, e o coordenador avisa.

## Por que os limites são escritos como proibições

Cada arquivo tem uma seção "What you do not do". Ela existe porque o erro mais caro de um agente é fazer além do pedido: um merge que ninguém autorizou, uma edição durante a revisão, um palpite sem prova. Escrever o limite no próprio agente deixa a regra perto de quem pode quebrá-la.

## Modelos e esforço

Cada agente declara o modelo e o esforço no próprio arquivo. Papéis que pedem julgamento amplo, como coordenar, aprender e documentar, rodam num modelo maior. Papéis de execução, como autor, revisor e QA, rodam num modelo intermediário. O portão roda no menor, porque só executa um programa e repete o que ele disse. Os valores exatos estão nos arquivos de `agents/`.

## A voz de cada agente

Cada arquivo tem uma linha `Character:` que dá ao agente uma voz própria. A voz serve para a conversa entre agentes. A maioria dos arquivos proíbe o nome do personagem em código, commits, PRs, comentários e documentos. Cada arquivo diz se a regra vale para ele.
