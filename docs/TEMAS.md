# DataVault — Temas

> Como criar e importar um tema. A forma geral (tipografia, layout, tom) está em
> [ESTILO.md](ESTILO.md) e vale para todos os temas.

Um tema muda quatro coisas, sempre dentro do clima RPG de 1 bit:

- **as cores** do sistema;
- **os nomes** das páginas e das funcionalidades;
- **os ícones**, em pixel art 16×16 de 1 bit;
- **a arte do login**, o "DataVault" em ASCII acima do prompt.

## Temas que já vêm no DataVault

| Tema | Clima | Páginas |
|---|---|---|
| `souls` (padrão) | ruínas, lápides e livros antigos | grimoire, unspoken, siege, crypt, ledger |
| `arcane` | torre de mago à noite | spellbook, whispers, rituals, reliquary, tithes |
| `dungeon` | dungeon 8-bit à luz de tocha | keyring, journal, quest log, chest, upkeep |

Os arquivos ficam em [`internal/theme/builtin/`](../internal/theme/builtin/) e vão
embutidos no `dtvlt`. Escolha um em **settings → appearance → theme**: a tela muda na
hora, e **save** guarda a escolha (`theme` no `config.toml`). Sair sem salvar volta ao
tema salvo.

## Importar um tema

Um tema é um arquivo `.json`. Em **settings → appearance**:

- **export** salva o tema escolhido como `.json`, com todas as cores, nomes, ícones e a
  arte preenchidos. É o jeito mais fácil de começar: exporte, edite, importe.
- **import** lê um `.json`, valida e guarda o tema em `themes/` ao lado do
  `config.toml`. Importar de novo um tema com o mesmo `id` substitui o anterior.
- **remove** apaga um tema importado. Se era o tema em uso, o DataVault volta ao `souls`.

Só dá para importar com o cofre aberto. Um tema não tem segredos, então a tela de login
também o usa.

## O formato

Tudo é opcional, exceto `id` e `name`. O que o tema deixar de fora vem do `souls`,
então um tema pode trocar só as cores, ou só alguns nomes.

```json
{
  "id": "ember",
  "name": "ember",
  "description": "brasas de uma forja apagada",
  "colors": { "accent": "#ff6a3d", "bg": "#120a08" },
  "words": { "grimoire": "keys", "siege": "raids" },
  "icons": { "seal": ["................", "...16 linhas de 16 caracteres..."] },
  "large_icons": [],
  "login": ["linha 1 da arte", "linha 2 da arte"]
}
```

| Campo | Regras |
|---|---|
| `id` | 1 a 32 letras minúsculas, dígitos ou `-`. Não pode ser o de um tema embutido. |
| `name` | 1 a 32 caracteres; é o que aparece na lista. |
| `description` | até 160 caracteres; aparece ao lado da lista. |
| `colors` | cores em `#rgb` ou `#rrggbb` (veja abaixo). |
| `words` | nomes de uma linha, até 24 caracteres (veja abaixo). |
| `icons` | 16 linhas de 16 caracteres: `#` aceso, `.` apagado. |
| `large_icons` | ícones desenhados um passo maior na barra lateral, para arte que deixa mais vazio no quadro. |
| `login` | até 24 linhas de até 160 caracteres. A arte encolhe para caber em janelas estreitas. |

Chaves desconhecidas são recusadas, para que um erro de digitação não passe calado.

### Cores

| Chave | Uso |
|---|---|
| `bg` | fundo de tudo |
| `raise` | o que flutua: dicas, opções de dropdown |
| `line` | bordas, divisórias, sublinhado dos campos |
| `dim` | rótulos, dicas, rodapé, ícones em repouso |
| `muted` | botões de texto, valores secundários |
| `fg` | texto principal e item ativo |
| `accent` | **só mudança de estado** |

O DataVault é escuro: mantenha `bg` perto do preto e `fg` claro.

### Nomes

| Chave | Onde aparece | souls |
|---|---|---|
| `grimoire`, `unspoken`, `siege`, `crypt`, `ledger` | nome de cada página | os mesmos |
| `forge` | o gerador de senhas | forge |
| `seal`, `sealed` | proteger uma entrada com o PIN; "seal pin" | seal, sealed |
| `dead_words` | onde ficam as notas enterradas | dead words |
| `bury`, `unbury` | mandar uma nota para lá e trazer de volta | bury, unbury |
| `orders`, `in_battle`, `conquered`, `fallen` | as quatro colunas do siege | Orders, In Battle, Conquered, Fallen |
| `keep`, `kept` | guardar um arquivo no crypt, e o rótulo da data | bury, buried |
| `strike` | apagar uma conta do ledger | strike |
| `end` | o botão que encerra a sessão | end session |

### Ícones

| Chave | Onde aparece |
|---|---|
| `grimoire`, `unspoken`, `siege`, `crypt`, `ledger` | barra lateral e a placa de cada página |
| `end` | o botão de encerrar a sessão e a tela de sessão encerrada |
| `seal` | entradas e notas protegidas pelo PIN |

Desenhe como os do `souls`: silhueta legível em 32 px, vãos de 2 px, sem antisserrilhado.

### Arte do login

Uma fonte do [figlet](https://patorjk.com/software/taag/) funciona bem. O `souls` usa
**Poison**, o `arcane` usa **Elite** e o `dungeon` usa **ANSI Shadow**. Tire os espaços
do fim das linhas e as linhas vazias de cima e de baixo.
