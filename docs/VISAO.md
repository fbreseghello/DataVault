# DataVault — Visão do projeto

> Documento vivo. Registra o que já foi decidido e o que ainda está em aberto.

## O que é

Um cofre local, de código aberto, para segredos de quem vive no terminal:
logins, chaves de API, tokens, chaves SSH, seeds/passphrases, 2FA, notas seguras e comandos de servidor.

- **Local-first:** os dados ficam só na máquina do usuário. Sem nuvem, sem conta, sem telemetria.
  A única conexão de rede é a checagem de versão nova (no máximo uma vez por dia, desligável)
  e o `dtvlt update`.
- **Plataformas:** Windows e Linux.
- **Nomes:** o produto se chama **DataVault**. O comando é **`dtvlt`**, que aparece só no CLI.
- **Licença:** MIT (ver [`LICENSE`](../LICENSE)).

## Fluxo de uso

```
$ dtvlt
  [arte ASCII]
  pin: ****                         ← só se o PIN estiver ativado
  session 7f3a-91c2
  http://127.0.0.1:48213            ← abre o navegador sozinho

navegador
  [arte ASCII gótica]
  _                                 ← 1º: palavra-chave
  _                                 ← 2º: senha
  → vault aberto
```

1. O usuário roda `dtvlt` e vê a arte ASCII.
2. Se o PIN estiver ativado, o CLI pede o PIN.
3. O CLI sobe o servidor local, mostra o Session ID e abre o navegador.
4. O frontend mostra a tela de login: primeiro a palavra-chave, depois a senha.
5. Com as duas certas, o vault é descriptografado e o frontend é liberado.

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
- O CLI abre o navegador com um **token de uso único** (validade de 60 s), trocado
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

Todos os itens têm tags e campos personalizados.

## Estética

O clima é de ferramenta de terminal séria, **sem "hacker cringe"**.

- **Sim:** monocromático (branco e cinza sobre preto), uma cor de destaque usada com
  muita parcimônia, fonte monoespaçada de qualidade, linguagem seca e técnica no estilo
  de CLIs modernos (`gh`, `fly`, ferramentas da Charm), animações lentas e sutis.
- **Não:** verde Matrix, "ACCESS GRANTED", glitch, caveiras, chuva de caracteres,
  mensagens dramáticas.

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
