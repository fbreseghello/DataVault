# Tema: arcane

> Este documento dita o estilo do `arcane`: o clima, a paleta, os nomes, os ícones, a
> arte e os companheiros do hollow. O que vale para todos os temas está em
> [ESTILO.md](../ESTILO.md), o formato do arquivo em [TEMAS.md](../TEMAS.md) e o hollow
> em [PERSONAGEM.md](../PERSONAGEM.md).

## O clima

**Uma torre de mago à noite: grimórios, cristal e proteções.** Tudo acontece depois que
o mundo dorme, à luz fria da lua e de um cristal. O cofre é uma biblioteca selada por
encantamentos; usar o DataVault é estudar, conjurar e guardar.

- **Sim:** mistério calmo, estrelas, cristal, tinta e pergaminho, magia como disciplina.
- **Não:** explosões, raios coloridos, purpurina, unicórnios, "magia" de desenho animado.

Palavras que o tema usa: *ward*, *conjure*, *circle*, *bound*, *silenced*, *enshrine*.

## A paleta

Violeta escuro quase preto e cinzas lavanda, com um ciano de cristal.

| Papel | Cor | Uso |
|---|---|---|
| `bg` | `#0b0a12` | fundo |
| `raise` | `#12101c` | o que flutua |
| `line` | `#1f1b2e` | bordas e divisórias |
| `dim` | `#544d6e` | rótulos, dicas |
| `muted` | `#8b85a6` | botões, valores secundários |
| `fg` | `#e3dff2` | texto principal |
| `accent` | `#7fc8e0` | ciano de cristal: só mudança de estado |
| `red`, `blue`, `green` | `#e0707f`, `#8a9cf0`, `#7fd6a8` | só as tags do siege |

Os cinzas puxam para o violeta, como papel à luz da lua; o ciano é a única luz viva.

## Os nomes

| Página | Nome |
|---|---|
| grimoire | spellbook |
| unspoken | whispers |
| siege | rituals |
| crypt | reliquary |
| ledger | tithes |

As colunas do siege são `Scrolls`, `Casting`, `Bound`, `Fizzled`; o gerador de senhas
é `conjure`; selar é `ward`; guardar um arquivo é `enshrine`; o botão de encerrar é
`vanish`. O tom é de estudo e ritual: verbos de magia contida, nunca de batalha.

## Os ícones

**Contornos vazados**, finos e simétricos, como desenhos num caderno de alquimia: o
livro aberto sob uma estrela, a bola de cristal no pedestal, a varinha com a estrela,
o relicário, o cálice e a coruja do hollow. Mais ar dentro do quadro do que no souls,
e quase tudo espelhado.

## A arte

- **Login:** um "DataVault" em blocos pequenos e pontilhados (`▄`, `▀`, `▐`, `█`,
  `·`), leve e cheio de furos, como letras gravadas em cristal.
- **Tela de descanso:** a `candle` combina; as artes de pedra do souls (`pillar`,
  `skull-i`, `skull-ii`) funcionam, mas pesam mais do que o tema pede.

## O hollow

| Chave | Nome no arcane |
|---|---|
| `hollow` | familiar |
| `xp` | mana |
| `level` | circle |
| `deeds` | grimoire notes |
| atributos | will, lore, warding, relics, tithe |

**Títulos**, de 5 em 5 círculos: *apprentice*, *adept*, *magus*, *archmage*, *the
unbound*.

**O verso das cartas** é a coruja do tema em treliça.

### A categoria própria: fae

As **fadas**, mas as antigas: criaturas do bosque à noite, belas e um pouco
inquietantes, nunca fofas. Pequenas, de asas finas, com olhos de quem sabe mais do que
diz. Desenhadas no mesmo estilo de tarot dos outros companheiros, com o luar como
única luz.

| Chave | Quem é |
|---|---|
| `pixie` | pequenina, asas de libélula, cabelo desgrenhado, sentada num cogumelo |
| `wisp` | uma luz flutuante sobre o pântano, com um rosto que mal se vê dentro dela |
| `dryad` | uma mulher de pele de casca de árvore, galhos no lugar do cabelo, raízes nos pés |
| `moth` | uma fada de asas de mariposa, antenas finas, olhar parado, atraída por uma vela |
| `changeling` | parece uma criança, mas tem orelhas pontudas e olhos velhos demais |

Os prompts estão em [HOLLOW-ARTE.md](../HOLLOW-ARTE.md#fae).

### Os companheiros

Ordem das categorias na escolha (`companions`): **fae, wanderers, animals**, e depois
as outras. Todos os companheiros continuam disponíveis.

Os que mais combinam, fora da categoria própria:

| Companheiro | Por quê |
|---|---|
| `sorcerer` | o dono da torre |
| `witch` | a magia de ervas e velas, do lado de fora da torre |
| `owl` | o familiar clássico do mago, e o próprio ícone do tema |
| `cat` | o outro familiar clássico |
| `masked` | o espírito que anota tudo num caderno, como um estudioso do além |

## Ao criar algo novo para o arcane

- Pergunte se aquilo caberia numa estante de alquimista. Se faz barulho ou explode, não
  é do arcane.
- Simetria, círculos, estrelas de muitas pontas, cristal, pergaminho.
- Luz fria e de lua; o ciano só quando algo muda.
