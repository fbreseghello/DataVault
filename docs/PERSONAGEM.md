# DataVault — Personagem

> **Ideia registrada, ainda não implementada.** Este documento descreve um personagem
> que evolui com o uso do DataVault. Nada aqui existe no código ainda; serve de
> referência para quando a ideia for construída. A forma geral segue o
> [guia de estilo](ESTILO.md) e o formato de temas segue [TEMAS.md](TEMAS.md).

## A ideia

Um personagem em arte ASCII que acompanha quem usa o cofre. Ele ganha **experiência
(XP)** com o que acontece no sistema (concluir tarefas no Siege, escrever no Unspoken,
renovar uma chave no prazo) e sobe de **nível**. Quem usa escolhe a **classe** dele, e
com os níveis vêm títulos e peças novas para **personalizar** o boneco. Ele também
aparece fora da página: na **tela de descanso** e no terminal, com o `dtvlt hollow`.

Não é um jogo que pede atenção. É um diário de campanha silencioso: o personagem fica
parado, como as outras artes do DataVault, e só registra o caminho percorrido.

- **Sim:** progresso lento e honesto, texto seco, arte parada, tudo local e criptografado.
- **Não:** "LEVEL UP!!!", confete, sons, notificações insistentes, ranking, metas que
  culpam quem não abriu o app, recompensa por ações sem sentido.

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
Os desenhos abaixo são a proposta inicial; antes de entrar no código, cada um precisa
ser testado em 32 px ao lado dos vizinhos, como pede o [guia](ESTILO.md#ícones).

### souls: elmo

Elmo fechado de cavaleiro, visto de frente. A fenda dos olhos é um vão horizontal de
2 pixels e a respiração, um vão vertical de 2 pixels.

```
................
.....######.....
...##########...
..############..
..############..
.##############.
.##############.
.##..........##.
.##..........##.
.##############.
.######..######.
.######..######.
.######..######.
..#####..#####..
...####..####...
................
```

### arcane: chapéu de mago

O cone torto do chapéu, a faixa como um vão de 2 pixels e a aba larga embaixo.

```
................
...........##...
..........###...
.........###....
........####....
.......#####....
......#######...
......#######...
.....#########..
.....#########..
................
................
.##############.
..############..
................
................
```

### dungeon: escudo

Escudo de herói de 8 bits, com uma cruz gravada em vãos de 2 pixels.

```
................
.##############.
.######..######.
.######..######.
.######..######.
.##..........##.
.##..........##.
.######..######.
.######..######.
..#####..#####..
..#####..#####..
...####..####...
....###..###....
.....##..##.....
................
................
```

Um tema importado que não defina `hollow` herda o elmo do `souls`, como acontece com
qualquer ícone que falte.

## Nomes por tema

Novas chaves em `words`, com as mesmas regras (uma linha, até 24 caracteres):

| Chave | Onde aparece | souls | arcane | dungeon |
|---|---|---|---|---|
| `hollow` | nome da página e da dica | hollow | familiar | hero |
| `xp` | a experiência | souls | mana | xp |
| `level` | o nível (abreviado `lv` na dica) | level | circle | level |
| `deeds` | o histórico de XP | deeds | grimoire notes | quest journal |
| `attire` | a tela de personalizar | attire | regalia | gear |

## A página

Segue o mesmo quadro das outras: barra da página no topo, coluna da esquerda e painel.

```
datavault / hollow
──────────────────────────────────────────────────────────────────
 lv 7 · the unkindled                      deeds · attire      ← barra
──────────────────────────────────────────────────────────────────
 today                          │
 +10  task conquered      14:20 │           .###.
 +5   note written        11:02 │          :#####:
 +15  key renewed         09:40 │          ##-=-##
                                │          :#####:
 yesterday                      │        .:#######:.
 +10  task conquered      18:03 │       :##:#####:##:
 +10  task conquered      16:41 │       ## :#####: ##
 ...                            │          :##:##:
                                │         .###:###.
                                │
                                │      level 7 · the unkindled
                                │   [##########......]  340 / 500
──────────────────────────────────────────────────────────────────
 lv 7 · 160 to next              0 open · a attire · d deeds · v1.x
```

- **Painel:** o personagem no lugar do ícone da placa, o nível e o título embaixo, e a
  barra de XP em ASCII: `[`, `#` para o ganho, `.` para o que falta, `]`, e os
  números à direita.
- **Coluna da esquerda (`deeds`):** o histórico de XP agrupado por dia, com o ganho, o
  motivo e a hora. Nunca mostra o título de notas, entradas ou tarefas, só o tipo de
  ação, como o log da sessão no terminal.
- **`attire`:** troca a coluna da esquerda pela lista de peças (veja
  [Personalizar](#personalizar)); escolher uma peça mostra o boneco já com ela.
- O personagem olha para a esquerda, para dentro da tela, como as artes da tela de
  descanso.

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

## Classe

Quem usa **escolhe a classe** do personagem. Ela define a silhueta base, as peças
iniciais e o atributo de afinidade (veja [Atributos](#atributos)). Cada tema dá nomes e
peças próprias às classes; a chave interna é a do `souls`.

| Chave | souls | arcane | dungeon | Afinidade | Silhueta |
|---|---|---|---|---|---|
| `knight` | knight | battlemage | warrior | vigor | ombros largos, elmo, espada |
| `scribe` | scribe | scholar | bard | memory | encurvado, capuz, livro e pena |
| `warden` | warden | enchanter | paladin | ward | escudo à frente, armadura fechada |
| `thief` | thief | trickster | rogue | hoard | magro, capuz baixo, adagas |
| `cleric` | cleric | oracle | priest | thrift | manto longo, lanterna ou cajado |

- **A primeira vez** que a página abre, ela mostra as cinco classes lado a lado, em
  ASCII, com o nome e uma linha de descrição. Setas escolhem, `Enter` confirma.
- **Trocar depois** é livre e não custa XP nem nível: em `attire`, **class** abre a mesma
  escolha. As peças que não servem na classe nova voltam para as iniciais dela; as
  destravadas continuam destravadas.
- O painel mostra a classe entre o nível e o título: `level 7 · knight · the unkindled`.
- **Afinidade:** o atributo da classe ganha 10% a mais de pontos. É um empurrão, não
  uma trava: um knight que só escreve notas ainda fica com memory alto.

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
- Peças especiais de `attire` podem pedir um atributo mínimo além do nível
  (`lv 10 · ward 30`).

## Personalizar

O boneco é montado em **camadas de ASCII**, sobrepostas de baixo para cima. Um caractere
diferente de espaço numa camada cobre o que está embaixo.

| Camada | O que é | Exemplos (souls) |
|---|---|---|
| `body` | silhueta base | magro, robusto, encurvado |
| `head` | cabeça ou elmo | capuz, elmo fechado, coroa partida |
| `cloak` | o que vai nas costas | capa rasgada, manto, nada |
| `hand` | o que segura | espada, cajado, lanterna, livro |
| `mark` | um detalhe pequeno | cicatriz, brasão, runa |

- Cada peça é uma grade de texto do tamanho do boneco (proposta: **20 × 12 células**),
  com espaço onde é transparente.
- A classe escolhe o `body` e as peças iniciais; as outras camadas são livres.
- **Peças se destravam por nível**, por atributo ou por epitáfio (veja
  [Epitáfios](#epitáfios)), nunca por compra nem por sorte. A lista em `attire`
  mostra todas; as que faltam aparecem desabilitadas, em `--muted`, com a dica do
  nível que as libera (`lv 10`), como as opções futuras em settings.
- Cada tema traz o seu conjunto de peças, combinando com o clima: elmos e capas no
  `souls`, chapéus e cajados no `arcane`, armaduras e tochas de 8 bits no `dungeon`. Ao
  trocar de tema, a escolha de cada camada vai para a peça de mesma posição no novo
  conjunto.
- Temas importados podem trazer peças num campo novo, `attire`, com as mesmas regras
  dos ícones: chaves conhecidas, tamanho fixo, nada de caracteres fora do ASCII.

### Regras da arte

As de [arte ASCII](ESTILO.md#arte-ascii), mais estas:

- Só a rampa ` .:-=+*#%@` e os traços `/ \ | _ ( ) [ ] < >`. Sem blocos Unicode.
- **Parado.** O personagem não pisca, não respira, não anda.
- Olha para a esquerda. Peças com direção (espada, capa) são desenhadas assim; para
  espelhar, use a regra do guia (`/`↔`\`, `(`↔`)` e o resto).
- Em `--fg`, sem brilho. A cor de destaque nunca entra no boneco.
- Toda combinação de peças precisa continuar legível: cada camada é testada sobre cada
  `body`.

Rascunho do boneco base do `souls` (elmo, corpo, espada baixa):

```
       .###.
      :#####:
      ##-=-##
      :#####:
    .:#######:.
   :##:#####:##:
   ## :#####: ##
  /   :#####:  #
      :##:##:  #
      :##:##:
     .###:###.
```

A arte final deve sair de um gerador, como a vela e o crânio da tela de descanso
(`scripts/art/`), para as peças manterem o mesmo traço.

## O cofre no corpo

O boneco mostra a saúde do cofre. São camadas extras, por cima do `attire`, que
aparecem e somem sozinhas; quem usa não escolhe nem remove, só resolve a causa.

| Marca | Aparece quando | Some quando |
|---|---|---|
| **armadura gasta:** rachaduras e ferrugem (`:` e `.` no lugar de `#`) | há entradas vencidas no Grimoire ou uma entrada com 2 códigos de recuperação ou menos | tudo é renovado |
| **peso nas costas:** o boneco fica encurvado | há tarefas em In Battle há mais de 3 dias | elas saem de In Battle |
| **saco de moedas:** pendurado no cinto, maior com o gasto | sempre, proporcional ao gasto mensal do Ledger, em 3 tamanhos | contas recebem strike |

- Não tiram XP nem nível. É uma auditoria de relance, não uma punição.
- A dica do ícone na barra lateral diz a causa: `hollow · lv 7 · 2 keys rotting`.
- As marcas seguem as regras da arte: paradas, em `--fg`, sem cor de destaque.

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

- Cada um dá XP uma vez e pode destravar uma peça que não sai por nível.
- A lista mostra a data e o texto; os que faltam ficam ocultos, só com `???`, para não
  virarem uma lista de tarefas.
- Mesmo tom do resto: minúsculas, sem exclamação.

## Ritmo

- **A vela do dia:** ao lado do boneco, uma vela em ASCII, feita com a mesma arte da
  vela da tela de descanso, mas parada. Ela vai derretendo conforme as tarefas do dia
  saem de Orders e é trocada por uma nova à meia-noite. Não apaga por falta de uso.
- **A crônica do ano:** no primeiro acesso de cada ano, a página oferece a crônica do
  ano anterior, uma página de texto com o resumo: tarefas conquistadas, notas escritas,
  chaves renovadas, quanto se cortou no Ledger, níveis e epitáfios. Só números e tipos
  de ação, nunca títulos.

## Tela de descanso

Um estilo novo da tela de descanso, **`hollow`**, ao lado de `pillar`, `skull-i`,
`skull-ii` e `candle`:

- O personagem, com a classe, o `attire` e as marcas do cofre, no canto inferior
  direito, olhando para a esquerda, como as outras artes.
- Parado. Embaixo do boneco, em `--dim` e 11 px, só `lv 7 · the unkindled`.
- O tamanho segue a regra da tela de descanso (`Rest.svelte`): cabe na altura pedida
  sem passar de metade da largura.
- Com o hollow desligado em settings, o estilo some da lista e, se estava em uso, a
  tela volta para o padrão.
- O cofre travado não tem acesso ao personagem; a tela de descanso só mostra o hollow
  enquanto a sessão está aberta, e volta ao estilo padrão depois do lock.

## Dados

- Uma coleção própria, `hollow`, dentro do payload criptografado do vault, como as
  tarefas e as notas. Nada fica fora do arquivo do cofre.
- Guarda: a classe, as escolhas de `attire`, o último nível visto, os epitáfios
  (chave e data) e a lista de **deeds** (XP, atributo, tipo de ação, id da origem,
  data). O XP total, o nível e os atributos são calculados a partir dos deeds, então
  desfazer é só apagar o deed da origem.
- As marcas do cofre não são guardadas: são calculadas na hora a partir do Grimoire,
  do Siege e do Ledger.
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

         .###.         hollow   level 7 · knight · the unkindled
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

- Lê o personagem pela API de controle, como `dtvlt status`. Com o cofre travado, mostra
  só a silhueta da classe e `vault locked · the hollow sleeps`, sem números.
- Segue as regras do [terminal](ESTILO.md#terminal): rótulos esmaecidos em 8 colunas,
  valores normais, o nome em negrito, dois espaços de indentação. Sem cores.
- A arte é a mesma do navegador, com o `attire` e as marcas do cofre.
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

Quando for a hora, uma versão MINOR para cada passo:

1. **Ícone e página vazia:** a chave `hollow` nos três temas, o botão no topo da barra
   e a placa "nothing here yet".
2. **Classe e arte base:** a escolha de classe e a silhueta de cada uma nos três temas.
3. **XP, níveis e atributos:** a coleção de deeds, os ganhos do Siege, a barra e o
   histórico.
4. **Tela de descanso e `dtvlt hollow`:** o boneco fora da página.
5. **As outras fontes de XP:** Unspoken, Grimoire, Crypt, Ledger e o rested.
6. **Personalizar:** as camadas, as peças por tema e a tela `attire`.
7. **O cofre no corpo, epitáfios e ritmo:** as marcas, a lista de epitáfios, a vela e
   a crônica.

## Fora de questão

- Moeda, loja, sorteio de peças: tudo vem de esforço.
- Perder XP, nível ou peças por ficar longe do cofre.
- Mostrar o título de entradas, notas ou tarefas no histórico, na ficha ou na crônica.
- Qualquer coisa que dependa de rede.

## Em aberto

- [ ] Os nomes do arcane (`familiar`) e do dungeon (`hero`) estão bons?
- [ ] Os nomes das classes e dos atributos por tema.
- [ ] Os valores de XP e a curva de níveis precisam de uso real para ajustar.
- [ ] O boneco também aparece pequeno ao lado da arte do login (opcional, desligado
      por padrão)? O login não tem acesso ao cofre, então seria só a silhueta da classe
      guardada no `config.toml`.
- [ ] Exportar o boneco como `.txt` pela página, além do `dtvlt hollow --art`.
- [ ] Exportar o personagem junto com o tema, ou separado?
