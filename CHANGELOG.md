# Changelog

Todas as mudanças relevantes do DataVault ficam aqui.

O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/)
e o projeto usa [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Unreleased]

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

[Unreleased]: https://github.com/fbreseghello/DataVault/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/fbreseghello/DataVault/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/fbreseghello/DataVault/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/fbreseghello/DataVault/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/fbreseghello/DataVault/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/fbreseghello/DataVault/releases/tag/v0.1.0
