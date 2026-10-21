# DataVault

Cofre local e de código aberto para segredos de quem vive no terminal:
logins, chaves de API, chaves SSH, seeds e notas seguras.

- Os dados ficam só na sua máquina, num único arquivo criptografado
  (Argon2id + XChaCha20-Poly1305). Sem nuvem, sem conta, sem telemetria.
- Você inicia pelo terminal com `dtvlt`, e a interface abre no navegador em `127.0.0.1`.
- Windows e Linux.

A visão completa e as decisões de design estão em [docs/VISAO.md](docs/VISAO.md).

> Projeto em estágio inicial (0.x). O formato do arquivo ainda pode mudar.

## Instalar

### Windows

1. Baixe `dtvlt-<versão>-windows-amd64.exe` na página de
   [Releases](https://github.com/fbreseghello/DataVault/releases/latest)
   (use `arm64` em PCs ARM).
2. Dê dois cliques no arquivo e responda `y` para instalar.

Ele se copia para `%LOCALAPPDATA%\Programs\DataVault` e entra no seu PATH, sem
pedir permissão de administrador. Depois, é só digitar `dtvlt` em qualquer terminal.

Para atualizar, baixe a versão nova e abra do mesmo jeito: ela oferece substituir a instalada.

> O executável ainda não é assinado digitalmente, então o Windows SmartScreen
> pode mostrar um aviso. Clique em "Mais informações" → "Executar assim mesmo".
> Para conferir o download, compare o hash com o `SHA256SUMS` da release.

### Linux

```bash
tar -xzf dtvlt-<versão>-linux-amd64.tar.gz
./dtvlt install
```

Ele se instala em `~/.local/bin`, que já está no PATH da maioria das distribuições.

## Usar

```
dtvlt            inicia a sessão e abre o DataVault no navegador
dtvlt status     mostra a sessão em execução
dtvlt lock       trava o vault e mantém a sessão
dtvlt stop       encerra a sessão
dtvlt config     altera as preferências
dtvlt uninstall  remove o dtvlt (o vault é mantido)
dtvlt version    mostra a versão
```

Na primeira execução, o `dtvlt` pergunta as preferências e cria o vault.
O vault é protegido por uma **palavra-chave** e uma **senha**: as duas são
necessárias e **não existe recuperação**.

O PIN opcional no terminal só impede que alguém inicie o `dtvlt` casualmente.
Ele não criptografa nada.

### Variáveis de ambiente

| Variável | Efeito |
|---|---|
| `DTVLT_HOME` | Guarda config, vault e estado da sessão nesse diretório (instalação portátil, testes). |
| `DTVLT_INSTALL_DIR` | Instala nesse diretório. Nesse caso o PATH não é alterado. |
| `DTVLT_NO_BROWSER=1` | Mostra o link de acesso em vez de abrir o navegador (SSH, outro navegador). |
| `NO_COLOR` | Desliga as cores do terminal. |

## Desenvolvimento

Requer Go 1.26+ e Node 20+. Um comando compila tudo (frontend + binário):

```bash
go run ./scripts/build
```

O resultado fica em `bin/dtvlt.exe` (Windows) ou `bin/dtvlt` (Linux). Builds locais
têm a versão `dev` e não oferecem instalação.

Testes:

```bash
go test ./...
npm --prefix web run check
```

## Versões

O DataVault segue o [Versionamento Semântico](https://semver.org/lang/pt-BR/) (`MAJOR.MINOR.PATCH`):

- **PATCH** (0.1.0 → 0.1.1): correções que não mudam o comportamento.
- **MINOR** (0.1.0 → 0.2.0): funcionalidades novas e compatíveis.
- **MAJOR** (1.x → 2.0): mudanças incompatíveis, como um formato de vault que
  versões antigas não conseguem abrir.

Enquanto estiver em `0.x`, uma versão MINOR também pode trazer mudanças incompatíveis.
O histórico fica no [CHANGELOG.md](CHANGELOG.md).

### Publicar uma versão

1. Mova as notas de `[Unreleased]` para uma seção nova no `CHANGELOG.md`, por exemplo `## [0.2.0] - 2026-11-05`.
2. Faça o commit e crie a tag: `git tag v0.2.0`.
3. Envie a tag: `git push origin v0.2.0`.

O GitHub Actions testa, compila os binários de Windows e Linux e publica a release
com as notas do changelog. Tags com sufixo (`v0.2.0-rc.1`) saem como pré-release.

## Licença

[MIT](LICENSE)
