// Generates a registry of 100 generative mathematical background patterns.
// Each pattern is a deterministic, faint emerald SVG written to
// website/public/patterns/pNNN.svg, plus a typed manifest at
// website/lib/patterns.generated.ts. Families: phyllotaxis, wave-interference
// (quasicrystal), Penrose tilings, spirals (log/fermat/archimedean), rose &
// Maurer-rose curves, Lissajous, spirographs (hypotrochoids), modular
// times-tables, star polygons, circle ripples, moiré, and harmonographs.
//
// Run: node scripts/gen-patterns.mjs   (or: npm run gen:patterns)

import { writeFileSync, mkdirSync, rmSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const OUT_DIR = resolve(__dirname, '../public/patterns');
const MANIFEST = resolve(__dirname, '../lib/patterns.generated.ts');

const S = 1000;          // viewBox size
const C = S / 2;         // center
const TAU = Math.PI * 2;
const PHI = (1 + Math.sqrt(5)) / 2;
const EM = '#0f3d2e';    // deep emerald
const f = (n) => (Math.round(n * 10) / 10).toString();

// ── svg helpers ─────────────────────────────────────────────────────────
const svg = (inner) =>
  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${S} ${S}" ` +
  `preserveAspectRatio="xMidYMid slice" fill="none" ` +
  `stroke-linecap="round" stroke-linejoin="round">${inner}</svg>`;

const dotsG = (dots) => `<g fill="${EM}">${dots}</g>`;
const linesG = (lines, sw = 0.8) =>
  `<g stroke="${EM}" stroke-width="${sw}">${lines}</g>`;
const dot = (x, y, r, op) =>
  `<circle cx="${f(x)}" cy="${f(y)}" r="${r}" opacity="${op}"/>`;
const polyline = (pts, op) => {
  let d = '';
  for (let i = 0; i < pts.length; i++) d += (i ? 'L' : 'M') + f(pts[i][0]) + ' ' + f(pts[i][1]);
  return `<path d="${d}" opacity="${op}"/>`;
};
const seg = (x1, y1, x2, y2, op) =>
  `<line x1="${f(x1)}" y1="${f(y1)}" x2="${f(x2)}" y2="${f(y2)}" opacity="${op}"/>`;
const inBox = (x, y, m = 6) => x >= -m && x <= S + m && y >= -m && y <= S + m;
const gcd = (a, b) => (b ? gcd(b, a % b) : a);

// ══ FAMILIES ════════════════════════════════════════════════════════════

// 1. Phyllotaxis — golden-angle sunflower; spiral arms emerge.
function phyllotaxis({ angle, n, c }) {
  const ga = (angle * Math.PI) / 180;
  const maxr = C * 0.95;
  let s = '';
  for (let i = 1; i < n; i++) {
    const r = c * Math.sqrt(i);
    const a = i * ga;
    const x = C + r * Math.cos(a), y = C + r * Math.sin(a);
    if (!inBox(x, y)) continue;
    const e = Math.min(1, r / maxr);
    s += dot(x, y, +(0.6 + 1.9 * e).toFixed(2), +(0.05 + 0.16 * e).toFixed(3));
  }
  return dotsG(s);
}

// 2. Wave interference — k rotated line gratings -> quasicrystal moiré.
function waveGrating({ k, spacing }) {
  let s = '';
  for (let j = 0; j < k; j++) {
    const ang = (j * Math.PI) / k;
    const dx = Math.cos(ang), dy = Math.sin(ang);
    for (let off = -760; off <= 760; off += spacing) {
      const px = C + -dy * off, py = C + dx * off;
      s += seg(px - dx * 950, py - dy * 950, px + dx * 950, py + dy * 950, 0.05);
    }
  }
  return linesG(s, 0.7);
}

// Penrose deflation shared core -> triangles within unit circle.
function penroseTris(sub, rotDeg) {
  const rot = (rotDeg * Math.PI) / 180;
  const lerp = (a, b, t) => [a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t];
  let tris = [];
  for (let i = 0; i < 10; i++) {
    let b = [Math.cos((2 * i - 1) * Math.PI / 10 + rot), Math.sin((2 * i - 1) * Math.PI / 10 + rot)];
    let c = [Math.cos((2 * i + 1) * Math.PI / 10 + rot), Math.sin((2 * i + 1) * Math.PI / 10 + rot)];
    if (i % 2 === 0) [b, c] = [c, b];
    tris.push([0, [0, 0], b, c]);
  }
  for (let it = 0; it < sub; it++) {
    const out = [];
    for (const [col, A, B, D] of tris) {
      if (col === 0) {
        const P = lerp(A, B, 1 / PHI);
        out.push([0, D, P, B], [1, P, D, A]);
      } else {
        const Q = lerp(B, A, 1 / PHI), R = lerp(B, D, 1 / PHI);
        out.push([1, R, D, A], [1, Q, R, B], [0, R, Q, A]);
      }
    }
    tris = out;
  }
  return tris;
}

// 3. Penrose vertices as dots.
function penroseDots({ sub, rot }) {
  const tris = penroseTris(sub, rot);
  const scale = C * 0.95, seen = new Set();
  let s = '';
  for (const [, A, B, D] of tris) {
    for (const p of [A, B, D]) {
      const key = p[0].toFixed(3) + ',' + p[1].toFixed(3);
      if (seen.has(key)) continue;
      seen.add(key);
      const d = Math.hypot(p[0], p[1]);
      if (d > 1.02) continue;
      s += dot(C + p[0] * scale, C + p[1] * scale, 1.6, +(0.06 + 0.13 * d).toFixed(3));
    }
  }
  return dotsG(s);
}

// 4. Penrose rhombi as line-art.
function penroseRhombi({ sub, rot }) {
  const tris = penroseTris(sub, rot);
  const scale = C * 0.95;
  let s = '';
  for (const [, A, B, D] of tris) {
    const a = [C + A[0] * scale, C + A[1] * scale];
    const b = [C + B[0] * scale, C + B[1] * scale];
    const d = [C + D[0] * scale, C + D[1] * scale];
    s += `<path d="M${f(a[0])} ${f(a[1])}L${f(b[0])} ${f(b[1])}L${f(d[0])} ${f(d[1])}" opacity="0.13"/>`;
  }
  return linesG(s, 0.6);
}

// 5. Logarithmic (golden) spirals.
function logSpiral({ arms, b }) {
  let s = '';
  const diag = Math.hypot(S, S);
  for (let k = 0; k < arms; k++) {
    const rot = (k * TAU) / arms;
    const pts = [];
    for (let th = -3 * Math.PI; th < 8 * Math.PI; th += 0.05) {
      const r = 0.8 * Math.exp(b * th);
      if (r > diag) break;
      pts.push([C + r * Math.cos(th + rot), C + r * Math.sin(th + rot)]);
    }
    if (pts.length > 1) s += polyline(pts, 0.13);
  }
  return linesG(s, 1);
}

// 6. Fermat spiral (two-branch) as dots.
function fermat({ n, c }) {
  let s = '';
  for (let i = 1; i < n; i++) {
    const r = c * Math.sqrt(i);
    const a = i * 2.39996; // golden angle radians
    for (const sgn of [1, -1]) {
      const x = C + sgn * r * Math.cos(a), y = C + sgn * r * Math.sin(a);
      if (!inBox(x, y)) continue;
      const e = Math.min(1, r / (C * 0.95));
      s += dot(x, y, +(0.7 + 1.3 * e).toFixed(2), +(0.05 + 0.13 * e).toFixed(3));
    }
  }
  return dotsG(s);
}

// 7. Archimedean spirals (rotated arms).
function archimedean({ arms, turns }) {
  let s = '';
  const a = (C * 0.95) / (turns * TAU);
  for (let k = 0; k < arms; k++) {
    const rot = (k * TAU) / arms;
    const pts = [];
    for (let th = 0; th < turns * TAU; th += 0.05) pts.push([C + a * th * Math.cos(th + rot), C + a * th * Math.sin(th + rot)]);
    s += polyline(pts, 0.12);
  }
  return linesG(s, 0.9);
}

// 8. Rose curves (rhodonea) r = cos(k θ).
function rose({ num, den }) {
  const k = num / den;
  const R = C * 0.92;
  const period = (num * den) % 2 === 0 ? den * TAU : den * Math.PI;
  const pts = [];
  for (let t = 0; t <= period + 0.01; t += 0.01) {
    const r = Math.cos(k * t);
    pts.push([C + R * r * Math.cos(t), C + R * r * Math.sin(t)]);
  }
  return linesG(polyline(pts, 0.15), 0.9);
}

// 9. Maurer rose — straight chords through a rose.
function maurerRose({ n, d }) {
  const R = C * 0.92;
  const pts = [];
  for (let i = 0; i <= 360; i++) {
    const k = i * d;
    const rad = (k * Math.PI) / 180;
    const r = Math.sin(n * rad);
    const t = (k * Math.PI) / 180;
    pts.push([C + R * r * Math.cos(t), C + R * r * Math.sin(t)]);
  }
  return linesG(polyline(pts, 0.11), 0.55);
}

// 10. Lissajous curves.
function lissajous({ a, b, delta }) {
  const R = C * 0.92, dl = (delta * Math.PI) / 180;
  const pts = [];
  for (let t = 0; t <= TAU + 0.01; t += 0.005) pts.push([C + R * Math.sin(a * t + dl), C + R * Math.sin(b * t)]);
  return linesG(polyline(pts, 0.14), 0.9);
}

// 11. Spirograph (hypotrochoid).
function spirograph({ R, r, d }) {
  const reps = r / gcd(R, r);
  const pts = [];
  const norm = R - r + d || 1;
  const sc = (C * 0.95) / norm;
  for (let t = 0; t <= TAU * reps + 0.01; t += 0.01) {
    const x = (R - r) * Math.cos(t) + d * Math.cos(((R - r) / r) * t);
    const y = (R - r) * Math.sin(t) - d * Math.sin(((R - r) / r) * t);
    pts.push([C + sc * x, C + sc * y]);
  }
  return linesG(polyline(pts, 0.12), 0.7);
}

// 12. Modular times-table on a circle (cardioid family).
function timesTable({ n, m }) {
  const R = C * 0.95;
  const P = (i) => [C + R * Math.cos((i / n) * TAU - Math.PI / 2), C + R * Math.sin((i / n) * TAU - Math.PI / 2)];
  let s = '';
  for (let i = 0; i < n; i++) {
    const a = P(i), b = P((i * m) % n);
    s += seg(a[0], a[1], b[0], b[1], 0.1);
  }
  return linesG(s, 0.5);
}

// 13. Star polygon {n/k}.
function starPolygon({ n, k }) {
  const R = C * 0.92;
  const P = (i) => [C + R * Math.cos((i / n) * TAU - Math.PI / 2), C + R * Math.sin((i / n) * TAU - Math.PI / 2)];
  const pts = [];
  let i = 0;
  do { pts.push(P(i)); i = (i + k) % n; } while (i !== 0);
  pts.push(P(0));
  return linesG(polyline(pts, 0.15), 1);
}

// 14. Circle ripples — concentric circles from several centers (interference).
function ripple({ centers, step }) {
  let s = '';
  for (const [cxF, cyF] of centers) {
    const cx = cxF * S, cy = cyF * S;
    for (let r = step; r < S * 0.95; r += step) s += `<circle cx="${f(cx)}" cy="${f(cy)}" r="${f(r)}" opacity="0.07"/>`;
  }
  return `<g fill="none" stroke="${EM}" stroke-width="0.8">${s}</g>`;
}

// 15. Moiré — two square grids, one rotated.
function moire({ spacing, angle }) {
  const mk = (rot) => {
    let s = '';
    const co = Math.cos(rot), si = Math.sin(rot);
    for (let off = -900; off <= 900; off += spacing) {
      // vertical-ish then horizontal-ish lines rotated about center
      for (const [dx, dy] of [[0, 1], [1, 0]]) {
        const nx = dx ? 0 : 1, ny = dx ? 1 : 0; // unused placeholder
      }
      const p1 = rotPt(off, -950, co, si), p2 = rotPt(off, 950, co, si);
      const q1 = rotPt(-950, off, co, si), q2 = rotPt(950, off, co, si);
      s += seg(p1[0], p1[1], p2[0], p2[1], 0.06) + seg(q1[0], q1[1], q2[0], q2[1], 0.06);
    }
    return s;
  };
  function rotPt(x, y, co, si) { return [C + x * co - y * si, C + x * si + y * co]; }
  return linesG(mk(0) + mk((angle * Math.PI) / 180), 0.6);
}

// 16. Harmonograph — damped Lissajous (pendulum art).
function harmonograph({ a1, a2, a3, a4, p1, p2, damp }) {
  const R = C * 0.46;
  const pts = [];
  for (let t = 0; t < 80; t += 0.02) {
    const e1 = Math.exp(-damp * t);
    const x = R * (Math.sin(a1 * t + p1) * e1 + Math.sin(a2 * t) * e1);
    const y = R * (Math.sin(a3 * t + p2) * e1 + Math.sin(a4 * t) * e1);
    pts.push([C + x, C + y]);
  }
  return linesG(polyline(pts, 0.1), 0.6);
}

// ══ REGISTRY: build exactly 100 entries ════════════════════════════════
const reg = [];
const add = (family, name, formula, inner) => reg.push({ family, name, formula, inner });

// 1. phyllotaxis (10)
[137.5, 137.3, 137.6, 137.508, 99.5, 137.5, 137.5, 137.5, 137.4, 137.5].forEach((angle, i) => {
  const n = [1100, 900, 1200, 1400, 800, 700, 1500, 1000, 950, 1300][i];
  const c = [11, 12, 10, 9.5, 13, 14, 9, 12, 11.5, 10.5][i];
  add('Phyllotaxis', `Phyllotaxis ${angle}°`, `r=c√n, θ=n·${angle}°`, phyllotaxis({ angle, n, c }));
});
// 2. wave interference (8)
[5, 6, 7, 9, 11, 8, 10, 13].forEach((k, i) => {
  const spacing = [30, 28, 26, 24, 22, 27, 23, 20][i];
  add('Quasicrystal', `Wave interference ×${k}`, `Σ cos(k·r⊥θᵢ), ${k} angles`, waveGrating({ k, spacing }));
});
// 3. penrose dots (5)
[5, 6, 6, 7, 6].forEach((sub, i) => {
  const rot = [0, 9, 18, 0, 27][i];
  add('Penrose', `Penrose vertices · ${sub}↓`, 'P3 deflation vertices', penroseDots({ sub, rot }));
});
// 4. penrose rhombi (5)
[4, 5, 5, 6, 5].forEach((sub, i) => {
  const rot = [0, 9, 18, 0, 36][i];
  add('Penrose', `Penrose rhombi · ${sub}↓`, 'aperiodic P3 tiling', penroseRhombi({ sub, rot }));
});
// 5. log spirals (6)
[3, 5, 6, 8, 4, 7].forEach((arms, i) => {
  const b = [0.18, 0.2, 0.22, 0.15, 0.3, 0.17][i];
  add('Spiral', `Golden spiral ×${arms}`, `r=ae^{bθ}, b=${b}`, logSpiral({ arms, b }));
});
// 6. fermat (4)
[700, 900, 600, 1000].forEach((n, i) => {
  const c = [16, 14, 18, 13][i];
  add('Spiral', `Fermat spiral`, 'r=±c√θ', fermat({ n, c }));
});
// 7. archimedean (4)
[[1, 14], [3, 11], [5, 9], [6, 16]].forEach(([arms, turns]) => {
  add('Spiral', `Archimedean ×${arms}`, 'r=aθ', archimedean({ arms, turns }));
});
// 8. rose (10)
[[2, 1], [3, 1], [4, 1], [5, 1], [7, 1], [5, 2], [7, 2], [7, 3], [8, 3], [3, 2]].forEach(([num, den]) => {
  add('Rose', `Rose ${num}/${den}`, `r=cos(${num}/${den}·θ)`, rose({ num, den }));
});
// 9. maurer rose (5)
[[6, 71], [5, 97], [4, 47], [7, 19], [2, 39]].forEach(([n, d]) => {
  add('Rose', `Maurer rose n=${n}`, `walk θ=k·${d}°`, maurerRose({ n, d }));
});
// 10. lissajous (6)
[[3, 2, 90], [3, 4, 0], [5, 4, 30], [5, 6, 45], [7, 5, 90], [4, 5, 60]].forEach(([a, b, delta]) => {
  add('Lissajous', `Lissajous ${a}:${b}`, `x=sin(${a}t+δ), y=sin(${b}t)`, lissajous({ a, b, delta }));
});
// 11. spirograph (10)
[[5, 3, 5], [7, 3, 4], [8, 5, 4], [11, 4, 6], [13, 5, 7], [7, 4, 3], [9, 4, 5], [11, 7, 6], [13, 8, 5], [10, 3, 6]].forEach(([R, r, d]) => {
  add('Spirograph', `Hypotrochoid ${R}/${r}`, `R=${R}, r=${r}, d=${d}`, spirograph({ R, r, d }));
});
// 12. times-table (8)
[2, 3, 4, 5, 6, 7, 29, 51].forEach((m) => {
  const n = [200, 200, 200, 200, 200, 200, 200, 300][[2, 3, 4, 5, 6, 7, 29, 51].indexOf(m)] || 200;
  add('Modular', `Times table ×${m}`, `i → ${m}i mod ${n}`, timesTable({ n, m }));
});
// 13. star polygons (5)
[[12, 5], [7, 3], [9, 4], [11, 5], [16, 7]].forEach(([n, k]) => {
  add('Star', `Star {${n}/${k}}`, `{${n}/${k}} polygon`, starPolygon({ n, k }));
});
// 14. ripples (4)
[
  [[0.5, 0.5]], [[0.3, 0.4], [0.7, 0.6]], [[0.25, 0.3], [0.75, 0.3], [0.5, 0.8]], [[0.2, 0.5], [0.8, 0.5]],
].forEach((centers, i) => {
  add('Ripple', `Ripple ×${centers.length}`, 'concentric interference', ripple({ centers, step: [22, 26, 28, 24][i] }));
});
// 15. moiré (4)
[[34, 6], [40, 10], [28, 4], [46, 14]].forEach(([spacing, angle]) => {
  add('Moiré', `Moiré ${angle}°`, `grids Δ=${angle}°`, moire({ spacing, angle }));
});
// 16. harmonograph (remainder up to 100)
const harmSpecs = [
  { a1: 2, a2: 3, a3: 3, a4: 2, p1: 0, p2: 1.2, damp: 0.012 },
  { a1: 3, a2: 2, a3: 2, a4: 3, p1: 0.7, p2: 0, damp: 0.01 },
  { a1: 4, a2: 3, a3: 3, a4: 4, p1: 1.1, p2: 0.4, damp: 0.014 },
  { a1: 5, a2: 4, a3: 4, a4: 5, p1: 0, p2: 1.6, damp: 0.011 },
  { a1: 2, a2: 5, a3: 5, a4: 2, p1: 0.5, p2: 0.5, damp: 0.013 },
  { a1: 3, a2: 5, a3: 4, a4: 2, p1: 1.4, p2: 0.2, damp: 0.009 },
  { a1: 6, a2: 5, a3: 5, a4: 6, p1: 0, p2: 1.0, damp: 0.015 },
  { a1: 4, a2: 5, a3: 6, a4: 3, p1: 0.9, p2: 1.3, damp: 0.012 },
];
let hi = 0;
while (reg.length < 100) {
  const sp = harmSpecs[hi % harmSpecs.length];
  add('Harmonograph', `Harmonograph ${hi + 1}`, 'damped Σ sin pendulum', harmonograph(sp));
  hi++;
}
reg.length = 100;

// ── write ───────────────────────────────────────────────────────────────
rmSync(OUT_DIR, { recursive: true, force: true });
mkdirSync(OUT_DIR, { recursive: true });

const manifest = [];
let totalBytes = 0;
reg.forEach((p, i) => {
  const id = 'p' + String(i + 1).padStart(3, '0');
  const file = `/patterns/${id}.svg`;
  const out = svg(p.inner);
  totalBytes += out.length;
  writeFileSync(resolve(OUT_DIR, `${id}.svg`), out, 'utf8');
  manifest.push({ id, name: p.name, family: p.family, formula: p.formula, file });
});

const ts =
  `// AUTO-GENERATED by scripts/gen-patterns.mjs - DO NOT EDIT.\n` +
  `// 100 generative mathematical background patterns. Run \`npm run gen:patterns\`.\n\n` +
  `export type MathPattern = {\n  id: string;\n  name: string;\n  family: string;\n  formula: string;\n  file: string;\n};\n\n` +
  `export const PATTERNS: MathPattern[] = ${JSON.stringify(manifest, null, 0)};\n\n` +
  `export const PATTERN_COUNT = ${manifest.length};\n\n` +
  `export const PATTERN_FAMILIES = ${JSON.stringify([...new Set(manifest.map((m) => m.family))])};\n\n` +
  `export function patternById(id: string): MathPattern {\n  return PATTERNS.find((p) => p.id === id) ?? PATTERNS[0];\n}\n`;

mkdirSync(dirname(MANIFEST), { recursive: true });
writeFileSync(MANIFEST, ts, 'utf8');

console.log(
  `Wrote ${manifest.length} patterns to public/patterns/ ` +
    `(${(totalBytes / 1024).toFixed(0)} KB total) + lib/patterns.generated.ts\n` +
    `Families: ${[...new Set(manifest.map((m) => m.family))].join(', ')}`,
);
