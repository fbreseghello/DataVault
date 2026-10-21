# Changelog

Todas as mudanças relevantes do DataVault ficam aqui.

O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/)
e o projeto usa [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Adicionado

- Barra lateral só com ícones em pixel art de 1 bit, centralizados. As teclas 1–3 trocam de página.
- **Grimoire** (livro): a tela onde ficam logins, chaves e segredos.
- Duas páginas seladas (coração partido e espada), reservadas para funcionalidades futuras.

### Corrigido

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

[Unreleased]: https://github.com/fbreseghello/DataVault/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/fbreseghello/DataVault/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/fbreseghello/DataVault/releases/tag/v0.1.0
