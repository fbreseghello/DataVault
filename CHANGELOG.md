# Changelog

Todas as mudanças relevantes do DataVault ficam aqui.

O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/)
e o projeto usa [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Alterado

- O console não mostra mais a dica "type ,,, to open DataVault in the browser" ao abrir.
  Os comandos continuam no `help`.

## [0.7.0] - 2026-10-22

### Adicionado

- **Console do `dtvlt`.** Digitar `dtvlt` mostra a arte, a versão e o aviso de
  atualização, pede o PIN (se ativado) e abre um console que espera comandos:
  `,,,` (abre o DataVault no navegador), `status`, `lock`, `update`, `config`,
  `version`, `clear`, `stop` e `exit`. Os eventos da sessão aparecem acima do prompt.

### Alterado

- O navegador não abre mais sozinho ao digitar `dtvlt`: só com `,,,` no console.
- O aviso de atualização é conferido a cada início (não mais uma vez por dia) e
  pede para digitar `update`. Sem rede, vale a última resposta guardada.
- Conectar o console a uma sessão já aberta também pede o PIN.

### Corrigido

- `dtvlt update` falhava com "Access is denied" no Windows quando uma sessão de
  uma versão ainda mais antiga continuava aberta, porque ela segurava o arquivo
  `dtvlt.exe.old`. Agora o binário atual vai para um nome livre (`dtvlt.exe.old-N`)
  e todas essas cópias são apagadas quando ficam livres.

## [0.6.1] - 2026-10-22

### Corrigido

- Depois de atualizar, o login ainda mostrava a versão antiga: a sessão que já
  estava aberta continuava rodando o binário anterior. Agora o `dtvlt` percebe
  quando a sessão aberta é de outra versão, encerra essa sessão e abre uma nova
  na versão atual.

### Alterado

- `dtvlt status` mostra a versão da sessão em execução.

## [0.6.0] - 2026-10-22

### Adicionado

- Comandos de servidor no Grimoire: um novo tipo de entrada com o comando (várias
  linhas), o servidor e o diretório. O comando aparece como num terminal, com um `$`
  por linha, e o `copy` copia só o comando. O DataVault guarda e copia, nunca executa.

### Alterado

- O neon da arte do login agora é branco: letras brancas com um brilho fino e um
  halo suave que respira devagar. Com "reduzir movimento" no sistema, o halo fica parado.
- A versão instalada aparece no rodapé do DataVault, em todas as telas.

### Corrigido

- Nomes de campo longos (como `POSTGRES_PASSWORD`) passavam por cima do valor.
  A coluna do nome agora se ajusta ao nome mais longo da credencial, e nomes
  muito longos quebram linha.
- A lista do Grimoire recarrega ao voltar para ele.

## [0.5.1] - 2026-10-21

### Alterado

- O terminal mostra a versão instalada logo abaixo da arte do `dtvlt`.

## [0.5.0] - 2026-10-21

### Adicionado

- Credenciais seladas: abrir uma credencial selada pede o **PIN de selo**, um PIN
  único guardado (como hash) dentro do vault. Sem ele, o servidor não entrega os
  campos nem deixa editar ou apagar. Ela volta a ficar selada ao selecionar outra,
  trocar de tela ou travar o vault; trocar de aba do navegador não fecha.
- Cadeado no detalhe da credencial para selar (ou selar de novo) na hora, e chave
  "sealed" no editor.
- Troca do PIN de selo na tela settings, exigindo o PIN atual.

### Alterado

- Entrar no Grimoire sempre mostra a tela de abertura, sem reabrir a última
  credencial nem o editor.

## [0.4.0] - 2026-10-21

### Adicionado

- `dtvlt update`: baixa a release mais nova do GitHub, confere o SHA256 com o
  `SHA256SUMS` da release e instala.
- Aviso de atualização antes do PIN. O `dtvlt` consulta o GitHub no máximo uma vez
  por dia, com limite de 1,5 s. A opção "check for updates" desliga a consulta
  (vem ligada).

## [0.3.0] - 2026-10-21

### Adicionado

- Barra lateral só com ícones em pixel art de 1 bit, centralizados. As teclas 1–3 trocam de página.
- **Grimoire** (livro): a tela onde ficam logins, chaves e segredos.
- Duas páginas seladas (coração partido e espada), reservadas para funcionalidades futuras.
- Efeito neon leve na arte do login: ela liga piscando e fica com um halo suave.

### Corrigido

- Letras desalinhadas na arte gótica: "u" e "t" (e o "v" do banner do terminal) ficavam
  uma linha acima das outras. A arte agora é montada letra por letra, na mesma linha de base.
- Um build local com alterações não commitadas não herda mais a versão da tag; ele sai como `dev`.

## [0.2.0] - 2026-10-21

### Adicionado

- Tela de configurações no DataVault: auto-lock, limpeza da área de transferência,
  PIN no início e manter terminal aberto. Auto-lock e área de transferência valem na hora.

### Alterado

- O `.exe` baixado agora só instala (ou atualiza) e fecha. A sessão começa ao
  digitar `dtvlt` no terminal. Abrir uma versão mais antiga que a instalada pede confirmação.

## [0.1.0] - 2026-10-21

Primeira versão.

### Adicionado

- `dtvlt`: inicia a sessão, abre o navegador e mostra o Session ID.
- Configuração inicial no terminal: PIN opcional, manter terminal aberto,
  auto-lock e limpeza da área de transferência.
- Vault criptografado em um único arquivo: Argon2id (palavra-chave + senha)
  e XChaCha20-Poly1305, com backup da versão anterior.
- Servidor local em `127.0.0.1` com link de acesso de uso único, cookie de sessão
  e checagem de Host e Origin.
- Tela de login com a arte gótica, um campo por vez.
- Itens: login, chave de API, chave SSH, seed e nota segura, com tags, busca,
  mostrar/ocultar e copiar.
- Sessão em primeiro plano ou em segundo plano: `dtvlt status`, `lock`, `stop`, `config`.
- `dtvlt install` e `uninstall`. O `.exe` baixado oferece instalar ou atualizar
  sozinho ao ser aberto.

[Unreleased]: https://github.com/fbreseghello/DataVault/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/fbreseghello/DataVault/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/fbreseghello/DataVault/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/fbreseghello/DataVault/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/fbreseghello/DataVault/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/fbreseghello/DataVault/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/fbreseghello/DataVault/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/fbreseghello/DataVault/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/fbreseghello/DataVault/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/fbreseghello/DataVault/releases/tag/v0.1.0
