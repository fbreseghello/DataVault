# DataVault — Arte do hollow

Os prompts do Midjourney para as ilustrações dos companheiros do hollow (veja
[PERSONAGEM.md](PERSONAGEM.md#a-arte)). Todos seguem o mesmo molde, para os
companheiros parecerem saídos do mesmo baralho.

## Como gerar

1. **Gere o `hermit` primeiro** e escolha a melhor variação. Ele é a referência de estilo
   de todos os outros.
2. Nos outros prompts, troque `<sref>` pelo link da imagem do hermit (ou pelo código de
   estilo que o Midjourney mostra). Se ainda variarem muito, use também o mesmo
   `--seed` do job do hermit.
3. Salve cada escolhida em `art/hollow/` com a chave no nome: `hermit.png`,
   `raven.png`, `reaper.png`...
4. A primeira leva é um de cada categoria: **`hermit`, `raven`, `reaper`, `imp` e `ronin`**. Com
   eles aprovados no app, vêm os outros.

As cores que vierem não importam: o app converte tudo para os tons do tema. O que conta
é o traço, a composição e o fundo escuro e liso.

## O molde

Cada prompt é a descrição do companheiro entre duas partes fixas:

**Começo:**
```
illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit,
```

**Fim:**
```
, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

No hermit, que é o primeiro, tire o `--sref <sref>`.

Os prompts abaixo já vêm montados, prontos para colar.

## Wanderers

**hermit**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, an old man in a long grey hooded cloak, a long white beard, raising a lantern with a six-pointed star inside, leaning on a tall wooden staff, standing on a bare snowy peak, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**mercenary**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a scarred mercenary in dark battered armor and a torn cape, an enormous plain greatsword resting on his shoulder, a hard tired face, standing on bare rocky ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**sorcerer**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a gaunt sorcerer in a long robe with a deep hood, a carved wooden staff in one hand, a heavy grimoire chained to his belt, standing on a bare stone floor, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**witch**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, an old witch in a wide-brimmed hat and a ragged shawl, bundles of dried herbs hanging from her belt, holding a lit candle, standing on bare ground with dry grass, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**necromancer**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a thin hooded necromancer in a black robe, a lantern of cold pale light in one hand, a few scattered bones and a skull at his feet, standing on bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

## Animals

**dog**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a lean loyal dog sitting upright with ears raised, a worn leather collar, on bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**cat**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a black cat sitting upright with its tail curled around its paws, calm watchful eyes, on a bare stone step, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**raven**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a large raven perched on a bare dead branch, feathers ruffled, a sharp heavy beak, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**bat**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a bat hanging upside down from a bare twisted branch, wings folded around its body, small sharp ears, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**owl**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a horned owl perched on a bare branch, large round eyes, feathers drawn in clean ink lines, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**wolf**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a grey wolf standing on all four legs, thick fur, alert ears, a steady gaze, on a bare snowy slope, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**serpent**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a serpent coiled on a flat stone, its head raised and tongue out, scales drawn in simple ink patterns, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

## Shinigami

**reaper**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a tall skeletal reaper in a long tattered black robe and hood, a bony face in shadow, holding a great scythe upright, standing on bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**gaunt**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a gaunt angular death spirit, very tall and thin, large round staring eyes, a wide crooked grin, spiky ragged hair, thin tattered wings, long bony fingers, standing on bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**masked**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a death spirit wrapped in layers of grey cloth, wearing a smooth bone mask with two dark eye holes, holding a small black notebook, standing on bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

## Demons

**imp**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a small wiry imp with short curved horns, bat wings, a pointed tail and a sly grin, crouched on a bare rock, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**brute**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a huge hunched demon with twisted ram horns, broken iron chains hanging from its wrists, heavy shoulders, standing on cracked bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**fiend**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a slender winged fiend with great leathery wings folded behind it, long curved claws, a long thin tail, swept-back horns, standing on bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**hellhound**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a hellhound standing on all four legs, ribs showing through its hide, embers glowing in its fur, smoke rising from its back, bared teeth, on cracked bare ground, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

## Bushi

**samurai**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a samurai in full lacquered armor and a horned kabuto helmet, a katana and a short sword at the waist, standing firm with one hand on the hilt, on a bare wooden floor, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**ronin**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a masterless ronin in a wide straw hat that shades the eyes, a worn and patched kimono, a katana resting on his shoulder, standing on a bare dirt road, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**shinobi**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a shinobi in dark close-fitting clothes, the face covered but for the eyes, a short straight blade held low against the body, standing on a bare rooftop at night, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**komuso**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a komuso monk wearing a woven straw basket hat that hides the whole head, a plain dark robe and a kesa sash, holding a long bamboo shakuhachi flute, standing on a bare stone path, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

**sohei**
```
/imagine prompt: illustration in the style of the 1909 Rider-Waite-Smith tarot artwork, like The Hermit, a sohei warrior monk in a dark monastic robe and a white cloth hood around the head, a naginata polearm with a curved blade held upright, standing on bare stone steps, solitary full-length figure facing three-quarters to the left, quiet and solemn, bold black ink outlines, flat areas of muted tone, little texture, sparse composition, plain flat dark background, the whole figure inside the image with a margin around it --ar 2:3 --style raw --sref <sref> --no border, frame, card, text, letters, numbers, title, watermark, signature
```

## Uma imagem por companheiro

Não é preciso gerar uma versão "inquieta": quando o cofre tem pendências, o app mostra
a carta **invertida** (veja [o humor](PERSONAGEM.md#o-cofre-no-humor-a-carta-invertida)).
A moldura, o número e o nome também são desenhados pelo app, então as imagens devem vir
**sem** moldura, título ou número, como pedem os prompts.
