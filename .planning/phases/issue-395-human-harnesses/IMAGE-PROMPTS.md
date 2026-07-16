# ChatGPT Images Direction: Humans Need Harnesses Too

## Shared Art Direction

Use the following paragraph at the start of every ChatGPT Images prompt:

> Create a high-end editorial technology image for the Polymetrics engineering journal. The visual
> language is tactile evidence, workshop precision, and real operational scale: photographed paper,
> brushed metal, translucent acrylic, terminal glass, stamped test receipts, and restrained studio
> lighting. Palette: pale mineral green `#f1f6f3`, deep forest `#0f3d2e`, ink `#0c1f17`, emerald
> `#34d399`, with small semantic accents of signal red `#d95c5c` and electric cobalt `#3468d4`.
> Crisp rather than dreamy, intelligent rather than futuristic, with visible material texture and
> generous negative space. No gradients, bokeh, floating orbs, humanoid robots, fantasy imagery,
> stock-photo handshakes, fake logos, legible UI text, or decorative code. Do not place words,
> numbers, watermarks, or a border in the image.

Export each approved image as WebP at 85-90 quality. Preserve the requested crop; do not crop a
single master into every placement.

## 01. The Diff That Ate The Room

- File: `website/public/blog/human-harnesses/01-diff-that-ate-the-room.webp`
- Placement: full width immediately below the article header and above section 01.
- Aspect ratio: 16:9, 2400 x 1350.
- Alt text: `An enormous accordion-fold code diff filling a review room around one small terminal.`
- Prompt:

> [Shared art direction] Wide overhead editorial photograph of a quiet code-review room overwhelmed
> by one continuous accordion-fold printout that snakes across the floor, over tables, and out of
> frame. A single compact terminal sits at the center as the only dark geometric anchor. Add a few
> tiny red rejection stamps and green verification stamps as physical marks, with no readable text.
> The scale should feel absurd but plausible, like an engineering post-mortem photographed for a
> design magazine. Compose the paper flow from upper left toward the central terminal and leave calm
> negative space in the upper right for the page rhythm. Straight overhead camera, sharp detail,
> controlled daylight and studio fill, no people.

## 02. From Shared Checkout To Isolated Work

- File: `website/public/blog/human-harnesses/02-isolated-worktables.webp`
- Placement: floated left beside the second and third paragraphs of `The tool after the fire`; stack
  full width above those paragraphs below 768px.
- Aspect ratio: 4:5, 1440 x 1800.
- Alt text: `Four separated worktables handling bounded parts of one connector migration.`
- Prompt:

> [Shared art direction] Vertical editorial still life of four narrow worktables separated by thin
> translucent green partitions. Each table holds one bounded piece of the same engineering job: a
> compact terminal, one connector card, a small fixture tray, and a single verification receipt.
> Colored cables run from the tables toward one shared integration rail at the far edge, but never
> cross between tables. The image should communicate isolated worktrees and explicit ownership
> without diagrams or labels. Three-quarter camera angle, precise architectural composition, darker
> background, brighter work surfaces, practical studio photography.

## 03. The Repository Becomes The Harness

- File: `website/public/blog/human-harnesses/03-branching-harness.webp`
- Placement: floated right after the first paragraph of `The repository became a harness`; stack
  full width between paragraphs one and two below 768px.
- Aspect ratio: 4:5, 1440 x 1800.
- Alt text: `A physical branch-and-review model routing small work units toward one human gate.`
- Prompt:

> [Shared art direction] Vertical museum-quality photograph of a tabletop architectural model for a
> delivery system. One dark parent rail divides into several narrow, clearly separated work lanes.
> Each lane carries a small translucent change block past a red test station, a green verification
> stamp, and a review lens before rejoining at one manual brass gate. The final gate is visibly
> human-operated but show only the mechanical lever, no person. Make the structure immediately
> understandable through form and spacing, not labels. Right-heavy composition so body copy can sit
> comfortably to the left, crisp macro detail, neutral background.

## 04. Review, Fix, Repeat

- File: `website/public/blog/human-harnesses/04-review-repair-loop.webp`
- Selected source: `~/Downloads/harness blog/8c6bb99b-7ee5-4d41-ae8e-40c25243a8d2.png`
- Status: placed at the original 3:2 composition and exported as WebP quality 88.
- Placement: full-width interruption between paragraphs three and four of `Review, fix, repeat,
  locally`.
- Aspect ratio: 3:2, 2100 x 1400.
- Alt text: `Separate review and repair stations passing the same verified artifact around a loop.`
- Prompt:

> [Shared art direction] Horizontal editorial scene showing two physically separate stations behind
> clear glass: a review station with a magnifying lens and evidence tray, and a repair station with
> precise hand tools. A single sealed dark-green artifact travels between them on a narrow oval rail.
> On the return path it passes through a bright verification aperture before reaching review again.
> Include three small finding tokens in the first tray and one tiny rejected whitespace strip on the
> second lap, all without readable text. The separation of reviewer and fixer must be unmistakable.
> No person, no robot, no shepherd reference. Slightly elevated camera, cinematic but factual,
> restrained motion implied through repeated positions rather than blur.

## 05. An Immutable Artifact Goes To Production

- File: `website/public/blog/human-harnesses/05-immutable-release.webp`
- Placement: panoramic image above `Release and deployment are mutations too`.
- Aspect ratio: 21:9, 2520 x 1080.
- Alt text: `One sealed build artifact moving unchanged through registry and deployment gates.`
- Prompt:

> [Shared art direction] Ultra-wide side-on editorial photograph of one sealed emerald-black artifact
> moving through three distinct physical stations: a verification press, a glass registry vault, and
> a restrained server rack deployment bay. The artifact carries one small embossed fingerprint mark
> that remains visibly identical at every station; do not use readable hashes. Each transition has a
> clear gate and the final bay shows a single healthy green status lamp. Strong horizontal rhythm,
> shallow but readable depth, no cloud iconography, no arrows, no people, no text.

## 06. The Next Question

- File: `website/public/blog/human-harnesses/06-shepherd-teaser.webp`
- Placement: below the Shepherd teaser paragraph and above the final star request.
- Aspect ratio: 16:9, 2400 x 1350.
- Alt text: `A closed workshop observation window watching an active engineering loop.`
- Prompt:

> [Shared art direction] Restrained closing image of a dark workshop observation window looking onto
> the same review-and-repair rail from a distance. The loop is active and orderly inside, while one
> additional unlabelled control console sits outside the glass, powered on but untouched. Suggest the
> question of who supervises the harness without depicting a shepherd, sheep, person, robot, eye, or
> surveillance camera. Quiet composition, more negative space than the earlier images, one green
> status light and one subtle cobalt reflection, precise architectural photography, unresolved but
> not ominous.

## Placement Contract

- Keep only image 01 and image 05 wider than the prose column; they are structural scene changes.
- Float images 02 and 03 at 38-42% of the prose width with at least 24px between image and text.
- Place image 04 between complete paragraphs; never split an annotation block around an image.
- Place image 06 between the Shepherd teaser and the star request so the star remains the final
  interactive action.
- On mobile, every image becomes full width in reading order with a 3:2 or 4:5 crop preserved.
- Captions should identify the represented system concept in one sentence. Do not repeat the alt
  text or explain how the visual was generated.
- Add assets one at a time and run 390px, 768px, 1024px, and 1440px screenshot checks after each
  placement. Do not ship empty frames or generated text inside an image.
