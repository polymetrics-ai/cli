// Generates website/public/penrose-dots.svg — a faint, aperiodic dot field
// placed at the VERTICES of a Penrose P3 (rhombus) tiling. This is the same
// quasiperiodic point set behind "math that goes on forever but never repeats"
// (Quanta, aperiodic monotile / Penrose tilings). Used as a subtle site
// background that never visibly repeats while keeping focus on content.
//
// Run: node scripts/gen-penrose.mjs   (or: npm run gen:penrose)

import { writeFileSync, mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const OUT = resolve(__dirname, '../public/penrose-dots.svg');

const PHI = (1 + Math.sqrt(5)) / 2;

// ── point helpers ([x, y]) ──────────────────────────────────────────────
const sub = (a, b) => [a[0] - b[0], a[1] - b[1]];
const lerp = (a, b, t) => [a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t];

// ── seed: wheel of 10 Robinson (acute) triangles around the origin ──────
let tris = [];
for (let i = 0; i < 10; i++) {
  let b = [Math.cos(((2 * i - 1) * Math.PI) / 10), Math.sin(((2 * i - 1) * Math.PI) / 10)];
  let c = [Math.cos(((2 * i + 1) * Math.PI) / 10), Math.sin(((2 * i + 1) * Math.PI) / 10)];
  if (i % 2 === 0) [b, c] = [c, b]; // mirror alternating wedges
  tris.push([0, [0, 0], b, c]);
}

// ── deflation: subdivide each triangle (golden-ratio split) ─────────────
function subdivide(triangles) {
  const out = [];
  for (const [color, A, B, C] of triangles) {
    if (color === 0) {
      // acute ("thin") triangle -> 1 acute + 1 obtuse
      const P = lerp(A, B, 1 / PHI);
      out.push([0, C, P, B], [1, P, C, A]);
    } else {
      // obtuse ("thick") triangle -> 2 obtuse + 1 acute
      const Q = lerp(B, A, 1 / PHI);
      const R = lerp(B, C, 1 / PHI);
      out.push([1, R, C, A], [1, Q, R, B], [0, R, Q, A]);
    }
  }
  return out;
}

const SUBDIVISIONS = 6; // density of the field
for (let i = 0; i < SUBDIVISIONS; i++) tris = subdivide(tris);

// ── collect unique vertices ─────────────────────────────────────────────
const seen = new Set();
const verts = [];
for (const [, A, B, C] of tris) {
  for (const p of [A, B, C]) {
    const key = `${p[0].toFixed(4)},${p[1].toFixed(4)}`;
    if (seen.has(key)) continue;
    seen.add(key);
    verts.push(p);
  }
}

// ── project to an SVG square; overfill so the unit circle reaches corners ─
const SIZE = 1600;
const CX = SIZE / 2;
const CY = SIZE / 2;
const SCALE = SIZE * 0.72; // unit radius 1 -> ~corner of the square
const DOT_R = 1.8;

// Mild radial fade: a touch lighter toward the centre, denser toward the
// edges, so when used as a fixed/cover background the reading area stays calm.
function dotOpacity(d) {
  const t = Math.min(1, Math.max(0, (d - 0.1) / 0.85));
  return 0.06 + 0.13 * t; // centre ~0.06 -> edge ~0.19
}

let circles = '';
let count = 0;
for (const [x, y] of verts) {
  const d = Math.hypot(x, y);
  if (d > 1.04) continue; // trim ragged outer fringe
  const px = (CX + x * SCALE).toFixed(1);
  const py = (CY + y * SCALE).toFixed(1);
  const op = dotOpacity(d).toFixed(3);
  circles += `<circle cx="${px}" cy="${py}" r="${DOT_R}" opacity="${op}"/>`;
  count++;
}

const svg =
  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${SIZE} ${SIZE}" ` +
  `preserveAspectRatio="xMidYMid slice" shape-rendering="geometricPrecision">` +
  // deep-emerald dots (matches --line-cta), opacity carried per dot
  `<g fill="#0f3d2e">${circles}</g>` +
  `</svg>`;

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, svg, 'utf8');
console.log(
  `Wrote ${count} Penrose-vertex dots to public/penrose-dots.svg ` +
    `(${SUBDIVISIONS} subdivisions, ${(svg.length / 1024).toFixed(0)} KB).`,
);
