# Tema: souls

> O tema padrão. Este documento dita o estilo do `souls`: o clima, a paleta, os nomes,
> os ícones, a arte e os companheiros do hollow. O que vale para todos os temas está
> em [ESTILO.md](../ESTILO.md), o formato do arquivo em [TEMAS.md](../TEMAS.md) e o
> hollow em [PERSONAGEM.md](../PERSONAGEM.md).

## O clima

**Ruínas, lápides, ossos e livros antigos, guardados como natureza-morta.** Um mundo
que já acabou e que alguém continua visitando: chama baixa, pedra fria, poeira em cima
de tudo. Nada se mexe, nada brilha. O cofre é um ossuário bem cuidado.

- **Sim:** silêncio, decadência digna, cinza e âmbar, coisas antigas que ainda servem.
- **Não:** sangue, gore, caveiras sorrindo, roxo neon, qualquer coisa "épica".

Palavras que o tema usa: *bury*, *unkindled*, *hollow*, *dead words*, *wayfarer*.

## A paleta

Quase preto e cinzas neutros, com um único âmbar de vela.

| Papel | Cor | Uso |
|---|---|---|
| `bg` | `#0a0a0a` | fundo |
| `raise` | `#121212` | o que flutua |
| `line` | `#1e1e1e` | bordas e divisórias |
| `dim` | `#555555` | rótulos, dicas |
| `muted` | `#8a8a8a` | botões, valores secundários |
| `fg` | `#e4e4e4` | texto principal |
| `accent` | `#c8a96a` | âmbar de vela: só mudança de estado |
| `red`, `blue`, `green` | `#c9564b`, `#6f93c4`, `#7fa66a` | só as tags do siege |

Os cinzas são neutros, sem puxar para nenhuma cor: o âmbar é a única temperatura.

## Os nomes

As páginas usam os nomes originais do DataVault, que nasceram neste tema:
**grimoire**, **unspoken**, **siege**, **crypt**, **ledger**. As colunas do siege são
`Orders`, `In Battle`, `Conquered`, `Fallen`; guardar um arquivo é `bury`; o botão de
encerrar é `end session`.

O tom é seco e fúnebre, sem drama: verbos simples (*bury*, *strike*), substantivos
antigos (*crypt*, *ledger*).

## Os ícones

**Formas cheias**, pesadas, como objetos de pedra e couro: o livro fechado, o coração
partido, a espada, a cruz sobre a cova, a balança, a lápide, o capuz
vazio do hollow. Vãos de 2 pixels separam as partes. O `grimoire` é desenhado um
passo maior (`large_icons`), porque deixa mais vazio no quadro.

## A arte

- **Login:** "DataVault" em letras de ASCII de traço grosso (`@`, `!`, `:`), que se
  desfazem de cima para baixo, como pedra gasta.
- **Tela de descanso:** as artes de natureza-morta combinam com ele: `pillar`,
  `skull-i`, `skull-ii` e `candle`.

## O hollow

| Chave | Nome no souls |
|---|---|
| `hollow` | hollow |
| `xp` | souls |
| `level` | level |
| `deeds` | deeds |
| atributos | vigor, memory, ward, hoard, thrift |

**Títulos**, de 5 em 5 níveis: *the nameless*, *the unkindled*, *the wayfarer*, *the
warden*, *the undying*.

**O verso das cartas** é a caveira encapuzada do tema em treliça.

### Os companheiros

**A categoria própria do souls é [wanderers](../PERSONAGEM.md#wanderers)**: o
eremita, o mercenário, o feiticeiro, a bruxa, o necromante. São figuras humanas,
solitárias e gastas, como quem ainda anda pelas ruínas.

Ordem das categorias na escolha (`companions`): **wanderers, shinigami, bushi**, e
depois as outras. Todos os companheiros continuam disponíveis.

Os que mais combinam, fora da categoria própria:

| Companheiro | Por quê |
|---|---|
| `reaper` | a morte paciente, sem pressa, do lado de quem anda entre lápides |
| `raven` | o pássaro das ruínas e dos cemitérios |
| `ronin` | o guerreiro sem senhor, sozinho na estrada |
| `komuso` | o monge de rosto escondido, quieto como uma estátua |

## Ao criar algo novo para o souls

- Pergunte se aquilo poderia estar numa ruína há cem anos. Se brilha ou é novo demais,
  não é do souls.
- Um objeto por vez, parado, em pedra, osso, couro ou ferro.
- O âmbar só quando algo muda; o resto em cinza.
