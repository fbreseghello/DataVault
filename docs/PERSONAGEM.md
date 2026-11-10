# DataVault — Personagem

> **Ideia registrada, ainda não implementada.** Este documento descreve um personagem
> que evolui com o uso do DataVault. Nada aqui existe no código ainda; serve de
> referência para quando a ideia for construída. A forma geral segue o
> [guia de estilo](ESTILO.md) e o formato de temas segue [TEMAS.md](TEMAS.md).

## A ideia

Um personagem em arte ASCII que acompanha quem usa o cofre. Ele ganha **experiência
(XP)** com o que acontece no sistema (concluir tarefas no Siege, escrever no Unspoken,
renovar uma chave no prazo) e sobe de **nível**. Com os níveis vêm títulos e peças
novas para **personalizar** o boneco.

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
| — | abrir o cofre em dias seguidos | 2 × dias da sequência | máximo 20 por dia |

O que **não** dá XP: criar e apagar coisas em sequência, mover uma tarefa de volta e de
novo para Conquered, copiar senhas, abrir páginas, tempo com o app aberto.

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
- **Peças se destravam por nível**, nunca por compra nem por sorte. A lista em `attire`
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

## Dados

- Uma coleção própria, `hollow`, dentro do payload criptografado do vault, como as
  tarefas e as notas. Nada fica fora do arquivo do cofre.
- Guarda: as escolhas de `attire`, o último nível visto e a lista de **deeds** (XP,
  tipo de ação, id da origem, data). O XP total e o nível são calculados a partir dos
  deeds, então desfazer é só apagar o deed da origem.
- Os deeds guardam o **id** da origem, nunca o título. Apagar a origem apaga o deed.
- Deeds com mais de um ano podem ser compactados num deed só por mês, para o vault
  não crescer sem limite.
- **Sem rede.** Nenhum ranking, nenhuma comparação, nada sai da máquina.

## Terminal

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
2. **XP e níveis:** a coleção de deeds, os ganhos do Siege, a barra e o histórico.
3. **As outras fontes de XP:** Unspoken, Grimoire, Crypt, Ledger e a sequência de dias.
4. **Personalizar:** as camadas, as peças por tema e a tela `attire`.
5. **Terminal:** `dtvlt status` e o log.

## Em aberto

- [ ] Os nomes do arcane (`familiar`) e do dungeon (`hero`) estão bons?
- [ ] Os valores de XP e a curva de níveis precisam de uso real para ajustar.
- [ ] A sequência de dias pode virar pressão; vale manter?
- [ ] Atributos (força, sabedoria...) ligados a cada página, ou só o nível basta?
- [ ] Feitos únicos ("primeira chave renovada", "cem tarefas") como deeds especiais?
- [ ] Exportar o personagem junto com o tema, ou separado?
