# DataVault — Guia de estilo

> Como o DataVault se parece, se move e fala. Vale para o navegador e para o terminal.
> As decisões de produto estão em [VISAO.md](VISAO.md); aqui fica só a forma.

## Princípio

Uma ferramenta de terminal séria, com um clima gótico contido: preto, branco e cinza,
uma única cor de destaque, pixel art de 1 bit e arte ASCII parada. O tema atual se
chama **souls**: ruínas, lápides, ossos e livros antigos, tratados como natureza-morta.

- **Sim:** monocromático, silêncio visual, texto curto e seco, movimento lento e raro,
  imagens paradas (memento mori), atalhos de teclado sempre à vista.
- **Não:** verde Matrix, "ACCESS GRANTED", glitch, chuva de caracteres, caveiras
  animadas ou "assustadoras", sangue, mensagens dramáticas, emojis.

Se uma tela nova parece um filme de hacker, ela está errada. Se parece uma página de
um livro velho aberta num terminal, está certa.

## Cores

Todas as cores são variáveis em [`web/src/app.css`](../web/src/app.css). Não use cores
fora delas.

| Variável | Valor | Uso |
|---|---|---|
| `--bg` | `#0a0a0a` | fundo de tudo |
| `--raise` | `#121212` | o que flutua: dicas, opções de dropdown |
| `--line` | `#1e1e1e` | bordas, divisórias, sublinhado dos campos |
| `--dim` | `#555` | rótulos, dicas, rodapé, ícones em repouso |
| `--muted` | `#8a8a8a` | botões de texto, valores secundários, hover |
| `--fg` | `#e4e4e4` | texto principal, valores, item ativo |
| `--accent` | `#c8a96a` | **só mudança de estado** (ver abaixo) |
| `--neon-core` / `--neon-glow` / `--neon-haze` | branco | só o neon da arte do login |

**A cor de destaque (âmbar) nunca é decoração.** Ela aparece só quando algo mudou ou
pede atenção:

- erros e confirmações ("saved", "the pins don't match");
- chaves ligadas (`on`) na tela settings;
- itens selados e o cadeado aberto;
- ações destrutivas armadas: a lápide antes do segundo clique, o "delete" do Siege.

A hierarquia vem dos cinzas: `--dim` → `--muted` → `--fg`. Um elemento sobe de tom no
hover e fica em `--fg` quando está ativo.

## Tipografia

- **Uma fonte só:** JetBrains Mono (pesos 400 e 700), embutida via `@fontsource`, sem
  depender de rede. O resto da pilha (`Cascadia Mono`, `Consolas`) é só reserva.
- **Base:** 13 px, entrelinha 1.6.
- **Tamanhos usados:** 11 px (rodapé, contadores, dicas flutuantes), 12 px (dicas e
  descrições), 13 px (corpo), 15 px (o campo do login). Nada maior que isso fora das
  artes ASCII.
- **Negrito** só para a marca (`datavault`) e títulos: da tela, do item ou da nota
  aberta e das colunas do Siege. Rótulos e botões ficam no peso normal.
- **Itálico** só no texto de exemplo (placeholder) dos campos do Unspoken.

## Layout

O app é uma grade fixa: barra lateral, cabeçalho, conteúdo e rodapé.

- **Barra lateral:** 64 px, só ícones, centralizados na vertical com 20 px entre eles.
  A página aberta ganha uma barra de 2 px na borda esquerda, como um cursor. A lápide
  (encerrar sessão) fica sozinha no pé da barra, a 16 px do fundo. A borda da barra é
  uma sombra interna de 1 px, para não roubar largura dos ícones.
- **Cabeçalho:** `datavault / página` à esquerda; sessão, `settings` e `lock` à direita.
- **Barra da página** (`.page-toolbar`): sempre 52 px de altura. Controles da página
  à esquerda; o contador e o `+ new` sempre por último, à direita.
- **Coluna da lista:** `minmax(240px, 360px)`, com uma linha de 1 px à direita.
- **Placa** (`Placard.svelte`, a tela "ícone, nome, dica" quando nada está aberto): o
  ícone fica a 40% da altura do painel e o texto 52 px abaixo. Toda página usa o mesmo
  quadro, inclusive as vazias, para o ícone não pular ao trocar de página.
- **Rodapé:** 11 px, em `--dim`. À esquerda o estado (auto-lock, clipboard); à direita
  os atalhos da página e a versão.
- Abaixo de 720 px a lista some para cima do painel e as dicas flutuantes somem.

## Ícones

Os ícones são pixel art de **1 bit em 16×16**, desenhados como texto em
[`web/src/lib/icons.ts`](../web/src/lib/icons.ts) (`#` é um pixel aceso). O
`Icon.svelte` transforma cada grade num único caminho SVG com `shape-rendering:
crispEdges`.

| Ícone | Nome | Onde |
|---|---|---|
| livro fechado visto de cima, girado, com páginas e fita | `book` | Grimoire |
| coração partido, metades separadas | `heart` | Unspoken |
| espada, ponta para cima e à direita | `sword` | Siege |
| cruz sobre um monte de terra | `cross` | Crypt |
| lápide com inscrição em degraus | `tomb` | encerrar sessão, tela "session ended" |
| cadeado com buraco de chave | `seal` | itens selados |

Regras para desenhar um ícone novo:

- **Só múltiplos de 16** no tamanho final (16, 32, 64), para cada pixel da arte cair
  em pixels inteiros da tela. A barra usa 32; a placa usa 64.
- **Ocupe as linhas 1 a 14**, deixando uma margem; o peso visual deve bater com os
  vizinhos (compare lado a lado com o livro e o coração).
- **Uma cor:** o ícone herda `currentColor`. Em repouso é `--dim`, no hover `--muted`,
  ativo `--fg`. Nunca pinte um ícone com a cor de destaque, exceto num estado armado.
- **Fendas e vãos com 2 pixels.** Um vão de 1 pixel some no tamanho pequeno, e duas
  partes que se tocam só pela quina parecem uma peça só.
- **Silhueta primeiro.** Teste no tamanho real (32 px) antes de detalhar; se não se
  reconhece ali, detalhe nenhum salva.
- Um comentário de uma linha acima da grade diz o que o ícone é.

## Controles

- **Botões são texto**, em minúsculas, sem caixa nem fundo: `save`, `close`, `lock`,
  `+ new`. Cor `--muted`, `--fg` no hover; a ação principal (`save`, `+ new`) já nasce
  em `--fg`.
- **Chaves liga/desliga** são a palavra `on`/`off`; `on` fica na cor de destaque.
- **Campos** não têm caixa: só um sublinhado de 1 px em `--line`, que vai para
  `--muted` no foco. Dropdowns seguem o mesmo desenho, com as opções em `--raise`.
- **Opções que ainda não existem** aparecem desabilitadas, em `--muted`, com uma dica
  dizendo que vêm numa versão futura (ex.: o tema).
- **Ações destrutivas pedem duas vezes:** a primeira arma (cor de destaque e uma dica
  "click again to…"), a segunda executa. Desarma sozinha em 3 s ou ao perder o foco.
- **Foco visível** sempre: contorno de 1 px em `--muted`, afastado 2 px.

## Movimento

O DataVault quase não se mexe. Tudo que se mexe é lento e respeita
`prefers-reduced-motion`.

- **Neon do login:** a arte "liga" piscando (1,6 s), depois o halo respira devagar
  (ciclo de 4 s, só opacidade). De vez em quando, a cada 20–45 s ao acaso, o tubo
  gagueja por 0,9 s sem apagar de todo. Com "reduzir movimento", nada disso acontece.
- **Cursor:** pisca em degrau (sem fade), a cada 1,6 s.
- **Tela de descanso:** parada, sem neon nem animação, exceto o estilo `candle`, a
  única cena que se mexe: a chama balança a cerca de 15 quadros por segundo e a luz
  ao redor muda junto. Com "reduzir movimento", ela fica parada também.
- Sem transições de página, sem slides, sem efeitos de digitação.

## Arte ASCII

- **Login:** "DataVault" em letra gótica ([`art.txt`](../web/src/lib/art.txt)), branca
  com neon. As letras são montadas uma a uma na mesma linha de base.
- **CLI:** o banner do `dtvlt` com a versão alinhada à direita, embaixo, em `dim`.
- **Tela de descanso:** uma imagem no canto inferior direito, sobre o preto:
  - `pillar` ([`pillar.txt`](../web/src/lib/pillar.txt)): a coluna dórica partida,
    renderizada em 3D e convertida em ASCII por
    [`scripts/art/scene.py`](../scripts/art/scene.py);
  - `skull-i` e `skull-ii` ([`skull-i.txt`](../web/src/lib/skull-i.txt),
    [`skull-ii.txt`](../web/src/lib/skull-ii.txt)): o crânio sem mandíbula, em traços
    e em pó, olhando para a esquerda;
  - `candle` ([`candle.ts`](../web/src/lib/candle.ts)): o crânio numa prateleira, iluminado
    por uma vela derretida, contra uma parede de tijolos que só aparece até onde a luz
    chega. A vela é renderizada em 3D por
    [`scripts/art/candle3d.py`](../scripts/art/candle3d.py) para algumas posições da
    chama; o crânio tem três níveis de detalhe e um relevo do osso, feitos por
    [`scripts/art/skull_light.py`](../scripts/art/skull_light.py). A cada quadro, cada
    célula recebe a luz que chega da chama: quanto mais luz, mais denso o caractere e
    mais fino o traço do crânio.

Regras para uma arte nova:

- **Branco sobre preto, só caracteres ASCII.** Rampa de tons do mais leve ao mais
  denso: ` .:-=+*#%@`. Sem blocos Unicode nem braille.
- **Cada arte tem seu tamanho em células** (largura × altura), anotado no CSS que a
  exibe. A fonte é calculada para a arte caber na altura pedida sem passar de metade
  da largura da tela (`Rest.svelte`).
- **Linha de altura 1** (`line-height: 1`) e célula de 0,6 × 1: é a proporção que os
  geradores usam.
- **Espelhar uma arte** é inverter cada linha **e** trocar os caracteres com direção:
  `/`↔`\`, `(`↔`)`, `[`↔`]`, `{`↔`}`, `<`↔`>`.
- Imagens da tela de descanso olham para a esquerda, para dentro da tela.

## Texto e tom

- **A interface é em inglês e em minúsculas**, inclusive títulos: `settings`,
  `session ended`, `nothing here yet`. Nomes de páginas também (`grimoire`,
  `unspoken`, `siege`).
- **Curto e seco.** Diga o que é e o que fazer, sem exclamação nem drama:
  `empty · press n to write the first entry`.
- **O ponto do meio** (` · `) separa ideias numa mesma linha.
- **Atalhos sempre à vista** no rodapé ou na dica: `n new · / search · esc close`.
- **Datas** no formato inglês britânico (`en-GB`).
- **Erros** dizem o que aconteceu e, se der, o que fazer:
  `the pin needs at least 4 characters`.
- **Documentação, CHANGELOG e notas de versão** são em português.

## Terminal

- Só dois estilos: **negrito** e **esmaecido** (`dim`). Sem cores. Com `NO_COLOR` ou
  saída que não é terminal, nenhum estilo.
- Tudo indentado com dois espaços.
- Pares rótulo/valor com o rótulo esmaecido em 8 colunas: `session   7f3a-91c2`.
- Eventos da sessão com hora `HH:MM` esmaecida antes: `17:24  locked`.
- Falhas começam com um `!` em negrito.
- O prompt do console é `dtvlt ›`, com `dtvlt` esmaecido.

## Tema souls

O tema atual. Um tema define **os ícones e os nomes das páginas**; o resto deste guia
(cores, tipografia, layout, tom) vale para todos.

| Página | Ícone | Nomes |
|---|---|---|
| Grimoire | livro fechado em perspectiva | credenciais, comandos de servidor |
| Unspoken | coração partido | notas; **Dead Words** para as enterradas |
| Siege | espada | colunas **Orders**, **In Battle**, **Conquered**, **Fallen** |
| Crypt | cruz | documentos e arquivos guardados |
| — | lápide | encerrar a sessão |

A escolha de tema aparece na tela settings, fixa em **souls** até existir outro.

## Antes de entregar algo novo

- [ ] Usa só as variáveis de cor, e a de destaque só para estado?
- [ ] Tudo em JetBrains Mono, nos tamanhos da lista?
- [ ] O ícone é 16×16, 1 bit, reconhecível em 32 px e com vãos de 2 px?
- [ ] O ícone da placa cai no mesmo ponto das outras páginas?
- [ ] O texto está em minúsculas, curto, com os atalhos visíveis?
- [ ] Se se mexe, é lento e para com "reduzir movimento"?
- [ ] Ações destrutivas pedem confirmação?
- [ ] Funciona com o teclado, com foco visível?
