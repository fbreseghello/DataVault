# DataVault

Cofre local e de código aberto para segredos de quem vive no terminal:
logins, chaves de API, chaves SSH, seeds e notas seguras.

- Os dados ficam só na sua máquina, num único arquivo criptografado
  (Argon2id + XChaCha20-Poly1305). Sem nuvem, sem conta, sem telemetria.
- Você inicia pelo terminal com `dtvlt`, e a interface abre no navegador em `127.0.0.1`.
- Windows e Linux.

A visão completa e as decisões de design estão em [docs/VISAO.md](docs/VISAO.md).

> Projeto em estágio inicial. O formato do arquivo ainda pode mudar.

## Compilar

Requer Go 1.26+ e Node 20+.

```bash
npm --prefix web ci
npm --prefix web run build
go build -o bin/dtvlt ./cmd/dtvlt
```

No Windows, use `bin/dtvlt.exe` como saída.

## Usar

```
dtvlt            inicia a sessão e abre o DataVault no navegador
dtvlt status     mostra a sessão em execução
dtvlt lock       trava o vault e mantém a sessão
dtvlt stop       encerra a sessão
dtvlt config     altera as preferências
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
| `DTVLT_NO_BROWSER=1` | Mostra o link de acesso em vez de abrir o navegador (SSH, outro navegador). |
| `NO_COLOR` | Desliga as cores do terminal. |

## Desenvolvimento

```bash
go test ./...
npm --prefix web run check
```

## Licença

[MIT](LICENSE)
