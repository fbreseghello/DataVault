# DataVault — Personagem

> **Em construção.** Este documento descreve um companheiro que evolui com o uso do
> DataVault. Já existem o ícone no topo da barra lateral e a experiência vinda do Siege,
> com níveis, títulos, atributos e o histórico, e a carta do companheiro na página, com
> uma imagem provisória até as ilustrações chegarem (veja a
> [ordem de construção](#ordem-de-construção)). A versão 0.22 trouxe classes em pixel art;
> elas saem quando os companheiros descritos aqui entrarem. A forma geral segue o
> [guia de estilo](ESTILO.md) e o formato de temas segue [TEMAS.md](TEMAS.md). Os prompts
> das imagens estão em [HOLLOW-ARTE.md](HOLLOW-ARTE.md).

## A ideia

Um **companheiro** que acompanha quem usa o cofre, no espírito de um tamagotchi, mas com
a arte e o tom do DataVault. Quem usa escolhe **o que** ele é, numa das
categorias: uma figura de dark fantasy, um animal, um shinigami, um demônio, um
guerreiro do Japão feudal, uma fada, um habitante de masmorra. Ele
ganha **experiência (XP)** com o que acontece no sistema (concluir tarefas no Siege,
escrever no Unspoken, renovar uma chave no prazo) e sobe de **nível**, com títulos ao
longo do caminho. Também aparece fora da página: na **tela de descanso** e no terminal,
com o `dtvlt hollow`.

A arte é **ilustração no estilo do tarot Rider-Waite-Smith** (1909), como a carta do
Eremita: traço de nanquim, cor chapada, uma figura solitária num cenário quase vazio.
Nas telas, cada imagem aparece nos tons do tema.

Não é um jogo que pede atenção, e o companheiro não morre nem adoece por abandono, ao
contrário do tamagotchi. É um diário de campanha silencioso: ele fica parado, como as
outras artes do DataVault, e só registra o caminho percorrido.

- **Sim:** progresso lento e honesto, texto seco, arte parada, tudo local e criptografado.
- **Não:** "LEVEL UP!!!", confete, sons, notificações insistentes, ranking, metas que
  culpam quem não abriu o app, recompensa por ações sem sentido, fome, morte, sujeira.

O nome interno da funcionalidade é **`hollow`** (como as páginas, a chave usa o nome
do `souls`). Cada tema dá o seu nome a ela.

## Na barra lateral

Um ícone novo, **sozinho no topo da barra**, espelhando a lápide que fica sozinha no pé:

```
┌────┐
│ ▣  │  ← hollow, a 16 px do topo
│    │
│ ▤  │  ┐
│ ▥  │  │
│ ▦  │  ├ páginas, centralizadas como hoje
│ ▧  │  │
│ ▨  │  ┘
│    │
│ ▩  │  ← lápide, a 16 px do fundo
└────┘
```

- Mesmo botão `.page` das outras (64 × 48 px, ícone de 32 px), com `position: absolute;
  top: 16px`, como o `.end` usa `bottom: 16px`. As páginas continuam centralizadas.
- Mesmos estados: `--dim` em repouso, `--muted` no hover, `--fg` e a barra de 2 px
  quando a página está aberta.
- **Dica flutuante:** o nome do tema e o nível, `hollow · lv 7 <0>`.
- **Atalho:** `0` (as páginas usam `1` a `5`).
- **Cor de destaque:** só quando há um nível novo ainda não visto. O ícone fica em
  `--accent` até a página ser aberta, e volta ao normal. É uma mudança de estado, então
  respeita a regra do âmbar.
- Se o personagem estiver desligado em settings, o ícone não aparece.

## Ícones por tema

Cada tema traz o seu ícone na chave `hollow` de `icons` (16 × 16, 1 bit, `#` aceso).
Os três já estão nos temas embutidos e foram testados em 32 px ao lado dos vizinhos,
como pede o [guia](ESTILO.md#ícones).

### souls: capuz vazio

Um capuz sem ninguém dentro: o que sobra de quem se tornou hollow. É só a borda do
capuz sobre os ombros, sem rosto. (Uma caveira de órbitas vazias dentro do capuz veio
antes e, em 32 px, tinha informação demais; um elmo fechado, antes dela, ficou pesado.)

```
................
......####......
....########....
...###....###...
..###......###..
..##........##..
.###........###.
.##..........##.
.##..........##.
.###........###.
..###......###..
..####....####..
.##############.
################
................
................
```

### arcane: coruja

O `familiar` do mago: uma coruja vazada, com tufos nas orelhas, pousada num galho. É
vazada como os outros ícones do `arcane`. (Um chapéu de mago foi tentado antes e, em
32 px, parecia um barco à vela.)

```
................
..##........##..
..#.#......#.#..
..#..######..#..
..#..........#..
..#..##..##..#..
..#..##..##..#..
..#..........#..
..#....##....#..
..#..........#..
...#........#...
....########....
.....#....#.....
..############..
................
................
```

### dungeon: escudo

Escudo de herói de 8 bits, vazado como os outros ícones do `dungeon`, com uma cruz no
meio.

```
................
.##############.
.#............#.
.#............#.
.#.....##.....#.
.#.....##.....#.
.#..########..#.
.#..########..#.
.#.....##.....#.
.#.....##.....#.
..#....##....#..
..#..........#..
...#........#...
....##....##....
......####......
................
```

Um tema importado que não defina `hollow` herda a caveira do `souls`, como acontece com
qualquer ícone que falte.

## Nomes por tema

Novas chaves em `words`, com as mesmas regras (uma linha, até 24 caracteres):

| Chave | Onde aparece | souls | arcane | dungeon |
|---|---|---|---|---|
| `hollow` | nome da página e da dica | hollow | familiar | hero |
| `xp` | a experiência | souls | mana | xp |
| `level` | o nível (abreviado `lv` na dica) | level | circle | level |
| `deeds` | o histórico de XP | deeds | grimoire notes | quest journal |

## A página: a carta

A página é **uma carta de tarot**. As ilustrações vêm sem moldura nem título (veja
[A arte](#a-arte)); a moldura, o número e o nome são desenhados pelo app, em linhas
finas nos tons do tema, e carregam o que o companheiro é. À esquerda a carta, à
direita a leitura.

```
datavault / hollow
────────────────────────────────────────────────────────────────────
                                   │
    ┌──────────── IX ───────────┐   │   IX · the hermit
    │                          │   │   the unkindled
    │                          ·   │
    │                          ·   │   souls    340 / 500
    │       a ilustração        ·   │
    │      nos tons do tema    ·   │   vigor    ████████████   58
    │                          ·   │   memory   ██████         31
    │                          ·   │   ward     ████           24
    │                          ·   │   hoard    █               6
    │                          ·   │   thrift   ██             12
    │                          ·   │
    └───── ◆ THE HERMIT ◆ ··········┘   │   today
                                   │   +10  task conquered   14:20
     yesterday · four orders       │   +25  day cleared      14:20
     conquered, the day cleared    │
────────────────────────────────────────────────────────────────────
 IX · 160 souls to next                   space turn · c companion
```

### A frente

- **O número é o nível**, em algarismos romanos no topo da moldura, como nas cartas:
  nível 9 é `IX`. Num nível novo, só o numeral fica em `--accent`, até a página ser
  aberta.
- **O nome** do companheiro vai no pé da moldura, em versalete: `THE HERMIT`.
- **A moldura é a barra de XP.** O contorno é traçado em `--fg`, no sentido horário a
  partir do numeral, conforme o XP do nível sobe; o que falta fica pontilhado em
  `--line`. Quando o contorno fecha, o numeral muda. Não há outra barra de XP na tela.
- **A moldura envelhece com os títulos.** A cada título novo (de 5 em 5 níveis) ela
  ganha um detalhe: marcas nos cantos, depois uma linha dupla, depois um ornamento em
  volta do numeral, e assim por diante. A carta de quem está há muito tempo com o
  companheiro parece mais gasta e mais rica, sem nenhuma ilustração nova.
- **Selos de epitáfio.** Cada [epitáfio](#epitáfios) conquistado vira um pequeno selo
  (◆) no pé da moldura, ao lado do nome. Passar o mouse mostra o texto; os que faltam
  não aparecem.
- A ilustração ocupa o miolo, no maior tamanho que cabe no painel sem distorcer, na
  proporção 2:3 das imagens.

### O verso

`espaço`, ou um clique na carta, vira a carta. O verso guarda o que é permanente:

- o desenho do verso é o **ícone `hollow` do tema repetido em treliça**, em 1 bit, nos
  tons do tema, então cada tema tem o seu verso de baralho;
- por cima, a lista dos epitáfios gravados, com a data, e `since 12 Mar 2026`, o dia em
  que o companheiro foi escolhido;
- quando existir, a crônica do ano (veja [Ritmo](#ritmo)).

A frente é o momento; o verso, a história. `espaço` de novo volta à frente.

### A leitura, à direita

Só texto, na hierarquia de cinzas do resto do app, sem caixas nem cards, como uma
tiragem anotada num caderno:

- o numeral, o nome e o título: `IX · the hermit` e, embaixo, `the unkindled`;
- o XP do nível em números (`340 / 500`), já que a barra é a moldura;
- os cinco atributos, com barras curtas relativas ao maior;
- o histórico (`deeds`) agrupado por dia, com o ganho, o motivo e a hora. Nunca mostra o
  título de notas, entradas ou tarefas, só o tipo de ação, como o log da sessão no
  terminal.

### A leitura do dia

Na primeira vez que a página abre num dia, uma linha embaixo da carta resume o dia
anterior no tom de uma leitura: `yesterday · four orders conquered, the day cleared`.
Some quando outra página é aberta. Só números e tipos de ação, nunca títulos; num dia
sem nada, não aparece.

### O que fica de fora

Ícones e botões além de **companion**; animação, a não ser a carta virando (que
respeita "reduzir movimento"); a cor de destaque fora do numeral; sombra ou brilho na
carta.

## Experiência

### O que dá XP

Só ações que representam trabalho de verdade. Valores iniciais, para ajustar com uso:

| Página | Ação | XP | Limite |
|---|---|---|---|
| Siege | tarefa movida para **Conquered** | 10 | 20 tarefas por dia |
| Siege | dia fechado sem nada em Orders nem In Battle | 25 | 1 por dia |
| Unspoken | nota escrita (pelo menos 140 caracteres novos) | 5 | 3 por dia |
| Grimoire | entrada renovada no prazo (**renewed** antes de vencer) | 15 | por entrada e renovação |
| Grimoire | códigos de recuperação novos depois do aviso de "2 ou menos" | 15 | por entrada |
| Crypt | arquivo guardado | 3 | 5 por dia |
| Ledger | conta cancelada (**strike**) | 5 | 5 por dia |
| — | voltar ao cofre depois de 3 dias ou mais fora (**rested**) | os próximos 50 XP valem dobrado | 1 por volta |

O que **não** dá XP: criar e apagar coisas em sequência, mover uma tarefa de volta e de
novo para Conquered, copiar senhas, abrir páginas, tempo com o app aberto.

**Sem sequência de dias.** Nada se perde por ficar longe; quem volta descansado ganha
mais por um tempo. A ausência nunca é punida.

### Desfazer

O XP é ligado à origem. Se a tarefa sai de Conquered, a nota é apagada ou a renovação é
desfeita, o XP daquela origem sai junto. Uma mesma origem nunca rende duas vezes.
Tarefas **Fallen** não dão nem tiram XP: desistir também é uma decisão.

### Níveis

XP para passar do nível `n` para o `n + 1`: `round(100 × n^1.4)`, arredondado para a
dezena. Sem nível máximo.

| Nível | XP para o próximo | XP total ao chegar |
|---|---|---|
| 1 | 100 | 0 |
| 2 | 260 | 100 |
| 3 | 470 | 360 |
| 4 | 700 | 830 |
| 5 | 950 | 1 530 |
| 10 | 2 510 | 9 240 |

Quem conclui umas cinco tarefas por dia chega ao nível 5 em cerca de um mês e ao
nível 10 em uns seis meses. É de propósito: lento o bastante para cada nível
significar algo.

### Títulos

A cada cinco níveis o personagem ganha um título, que aparece no painel e na barra da
página. Cada tema define os seus numa lista nova, `titles`, em ordem; o último se
repete daí em diante.

| Níveis | souls | arcane | dungeon |
|---|---|---|---|
| 1–4 | the nameless | apprentice | peasant |
| 5–9 | the unkindled | adept | squire |
| 10–14 | the wayfarer | magus | knight |
| 15–19 | the warden | archmage | champion |
| 20+ | the undying | the unbound | legend |

### Nível novo

Sem tela cheia nem animação. Quando o XP passa do limite:

- o ícone do topo fica em `--accent` até a página ser aberta;
- o rodapé mostra, uma vez, `level 8 · the unkindled` em `--accent`, que some como
  as outras mensagens de estado;
- no terminal aberto, o log da sessão registra `17:24  level 8`.

## O companheiro

Quem usa **escolhe o companheiro** numa das categorias abaixo. Cada tema tem a sua
**categoria própria**, desenhada com a cara dele (veja
[Os temas e as categorias](#os-temas-e-as-categorias)); as outras são de todos. A chave interna é fixa
(`hermit`, `raven`, ...); os nomes aparecem em inglês, como o resto da interface.

### Wanderers

Figuras de dark fantasy: andarilhos, guerreiros e magos que seguem quem usa o cofre.
A categoria própria do [souls](temas/souls.md).

| Chave | Quem é |
|---|---|
| `hermit` | o eremita: velho de manto e capuz, lanterna erguida, cajado |
| `mercenary` | guerreiro marcado de cicatrizes, espada enorme no ombro, capa rasgada |
| `sorcerer` | mago de manto longo, cajado entalhado e um grimório preso ao cinto |
| `witch` | bruxa de chapéu largo, ervas penduradas e uma vela acesa na mão |
| `necromancer` | encapuzado e magro, entre ossos, com uma lanterna de luz fria |

### Animals

Bichos de companhia.

| Chave | Quem é |
|---|---|
| `dog` | um cão magro e leal, sentado, de orelhas em pé |
| `cat` | um gato preto sentado, cauda enrolada nas patas |
| `raven` | um corvo pousado num galho seco |
| `bat` | um morcego pendurado de cabeça para baixo, asas fechadas |
| `owl` | uma coruja pousada, olhos grandes |
| `wolf` | um lobo de pé, pelo espesso, olhar atento |
| `serpent` | uma serpente enrolada, cabeça erguida |

### Shinigami

Ceifadores que ficam ao lado, pacientes.

| Chave | Quem é |
|---|---|
| `reaper` | alto e esquelético, manto preto, foice |
| `gaunt` | magro e anguloso, olhos grandes e redondos, sorriso torto, asas finas e esfarrapadas |
| `masked` | envolto em panos, com uma máscara de osso, um caderno na mão |

### Demons

Demônios domesticados, meio contrariados de estar ali.

| Chave | Quem é |
|---|---|
| `imp` | pequeno, chifres curtos, asas de morcego, rabo pontudo |
| `brute` | grande e curvado, chifres retorcidos, correntes nos pulsos |
| `fiend` | esguio, asas grandes, cauda longa, garras |
| `hellhound` | um cão de fogo, costelas à mostra, brasas no pelo |

### Bushi

Guerreiros do Japão feudal, desenhados no mesmo estilo de tarot dos outros.

| Chave | Quem é |
|---|---|
| `samurai` | armadura completa e kabuto, a katana na cintura, de pé e firme |
| `ronin` | o samurai sem senhor: chapéu de palha largo, quimono gasto, a espada no ombro |
| `shinobi` | roupa escura, rosto coberto, uma lâmina curta junto ao corpo |
| `komuso` | o monge do cesto: um chapéu de palha que cobre a cabeça inteira, uma flauta shakuhachi |
| `sohei` | o monge guerreiro: hábito, capuz branco, a naginata (lança de lâmina curva) |

### Fae

As fadas antigas, do bosque à noite: belas e um pouco inquietantes, nunca fofas. A
categoria própria do [arcane](temas/arcane.md).

| Chave | Quem é |
|---|---|
| `pixie` | pequenina, asas de libélula, cabelo desgrenhado, sentada num cogumelo |
| `wisp` | uma luz flutuante sobre o pântano, com um rosto que mal se vê dentro dela |
| `dryad` | pele de casca de árvore, galhos no lugar do cabelo, raízes nos pés |
| `moth` | asas de mariposa, antenas finas, olhar parado, atraída por uma vela |
| `changeling` | parece uma criança, mas tem orelhas pontudas e olhos velhos demais |

### Denizens

Os habitantes da masmorra: os monstros clássicos que todo aventureiro já encontrou,
agora do seu lado. A categoria própria do [dungeon](temas/dungeon.md).

| Chave | Quem é |
|---|---|
| `goblin` | pequeno e magro, orelhas grandes, uma adaga e uma lanterna roubada |
| `skeleton` | de pé, escudo rachado e espada enferrujada, ainda de guarda |
| `slime` | uma massa gelatinosa, com moedas e uma chave presas dentro |
| `mimic` | um baú de tesouro com dentes na tampa e uma língua longa |
| `golem` | um gigante de pedra rachado, com runas apagadas no peito |

### Escolher e trocar: tirar uma carta

- **A primeira vez** que a página abre, ela mostra **as cartas das categorias viradas
  para baixo**, na ordem do tema. O verso de cada uma traz o selo da categoria no meio,
  um ícone de 16 × 16 em 1 bit como os da barra: uma **lanterna** (wanderers), uma
  **pegada** (animals), uma **foice** (shinigami), **chifres** (demons), um **chapéu de
  palha** (bushi), uma **asa de mariposa** (fae) e uma **tocha** (denizens). O nome vai
  embaixo.
- Escolher uma categoria abre os companheiros dela **em leque**, de frente, cada um com
  o nome no pé e uma linha de descrição. Setas escolhem, `Enter` confirma e a carta vira
  para o centro; `Esc` volta às categorias.
- **Trocar depois** é livre e não custa XP nem nível: **companion** (tecla `c`) abre a
  mesma escolha. O nível, o XP, os atributos e o histórico continuam.
- As classes da versão 0.22 (`knight`, `scribe`, `warden`, `thief`, `cleric`) saem.
  Quem já tinha escolhido uma vê a escolha de companheiro na próxima vez que abrir a
  página. A afinidade de classe sai junto: os atributos passam a refletir só o uso.

### Os temas e as categorias

O estilo de cada tema, com os companheiros que fazem sentido nele, está num documento
próprio: [souls](temas/souls.md), [arcane](temas/arcane.md) e
[dungeon](temas/dungeon.md).

**Todo companheiro serve para qualquer tema.** Um demônio pode viver no `dungeon` e um
corvo no `arcane`: a escolha é de quem usa, nunca do tema. A carta, a moldura e a
ilustração já mudam de cor com o tema, então qualquer companheiro combina com qualquer
um deles.

O que o tema faz é **dar o tom da escolha**:

- **A ordem das cartas.** Cada tema pode pôr na frente as categorias que combinam com
  ele, numa lista nova e opcional do tema, `companions`, com as chaves das categorias em
  ordem (`["wanderers", "shinigami"]`). As que o tema não lista vêm depois, na ordem de
  sempre. Nada some e nada fica trancado.
- **O verso do baralho.** O verso da carta usa o ícone `hollow` do tema (veja
  [O verso](#o-verso)), então o mesmo companheiro, em temas diferentes, parece vir de
  baralhos diferentes.
- **A categoria própria.** Cada tema tem uma categoria desenhada com a cara dele:
  **wanderers** no `souls`, **fae** no `arcane`, **denizens** no `dungeon`. Ela abre a
  escolha no tema dela, mas fica aberta a todos. Um tema novo vem com a sua. Um tema importado pode ordenar as
  categorias que existem, mas não cria categorias: as ilustrações vêm com o DataVault.

| Tema | Põe na frente |
|---|---|
| `souls` | wanderers, shinigami, bushi |
| `arcane` | fae, wanderers, animals |
| `dungeon` | denizens, bushi, demons |

## Atributos

Cada página alimenta um atributo. O XP é um só, mas cada ganho também soma pontos no
atributo da página de onde veio, e o personagem acaba tendo o formato de como o cofre é
usado.

| Chave | souls | arcane | dungeon | Vem de |
|---|---|---|---|---|
| `vigor` | vigor | will | strength | Siege |
| `memory` | memory | lore | wisdom | Unspoken |
| `ward` | ward | warding | defense | Grimoire |
| `hoard` | hoard | relics | loot | Crypt |
| `thrift` | thrift | tithe | gold | Ledger |

- No painel, abaixo da barra de XP, cada atributo é uma linha com uma barra curta:
  `vigor   ######....  42`. A barra é relativa ao maior atributo.

## A arte

### O estilo

**Ilustração no estilo do tarot Rider-Waite-Smith** (1909), tendo a carta do Eremita
como referência: contorno firme de nanquim, áreas de cor chapada, pouca textura, uma
figura solitária, de corpo inteiro, em três quartos olhando para a esquerda, num
cenário quase vazio. **Só o estilo**: nada de moldura, título nem número de carta.

- As imagens são geradas no Midjourney com os prompts de
  [HOLLOW-ARTE.md](HOLLOW-ARTE.md), todas com a mesma referência de estilo
  (`--sref`), para os companheiros parecerem do mesmo baralho.
- Formato **2:3** (vertical), com a figura inteira e uma margem em volta.
- Guerreiros, shinigamis e demônios podem lembrar o clima de obras conhecidas (o
  guerreiro marcado de Berserk, os shinigamis de Death Note), mas **nunca** reproduzem um
  personagem: os prompts descrevem traços genéricos e nunca citam nomes de obras ou de
  personagens.

### Nos tons do tema

A imagem é guardada **em tons de cinza** e pintada na hora com as cores do tema: o
preto vira `--bg`, o branco vira `--fg`, e os cinzas do meio passam por `--dim` e
`--muted`. É um degradê contínuo, sem os degraus do pixel art, então o traço e a
sombra da ilustração se mantêm. Ela muda sozinha quando o tema muda.

- **A cor de destaque nunca entra na imagem.**
- O fundo da ilustração deve ser escuro e liso, para virar o `--bg` da página sem
  marca de borda.
- Os ícones da barra continuam em pixel art de 1 bit; a ilustração do companheiro é a
  única arte desse tipo no app, e o guia deve registrar isso quando ela entrar no
  código.

### Os arquivos

- Os originais do Midjourney ficam em `art/hollow/`, um PNG por companheiro, com a
  chave no nome (`hermit.png`, `raven.png`).
- Um script em `scripts/art/` converte cada um: recorta a margem, passa para tons de
  cinza, ajusta o contraste para o preto do fundo cair no `--bg`, reduz para no máximo
  1024 px de altura e grava em WebP. É esse arquivo que vai no binário.
- O tamanho de todos juntos fica abaixo de alguns MB, para o `dtvlt` não crescer demais.

## O cofre no humor: a carta invertida

No tarot, uma carta de cabeça para baixo é a leitura **invertida**: o mesmo arcano, com
o sentido torto. Quando o cofre tem pendências, a carta do companheiro aparece **girada
180°**, e embaixo dela surge uma linha só, em `--dim`: `reversed · 2 keys rotting`.
Resolvida a causa, ela volta à posição normal.

| Fica invertida quando | Volta quando |
|---|---|
| há entradas vencidas no Grimoire ou uma com 2 códigos de recuperação ou menos | tudo é renovado |
| há tarefas em In Battle há mais de 3 dias | elas saem de In Battle |

- Não tira XP nem nível, e nada piora com o tempo. É uma auditoria de relance, não uma
  punição.
- O numeral e o nome giram junto com a carta, como numa carta de verdade; a leitura à
  direita continua de pé.
- A dica do ícone na barra lateral diz a causa: `hollow · lv 7 · 2 keys rotting`.
- Uma ilustração só por companheiro basta: a carta invertida dispensa as versões
  "inquietas" que o plano chegou a prever.

## Epitáfios

Marcos que acontecem uma vez só e ficam gravados numa lista, como inscrições numa
lápide. Cada tema dá o seu nome à lista (`epitaphs` no `souls`, `chronicles` no
`arcane`, `achievements` no `dungeon`).

| Epitáfio | Quando |
|---|---|
| `first key renewed before it rotted` | a primeira renovação no prazo |
| `wrote what was never said` | a primeira nota no Unspoken |
| `a hundred orders conquered` | 100 tarefas em Conquered |
| `thirty days at the vault` | 30 dias diferentes com o cofre aberto (não seguidos) |
| `buried a year of dead words` | uma nota enterrada há mais de um ano |
| `cut the ledger by half` | o gasto mensal cai à metade do maior já registrado |

- Cada um dá XP uma vez.
- A lista mostra a data e o texto; os que faltam ficam ocultos, só com `???`, para não
  virarem uma lista de tarefas.
- Mesmo tom do resto: minúsculas, sem exclamação.

## Ritmo

- **A vela do dia:** ao lado do companheiro, uma vela pequena no mesmo estilo de
  ilustração e nos tons do tema, parada. Ela vai derretendo conforme as tarefas do dia
  saem de Orders e é trocada por uma nova à meia-noite. Não apaga por falta de uso.
- **A crônica do ano:** no primeiro acesso de cada ano, a página oferece a crônica do
  ano anterior, uma página de texto com o resumo: tarefas conquistadas, notas escritas,
  chaves renovadas, quanto se cortou no Ledger, níveis e epitáfios. Só números e tipos
  de ação, nunca títulos.

## Tela de descanso

Cada companheiro também é um **estilo da tela de descanso**, ao lado de `pillar`,
`skull-i`, `skull-ii` e `candle`. Assim, as ilustrações valem por si, mesmo para quem
não liga para o nível.

- **`hollow`** mostra **a carta inteira** do companheiro escolhido, com a moldura, o
  numeral e o nome, invertida se for o caso, no canto inferior direito, como as outras
  artes. Embaixo, em `--dim` e 11 px, só `the unkindled`.
- Cada ilustração também pode ser escolhida diretamente como estilo (`hermit`,
  `raven`...), sem nível nem texto, mesmo com o hollow desligado.
- Parada, nos tons do tema, no maior tamanho que cabe na altura pedida sem passar de
  metade da largura.
- O cofre travado não tem acesso ao companheiro escolhido: o estilo `hollow` só aparece
  enquanto a sessão está aberta e volta ao padrão depois do lock. As ilustrações
  escolhidas diretamente não dependem do cofre e continuam valendo.

## Dados

- Uma coleção própria, `hollow`, dentro do payload criptografado do vault, como as
  tarefas e as notas. Nada fica fora do arquivo do cofre.
- Guarda: o companheiro, o último nível visto, os epitáfios (chave e data) e a lista de
  **deeds** (XP, atributo, tipo de ação, id da origem, data). O XP total, o nível e os
  atributos são calculados a partir dos deeds, então desfazer é só apagar o deed da
  origem.
- O humor não é guardado: é calculado na hora a partir do Grimoire e do Siege.
- Os deeds guardam o **id** da origem, nunca o título. Apagar a origem apaga o deed.
- Deeds com mais de um ano podem ser compactados num deed só por mês, para o vault
  não crescer sem limite.
- **Sem rede.** Nenhum ranking, nenhuma comparação, nada sai da máquina.

## Terminal

### `dtvlt hollow`

Um retrato do personagem no terminal, no estilo do `neofetch`: a arte à esquerda e a
ficha à direita.

```
$ dtvlt hollow

         .###.         hollow   level 7 · hermit · the unkindled
        :#####:        xp       [##########......]  340 / 500
        ##-=-##
        :#####:        vigor    ##########  58
      .:#######:.      memory   #####.....  31
     :##:#####:##:     ward     ####......  24
     ## :#####: ##     hoard    #.........   6
    /   :#####:  #     thrift   ##........  12
        :##:##:  #
        :##:##:        vault    142 entries · 2 keys rotting
       .###:###.       siege    3 orders · 1 in battle · 4 conquered today
                       since    12 Mar 2026 · 9 epitaphs
```

- Lê o personagem pela API de controle, como `dtvlt status`. Com o cofre travado, não
  mostra a arte nem números, só `vault locked · the hollow sleeps`.
- Segue as regras do [terminal](ESTILO.md#terminal): rótulos esmaecidos em 8 colunas,
  valores normais, o nome em negrito, dois espaços de indentação. Sem cores.
- A arte é **a carta do companheiro em ASCII**: a moldura em caracteres de caixa, com o
  numeral no topo e o nome no pé, de cabeça para baixo se estiver invertida, e dentro
  dela a ilustração convertida na hora: a imagem é reduzida a 40 colunas (cada linha de texto cobre duas linhas de
  pixels, porque a célula de 0,6 × 1 fica quase quadrada), e o tom de cada célula vira
  um caractere da rampa do guia (` `, `.`, `:`, `+`, `#`). O desenho acima é só um
  esboço da posição.
- Com `NO_COLOR` ou saída redirecionada, sai texto puro, bom para colar num README.
- No console é `hollow`.
- **`--art`** imprime só o boneco; **`--card`** só a ficha.

### Outros

- `dtvlt status` ganha uma linha: `hollow   lv 7 · 340 / 500`.
- O log da sessão registra ganhos e níveis com o mesmo tom dos outros eventos:
  ```
  14:20  +10 souls  task conquered
  17:24  level 8
  ```
- `dtvlt task` e `dtvlt note` contam igual ao navegador, porque passam pela mesma
  sessão.

## Settings

Em **settings → appearance**, uma linha nova:

| Opção | Valores | Padrão |
|---|---|---|
| hollow | `on` / `off` | `on` |

Com `off`, o ícone some e nada novo é registrado. O que já existia fica guardado e
volta com `on`. A escolha fica em `hollow` no `config.toml`.

## Ordem de construção

Uma versão MINOR para cada passo:

1. **Ícone e página vazia** *(feito)*: a chave `hollow` nos três temas, o botão no
   topo da barra e a placa "nothing here yet".
2. **XP, níveis e atributos** *(feito)*: a coleção de deeds, os ganhos do Siege, a
   barra, os atributos, o histórico, os títulos por tema, o ícone em destaque num nível
   novo e o log da sessão.
3. ~~**Classe e sprite em pixel art**~~ *(feito na 0.22 e 0.23, depois descartado)*: as
   cinco classes em pixel art ficaram longe do que se queria e dão lugar aos
   companheiros. O layout da página (arte à esquerda, status à direita) fica.
4. **As ilustrações:** gerar no Midjourney a primeira leva, um companheiro de cada
   categoria (`hermit`, `raven`, `reaper`, `imp`, `ronin`, `pixie`, `goblin`), e o script de conversão para tons
   de cinza e WebP.
5. **A carta na página** *(feito, com imagem provisória)*: a moldura com o numeral e o nome, a moldura como barra de XP,
   a imagem nos tons do tema, a leitura à direita, a escolha tirando uma carta (com os
   selos de categoria) e a troca livre. As classes saem.
6. **Os outros companheiros:** o resto das categorias, com as próprias de cada tema
   (fae e denizens), e a ordem de cada tema
   (`companions`).
7. **Tela de descanso e `dtvlt hollow`:** os estilos de descanso e o retrato em ASCII.
8. **As outras fontes de XP:** Unspoken, Grimoire, Crypt, Ledger e o rested.
9. **A carta completa:** a carta invertida, o verso, os selos de epitáfio, a moldura
   que envelhece, a leitura do dia, a vela e a crônica.
10. **Estágios:** cada companheiro cresce em três ilustrações, que mudam em níveis
    fixos (filhote e adulto, diabrete e demônio). Fica por último porque triplica as
    imagens.

## Fora de questão

- Reproduzir personagens de outras obras (Guts, Ryuk, a carta do Eremita em si) ou
  citar os nomes delas nos prompts. Elas são referência de clima, não de arte.
- Personalizar o companheiro peça por peça: as peças geradas não encaixam umas nas
  outras sem muito retoque, e a escolha do companheiro já faz esse papel.
- Moeda, loja, sorteio: tudo vem de esforço.
- Perder XP, nível ou o companheiro por ficar longe do cofre; fome, doença ou morte.
- Mostrar o título de entradas, notas ou tarefas no histórico, na ficha ou na crônica.
- Qualquer coisa que dependa de rede.

## Em aberto

- [ ] Os nomes do arcane (`familiar`) e do dungeon (`hero`) estão bons?
- [ ] O nome da categoria 1: `wanderers`?
- [ ] A categoria própria de cada tema novo, e se as categorias próprias devem ser
      exclusivas do tema (hoje ficam abertas a todos).
- [ ] Os valores de XP e a curva de níveis precisam de uso real para ajustar.
- [ ] Os níveis em que cada estágio muda (passo 10).
- [ ] O companheiro também aparece pequeno ao lado da arte do login (opcional,
      desligado por padrão)? O login não tem acesso ao cofre, então seria só a escolha
      guardada no `config.toml`.
- [ ] Exportar o retrato como `.txt` pela página, além do `dtvlt hollow --art`.
