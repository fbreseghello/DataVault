// The hollow's companions (docs/PERSONAGEM.md, "O companheiro"): what each
// category is, the seal on the back of its card, and who is in it. Any theme can
// choose any companion; a theme only sets which categories come first. The keys
// match internal/vault/hollow.go.

export interface Companion {
  id: string
  /** One line for the choice: what the illustration shows. */
  look: string
}

export interface Category {
  id: string
  /** What the seal on the back of the category's card shows. */
  sealName: string
  /** The seal: 16 rows of 16, '#' lit, drawn like the sidebar icons. */
  seal: string[]
  companions: Companion[]
}

export const CATEGORIES: Category[] = [
  {
    id: 'wanderers',
    sealName: 'a lantern',
    seal: [
      '................',
      '.......##.......',
      '......#..#......',
      '....########....',
      '....#......#....',
      '....#.####.#....',
      '....#.#..#.#....',
      '....#.####.#....',
      '....#..##..#....',
      '....#......#....',
      '....########....',
      '.....######.....',
      '................',
      '.......##.......',
      '......####......',
      '................',
    ],
    companions: [
      { id: 'hermit', look: 'an old man in a hooded cloak, a lantern raised' },
      { id: 'mercenary', look: 'scarred, a greatsword on the shoulder' },
      { id: 'sorcerer', look: 'a long robe, a carved staff, a chained grimoire' },
      { id: 'witch', look: 'a wide hat, dried herbs, a lit candle' },
      { id: 'necromancer', look: 'hooded among bones, a lantern of cold light' },
    ],
  },
  {
    id: 'animals',
    sealName: 'a paw',
    seal: [
      '................',
      '....##....##....',
      '...####..####...',
      '...####..####...',
      '....##....##....',
      '.##..........##.',
      '####........####',
      '####..####..####',
      '.##..######..##.',
      '....########....',
      '...##########...',
      '...##########...',
      '....########....',
      '.....######.....',
      '................',
      '................',
    ],
    companions: [
      { id: 'dog', look: 'lean and loyal, ears raised' },
      { id: 'cat', look: 'a black cat, tail around its paws' },
      { id: 'raven', look: 'perched on a dead branch' },
      { id: 'bat', look: 'hanging, wings folded' },
      { id: 'owl', look: 'perched, great round eyes' },
      { id: 'wolf', look: 'thick fur, a steady gaze' },
      { id: 'serpent', look: 'coiled on a stone, head raised' },
    ],
  },
  {
    id: 'shinigami',
    sealName: 'a scythe',
    seal: [
      '................',
      '...#########....',
      '..##.......###..',
      '.##..........##.',
      '.#.......#....#.',
      '.........#......',
      '........#.......',
      '........#.......',
      '.......#........',
      '.......#........',
      '......#.........',
      '......#.........',
      '.....#..........',
      '.....#..........',
      '....#...........',
      '................',
    ],
    companions: [
      { id: 'reaper', look: 'tall and skeletal, a great scythe' },
      { id: 'gaunt', look: 'thin and angular, a crooked grin' },
      { id: 'masked', look: 'wrapped in cloth, a bone mask, a notebook' },
    ],
  },
  {
    id: 'demons',
    sealName: 'horns',
    seal: [
      '................',
      '.##..........##.',
      '.##..........##.',
      '..##........##..',
      '..###.####.###..',
      '...##########...',
      '...##########...',
      '..####.##.####..',
      '..###..##..###..',
      '..############..',
      '...##########...',
      '....#.#..#.#....',
      '....########....',
      '.....######.....',
      '................',
      '................',
    ],
    companions: [
      { id: 'imp', look: 'small, short horns, bat wings' },
      { id: 'brute', look: 'huge and hunched, broken chains' },
      { id: 'fiend', look: 'slender, great wings, a long tail' },
      { id: 'hellhound', look: 'ribs showing, embers in its fur' },
    ],
  },
  {
    id: 'bushi',
    sealName: 'a straw hat',
    seal: [
      '................',
      '................',
      '................',
      '......####......',
      '.....######.....',
      '....##....##....',
      '...##......##...',
      '..##........##..',
      '.##############.',
      '################',
      '................',
      '......#..#......',
      '.....#....#.....',
      '................',
      '................',
      '................',
    ],
    companions: [
      { id: 'samurai', look: 'full armor and a horned helmet' },
      { id: 'ronin', look: 'a straw hat, a worn kimono, a sword on the shoulder' },
      { id: 'shinobi', look: 'dark clothes, the face covered' },
      { id: 'komuso', look: 'a basket hat over the whole head, a flute' },
      { id: 'sohei', look: 'a white hood, a curved polearm' },
    ],
  },
  {
    id: 'fae',
    sealName: 'a moth wing',
    seal: [
      '................',
      '................',
      '.####......####.',
      '#....#....#....#',
      '#.##..#..#..##.#',
      '#.##...##...##.#',
      '#......##......#',
      '.#.....##.....#.',
      '..#####..#####..',
      '..#...#..#...#..',
      '..#..#....#..#..',
      '...##......##...',
      '................',
      '................',
      '................',
      '................',
    ],
    companions: [
      { id: 'pixie', look: 'tiny, dragonfly wings, on a toadstool' },
      { id: 'wisp', look: 'a pale light over the marsh' },
      { id: 'dryad', look: 'skin of bark, branches for hair' },
      { id: 'moth', look: 'moth wings, drawn to a candle' },
      { id: 'changeling', look: 'a child with eyes far too old' },
    ],
  },
  {
    id: 'denizens',
    sealName: 'a torch',
    seal: [
      '................',
      '.......#........',
      '......##........',
      '......###.......',
      '.....####.#.....',
      '.....#####......',
      '.....##.##......',
      '......####......',
      '.....######.....',
      '.....######.....',
      '......####......',
      '.......##.......',
      '.......##.......',
      '.......##.......',
      '.......##.......',
      '................',
    ],
    companions: [
      { id: 'goblin', look: 'a dagger and a stolen lantern' },
      { id: 'skeleton', look: 'a cracked shield, still on guard' },
      { id: 'slime', look: 'coins and a key held inside' },
      { id: 'mimic', look: 'a chest with teeth' },
      { id: 'golem', look: 'cracked stone, faded runes' },
    ],
  },
]

/** The category a companion belongs to. */
export function categoryOf(companion: string): Category | undefined {
  return CATEGORIES.find((c) => c.companions.some((x) => x.id === companion))
}

/** The categories in a theme's order: the ones it names first, then the rest as listed. */
export function ordered(first: string[] = []): Category[] {
  const front = first.map((id) => CATEGORIES.find((c) => c.id === id)).filter((c): c is Category => !!c)
  return [...front, ...CATEGORIES.filter((c) => !front.includes(c))]
}

/** The level as a card numbers it: 9 is IX. Zero and below have no numeral. */
export function roman(n: number): string {
  if (n < 1) return ''
  const marks: [number, string][] = [
    [1000, 'M'], [900, 'CM'], [500, 'D'], [400, 'CD'], [100, 'C'], [90, 'XC'],
    [50, 'L'], [40, 'XL'], [10, 'X'], [9, 'IX'], [5, 'V'], [4, 'IV'], [1, 'I'],
  ]
  let out = ''
  for (const [v, m] of marks) {
    while (n >= v) {
      out += m
      n -= v
    }
  }
  return out
}
