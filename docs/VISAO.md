# DataVault — Visão do projeto

> Documento vivo. Registra o que já foi decidido e o que ainda está em aberto.

## O que é

Um cofre local, de código aberto, para segredos de quem vive no terminal:
logins, chaves de API, tokens, chaves SSH, seeds/passphrases, 2FA, notas seguras e comandos de servidor.

- **Local-first:** os dados ficam só na máquina do usuário. Sem nuvem, sem conta, sem telemetria.
  A única conexão de rede é a checagem de versão nova (a cada início, desligável)
  e o `dtvlt update`.
- **Plataformas:** Windows e Linux.
- **Nomes:** o produto se chama **DataVault**. O comando é **`dtvlt`**, que aparece só no CLI.
- **Licença:** MIT (ver [`LICENSE`](../LICENSE)).

## Fluxo de uso

```
$ dtvlt
  [arte ASCII]                                   v0.7.0
  update  0.7.1 available · type update          ← só se houver versão nova
  pin: ****                                      ← só se o PIN estiver ativado
  session 7f3a-91c2
  dtvlt › ,,,                                    ← abre o navegador

navegador
  [arte ASCII gótica]
  _                                 ← 1º: palavra-chave
  _                                 ← 2º: senha
  → vault aberto
```

1. O usuário roda `dtvlt` e vê a arte ASCII com a versão.
2. Se houver versão nova, aparece o aviso pedindo para digitar `update`.
3. Se o PIN estiver ativado, ele é pedido para destravar o console.
4. O CLI sobe (ou reencontra) a sessão e abre o **console**, que espera comandos.
   O navegador **não** abre sozinho: só quando o usuário digita `,,,`.
5. O frontend mostra a tela de login: primeiro a palavra-chave, depois a senha.
6. Com as duas certas, o vault é descriptografado e o frontend é liberado.

## Configuração inicial (primeiro `dtvlt`)

Assistente no próprio terminal:

| Opção | Valores | Padrão |
|---|---|---|
| PIN no CLI | ativado / desativado | desativado |
| Manter o terminal aberto | sim / não (roda em segundo plano) | sim |
| Auto-lock por inatividade | minutos | 15 |
| Limpar a área de transferência | segundos | 20 |

Em seguida o usuário cria a palavra-chave e a senha do vault, com aviso claro:
**não existe recuperação.** Se perder as duas, perde o vault.

As respostas ficam no `config.toml`. Tudo pode ser mudado depois na tela **settings**
do DataVault (com o vault aberto) ou com `dtvlt config`.

## Sessão

O processo do `dtvlt` é o dono da sessão.

- Cada execução gera um **Session ID** aleatório e curto (ex.: `7f3a-91c2`),
  exibido no terminal e no canto da tela do frontend. Se os dois batem, o usuário
  sabe que a aba fala com a instância dele.
- A chave descriptografada existe **só na memória desse processo**, nunca em disco.
- Encerrou o processo (`Ctrl+C`, fechou o terminal, `dtvlt stop`): a sessão acaba,
  as chaves são apagadas da memória e o frontend volta para a tela de login.
- **Terminal aberto (padrão):** o terminal funciona como log discreto da sessão:
  ```
  14:02  unlocked
  14:05  copied  github/token   · clears in 20s
  14:20  locked  (idle 15m)
  ```
- **Terminal fechado:** o processo vai para segundo plano. `dtvlt status` mostra
  a sessão ativa, `dtvlt lock` trava o vault e `dtvlt stop` encerra tudo.
- **Escrita rápida:** `dtvlt note "título" "conteúdo"` e `dtvlt task "título"
  "detalhe"` gravam na sessão aberta pela API de controle (o mesmo token de
  `status`, `lock` e `stop`), sem abrir o navegador. A nota entra no Unspoken, sem
  selo (selar pede o PIN, que fica no navegador), e a tarefa entra em Orders no
  Siege do dia. O segundo argumento é opcional, e `"título(conteúdo)"` também vale.
  Com o vault travado, o comando avisa e não grava nada. O log da sessão não mostra
  o título das notas.

## Tela de login

- Arte ASCII em letra gótica escrevendo **DataVault**, no estilo do print de referência.
  Animações ficam para depois.
- **Um campo de cada vez.** Um cursor `_` pisca devagar onde o usuário deve digitar.
- `Enter` avança da palavra-chave para a senha. `Esc` volta.
- A senha não mostra caracteres nem bolinhas.
- Erro em uma única linha discreta, sem dizer qual dos dois estava errado:
  `invalid credentials`.
- Atraso crescente entre tentativas erradas.

## Modelo de segurança

**A segurança real é o login (palavra-chave + senha).** As duas entram na chave.

```
entrada = len(palavra-chave) ‖ palavra-chave ‖ len(senha) ‖ senha
KEK     = Argon2id(entrada, salt)                   # chave derivada do login
DEK     = chave aleatória de 256 bits               # chave real dos dados
arquivo = cabeçalho ‖ XChaCha20-Poly1305(KEK, DEK) ‖ XChaCha20-Poly1305(DEK, dados)
```

- **Argon2id** com parâmetros guardados no cabeçalho (padrão inicial:
  64 MiB, 3 iterações, paralelismo 4), ajustáveis no futuro sem quebrar vaults antigos.
- **KEK/DEK separadas:** trocar a senha só recriptografa a DEK, não o vault inteiro.
- O cabeçalho (versão, salt, parâmetros) é autenticado como AAD, então não pode ser adulterado.
- Escrita atômica (arquivo temporário + rename) e um backup da versão anterior.
- Nada de criptografia caseira. Só bibliotecas consolidadas.

**O PIN do CLI é uma trava de conveniência e não protege os dados.** Ele só impede
que alguém na máquina inicie o `dtvlt` casualmente. Quem tiver o arquivo do vault
continua precisando da palavra-chave e da senha. A documentação pública vai dizer
isso claramente. O PIN fica salvo como hash Argon2id na config.

### Servidor local (127.0.0.1)

Outros processos e outras abas do navegador também alcançam portas locais. Por isso:

- O servidor escuta só em `127.0.0.1`, numa porta aleatória.
- O console (`,,,`) abre o navegador com um **token de uso único** (validade de 60 s), trocado
  por um cookie `HttpOnly` e `SameSite=Strict`. Sem esse token, nenhuma requisição é aceita.
- Os headers `Host` e `Origin` são validados (contra DNS rebinding e CSRF).
- Headers de segurança: CSP restrita, sem recursos externos e sem CDN. Tudo embutido no binário.

### Outros cuidados

- A área de transferência é limpa depois de N segundos.
- Auto-lock por inatividade.
- Segredos não aparecem em logs nem em mensagens de erro.

## Tipos de item

| Tipo | Campos principais |
|---|---|
| Login | serviço, usuário, senha, URL |
| Chave de API / token | serviço, chave, escopo, expiração |
| Chave SSH | nome, privada, pública, passphrase |
| Seed / passphrase | nome, palavras |
| TOTP (2FA) | serviço, secret → gera o código |
| Nota segura | texto livre |
| Comando de servidor | comando (várias linhas), servidor, diretório; só guarda e copia, nunca executa |
| Códigos de recuperação | conta e um campo por código; cada código se copia sozinho e pode ser marcado como morto |

Todos os itens têm tags e campos personalizados.

## Estética

O clima é de ferramenta de terminal séria, **sem "hacker cringe"**, com um tom gótico
contido (o tema **souls**: ruínas, lápides, ossos e livros antigos como natureza-morta).

- **Sim:** monocromático (branco e cinza sobre preto), uma cor de destaque usada com
  muita parcimônia, fonte monoespaçada de qualidade, linguagem seca e técnica no estilo
  de CLIs modernos (`gh`, `fly`, ferramentas da Charm), animações lentas e sutis,
  imagens paradas.
- **Não:** verde Matrix, "ACCESS GRANTED", glitch, caveiras animadas ou "assustadoras",
  chuva de caracteres, mensagens dramáticas.

Os detalhes (cores, tipografia, layout, ícones, arte ASCII, tom e terminal) estão no
[guia de estilo](ESTILO.md).

## Stack

| Parte | Escolha | Motivo |
|---|---|---|
| Linguagem | **Go** | Binário único, compila para Windows e Linux sem complicação, código fácil de ler para quem quiser contribuir. |
| Criptografia | `golang.org/x/crypto` (`argon2`, `chacha20poly1305`) | Bibliotecas padrão e auditadas do ecossistema Go. |
| CLI | `golang.org/x/term` + ANSI simples | Por enquanto bastam prompts de linha. Bubble Tea (Charm) fica para quando o CLI ganhar telas interativas. |
| Frontend | Svelte + Vite | Leve e sem runtime pesado. O build é embutido no binário com `go:embed`. |
| Armazenamento | Um único arquivo criptografado | Simples de entender, de fazer backup e de mover. |

**Ponto de atenção do Go:** o coletor de lixo pode deixar cópias de segredos na memória.
Mitigação: manter segredos em `[]byte` (nunca `string`), zerar os buffers depois do uso
e avaliar uma biblioteca de memória protegida (ex.: `memguard`).
Hoje a chave (KEK/DEK) segue essa regra, mas os itens descriptografados e as
credenciais recebidas pela API ainda passam por `string` (JSON). Melhorar isso está na lista.

### Onde ficam os arquivos

| | Windows | Linux |
|---|---|---|
| Vault | `%APPDATA%\DataVault\vault.dtv` | `~/.local/share/datavault/vault.dtv` |
| Config | `%APPDATA%\DataVault\config.toml` | `~/.config/datavault/config.toml` |
| Binário instalado | `%LOCALAPPDATA%\Programs\DataVault\dtvlt.exe` | `~/.local/bin/dtvlt` |

## Distribuição e versões

- Quem usa baixa **um arquivo só** da página de Releases. Não precisa de Go nem de Node:
  o frontend vai embutido no binário.
- O `.exe` baixado é só o instalador: aberto, ele se copia para a pasta do usuário,
  entra no PATH (sem pedir admin) e fecha. Uma versão mais nova, aberta do mesmo jeito,
  substitui a instalada. Depois disso, quem usa só digita `dtvlt` no terminal.
- Versionamento semântico. A versão vem da tag git (`v1.2.3`), é embutida no build e
  aparece em `dtvlt version` e no rodapé do login.
- A tag dispara o GitHub Actions, que testa, compila Windows e Linux (amd64 e arm64)
  e publica a release com `SHA256SUMS` e as notas do `CHANGELOG.md`.
- A versão do **formato do arquivo** (byte no cabeçalho do vault) é separada da versão
  do app. Uma mudança que versões antigas não conseguem ler exige versão MAJOR (depois do 1.0).

## Páginas

A barra lateral tem uma página por ícone. Os nomes seguem a estética do Grimoire.

| Ícone | Página | O que é |
|---|---|---|
| livro | **Grimoire** | credenciais, chaves, comandos de servidor |
| coração partido | **Unspoken** | notas íntimas, o que não foi dito em voz alta |
| espada | **Siege** | kanban das tarefas do dia: ir para a guerra |
| cruz | **Crypt** | documentos e arquivos guardados criptografados |
| saco de moedas | **Ledger** | assinaturas e gastos fixos |

### Grimoire: forge

O botão **forge**, à esquerda do `+ new` (tecla `g`), abre o gerador de senhas no
painel da direita.

- **Senha:** cada caractere sai de `crypto.getRandomValues`, sem viés, e cada tipo
  escolhido aparece pelo menos uma vez.
- **Opções:** tamanho de 8 a 128 (24 por padrão), maiúsculas, minúsculas, dígitos e
  símbolos, e "no lookalikes", que tira `0 O o 1 l I |` para senhas digitadas à mão.
- **Força:** em bits (tamanho × log2 do conjunto de caracteres).
- **Copiar:** como os outros campos, limpa a área de transferência depois do tempo
  configurado. A senha não é guardada em lugar nenhum.
- **Terminal:** `dtvlt pwgen [tamanho] [--plain] [--no-upper|--no-lower|--no-digits|--no-symbols]`
  gera a senha com as mesmas regras (`internal/forge`), sem sessão. No console é
  `pwgen`. Com a saída redirecionada, imprime só a senha.
- **Depois:** frases-senha (palavras aleatórias), que pedem uma lista de palavras no
  app.

### Grimoire: códigos de recuperação

Os códigos de backup que os serviços dão ao ligar o 2FA (GitHub, Google,
Discord...) para entrar quando o autenticador não está à mão.

- **Entrada:** tipo **recovery codes**, com a conta e um campo secreto por código.
  No editor, **+ code** acrescenta o próximo (`code 2`, `code 3`...).
- **Morto:** cada código tem **copy** e **dead**, que o marca como gasto. Um código
  morto fica riscado e não se copia mais; **revive** desfaz. A marca é o `dead` do
  campo, salvo no vault.
- **Aviso:** o topo mostra quantos sobram ("3 of 10 codes"). Com 2 ou menos, avisa
  que é hora de gerar códigos novos no serviço.
- **0.15.0:** entradas com todos os códigos num campo `codes`, um por linha, continuam
  funcionando: cada linha vira um código com **copy** e **dead**, e a marca é a linha
  começar com `dead` (ou `used`, como a 0.15.0 gravava).

### Grimoire: renovação

Algumas chaves vencem e precisam ser trocadas. Cada entrada do Grimoire pode ter:

- **renew every:** renovar a cada N dias, contando da última renovação (ou da
  criação, se nunca foi renovada);
- **renew by:** uma data fixa, que vale mais que o intervalo.

A lista marca em âmbar o que vence em até 14 dias e o que já venceu, e o topo mostra
quantas entradas precisam de renovação. No detalhe, o botão **renewed** registra a
renovação de hoje: o intervalo recomeça e uma data fixa, já usada, é apagada.

### Unspoken

Um lugar para escrever o que dói e ninguém mais lê. Muita gente usa o app de notas
do celular para isso; aqui o texto fica criptografado no vault.

- **Nota:** título opcional, texto livre, datas de criação e de edição. Na lista,
  aparece o título (ou a primeira linha) e a data.
- **Escrever:** editor de texto simples, monoespaçado, que salva sozinho quando você
  para de digitar e ao sair da nota. `n` cria uma nota nova.
- **Selo por nota:** como no Grimoire. Uma nota selada pede o PIN de selo para abrir,
  e volta a selar ao sair dela.
- **Dead Words:** uma nota pode ser **enterrada**. Ela sai da lista principal e vai
  para a seção Dead Words, onde continua guardada para reler, e pode ser desenterrada.
  Apagar de vez só existe dentro de Dead Words, com confirmação.
- **Detalhes de texto:** "dead words" também aparece como texto de clima, por exemplo
  na tela vazia e como dica do editor.
- **Dados:** uma coleção própria de notas dentro do payload criptografado do vault.
  As regras do selo (token por nota, que morre no lock) são as mesmas das credenciais.

### Siege

O plano de batalha do dia. **Um quadro por dia**, com quatro colunas:

| Orders | In Battle | Conquered | Fallen |
|---|---|---|---|
| a fazer | sendo feito | concluído | cancelado |

- **Tarefa:** título, nota opcional, coluna, posição na coluna e o dia a que pertence.
- **Dias:** o quadro abre no dia de hoje. Dá para navegar para os dias anteriores
  (histórico) e voltar para hoje. Um dia novo começa vazio e oferece trazer o que
  ficou em Orders e In Battle no dia anterior.
- **Mover:** arrastar com o mouse entre colunas e dentro da coluna, e também pelo
  teclado: setas para escolher a tarefa, um atalho para levá-la à coluna ao lado,
  `n` para uma ordem nova, `Enter` para editar.
- **Contagem:** o topo do quadro mostra quantas tarefas há em cada coluna.
- **Dados:** uma coleção de tarefas dentro do payload criptografado do vault.

### Crypt

Um lugar para guardar documentos e arquivos (contratos, documentos escaneados,
chaves `.pem`, comprovantes), junto com o resto.

- **Arquivos:** aparecem como blocos com o tipo e o tamanho; o título aparece por
  cima do arquivo ao passar o mouse. Um arquivo sem título usa o nome sem extensão.
- **Enterrar:** `+ new`, a tecla `n` ou arrastar arquivos para a página. Até 25 MB
  por arquivo.
- **Detalhe:** nome do arquivo, tipo, tamanho e data. **take** baixa uma cópia,
  **rename** troca o título e **destroy** apaga de vez, com confirmação.
- **Dados:** cada arquivo é criptografado com a chave do vault (XChaCha20-Poly1305,
  com o id do arquivo no AAD) num arquivo próprio, na pasta `crypt` ao lado do vault.
  Só a lista (título, nome, tipo, tamanho, datas) fica no payload do vault, que assim
  continua leve. O conteúdo nunca é mostrado dentro da página: sempre sai como
  download.

### Ledger

O livro-razão das assinaturas e gastos fixos: streaming, servidores, domínios,
academia. Mostra quanto vai embora todo mês e o que cobra nos próximos dias.

- **Conta:** nome, valor, moeda (código de três letras, como BRL ou USD), ciclo
  (mensal, anual ou semanal), data da primeira cobrança, cartão ou conta usada e
  uma nota.
- **Próxima cobrança:** calculada na página a partir da primeira cobrança e do
  ciclo, no calendário do navegador. Num mês sem o dia (31 em fevereiro) a cobrança
  cai no último dia.
- **Lista:** em ordem de cobrança; o que vence em até 7 dias fica destacado.
- **Totais:** no topo, por mês e por ano, separados por moeda. Contas anuais
  contam 1/12 por mês, e semanais 52/12.
- **Strike:** apaga a conta (com confirmação), para quando ela for cancelada.
- **Dados:** uma coleção dentro do payload criptografado do vault; o valor é
  guardado em centavos.
- **Depois:** as cobranças no calendário, quando ele existir.

Cada página deve sair numa versão MINOR própria.

## Escopo

### MVP

- Setup inicial no terminal
- `dtvlt` → PIN opcional → servidor local → navegador
- Login de um campo por vez
- Vault criptografado: criar, ler, editar e apagar itens (todos os tipos acima, exceto TOTP)
- Busca e tags
- Copiar com limpeza automática da área de transferência
- Auto-lock, `dtvlt lock`, `dtvlt stop`, `dtvlt status`

### Depois

- TOTP
- Comandos diretos sem navegador: `dtvlt get github --copy`, `dtvlt add ssh ~/.ssh/id_ed25519`
- `dtvlt` atuando como **ssh-agent**
- Gerador de senhas e passphrases
- Exportar e importar (incluindo de outros gerenciadores)
- Animações da arte ASCII

### Fora do escopo (por enquanto)

- Sincronização entre máquinas
- macOS
- Aplicativo mobile

## Em aberto

- [ ] Arte ASCII do CLI: qual texto e qual estilo?
