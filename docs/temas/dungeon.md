# Tema: dungeon

> Este documento dita o estilo do `dungeon`: o clima, a paleta, os nomes, os ícones, a
> arte e os companheiros do hollow. O que vale para todos os temas está em
> [ESTILO.md](../ESTILO.md), o formato do arquivo em [TEMAS.md](../TEMAS.md) e o hollow
> em [PERSONAGEM.md](../PERSONAGEM.md).

## O clima

**Uma exploração de masmorra em 8 bits, à luz de tocha: chaves, baús e marcas de
missão.** O cofre é uma aventura de RPG antigo: cada senha é uma chave do chaveiro, cada
arquivo vai para o baú, cada tarefa é uma missão no diário. Há um pouco de humor de
jogo antigo, mas sem piada.

- **Sim:** tocha, pedra, madeira, ferro, baús, mapas, a sensação de "mais uma sala".
- **Não:** sci-fi, ficção espacial, cores de fliperama, "8 bits" como enfeite colorido.

Palavras que o tema usa: *quest*, *loot*, *stash*, *cleared*, *save & quit*.

## A paleta

Marrom-quase-preto e cinzas de pedra quentes, com o laranja de uma tocha.

| Papel | Cor | Uso |
|---|---|---|
| `bg` | `#0d0b09` | fundo |
| `raise` | `#15120f` | o que flutua |
| `line` | `#241f1a` | bordas e divisórias |
| `dim` | `#5c5349` | rótulos, dicas |
| `muted` | `#968b7b` | botões, valores secundários |
| `fg` | `#ebe3d5` | texto principal |
| `accent` | `#e07b39` | laranja de tocha: só mudança de estado |
| `red`, `blue`, `green` | `#d14b3c`, `#5f9fd6`, `#8fbf4f` | só as tags do siege |

Os cinzas puxam para o ocre, como pedra sob a tocha; o laranja é o fogo.

## Os nomes

| Página | Nome |
|---|---|
| grimoire | keyring |
| unspoken | journal |
| siege | quest log |
| crypt | chest |
| ledger | upkeep |

As colunas do siege são `Rumors`, `On the Road`, `Cleared`, `Lost`; o gerador de
senhas é `craft`; selar é `chain`; guardar um arquivo é `stash`; o botão de encerrar é
`save & quit`. O tom é de jogo de RPG antigo: curto, direto, com cara de menu.

## Os ícones

**Contornos vazados de 8 bits**, grossos e quadrados, como um jogo de cartucho: a
chave, a pena, o ponto de exclamação de missão, o baú, o livro de contas e o escudo do
hollow. Cantos retos, poucos detalhes, leitura imediata.

## A arte

- **Login:** um "DataVault" em blocos com sombra de linhas (`█`, `╗`, `╚`), como a
  tela de título de um jogo antigo.
- **Tela de descanso:** a `candle` e o `pillar` combinam com a masmorra; as caveiras
  também, como ossos no chão de uma sala.

## O hollow

| Chave | Nome no dungeon |
|---|---|
| `hollow` | hero |
| `xp` | xp |
| `level` | level |
| `deeds` | quest journal |
| atributos | strength, wisdom, defense, loot, gold |

**Títulos**, de 5 em 5 níveis: *peasant*, *squire*, *knight*, *champion*, *legend*.

**O verso das cartas** é o escudo do tema em treliça.

### A categoria própria: denizens

Os **habitantes da masmorra**, os monstros clássicos que todo aventureiro já
encontrou, agora do seu lado. Desenhados no mesmo estilo de tarot dos outros
companheiros, à luz de tocha, entre pedra e madeira.

| Chave | Quem é |
|---|---|
| `goblin` | pequeno e magro, orelhas grandes, uma adaga e uma lanterna roubada |
| `skeleton` | um esqueleto de pé, escudo rachado e espada enferrujada, ainda de guarda |
| `slime` | uma massa gelatinosa, com moedas e uma chave presas dentro |
| `mimic` | um baú de tesouro com dentes na tampa e uma língua longa |
| `golem` | um gigante de pedra, rachado, com runas apagadas no peito |

Os prompts estão em [HOLLOW-ARTE.md](../HOLLOW-ARTE.md#denizens).

### Os companheiros

Ordem das categorias na escolha (`companions`): **denizens, bushi, demons**, e depois
as outras. Todos os companheiros continuam disponíveis.

Os que mais combinam, fora da categoria própria:

| Companheiro | Por quê |
|---|---|
| `mercenary` | o aventureiro de aluguel, de espada grande |
| `hellhound` | o guardião da última sala |
| `imp` | o diabrete que mora entre os baús |
| `bat` | o morcego de todo corredor de masmorra |
| `samurai` | o guerreiro de armadura completa, como um herói de RPG |

## Ao criar algo novo para o dungeon

- Pergunte se aquilo poderia estar num baú ou numa sala de masmorra. Se é tecnológico
  ou limpo demais, não é do dungeon.
- Objetos de aventura: chaves, baús, mapas, tochas, armas simples, moedas.
- Cara de menu de jogo antigo nos nomes; o laranja só quando algo muda.
