// Builds internal/serve/assets/vendor.min.js — the embedded rendering
// libraries for the serve WebUI (marked + DOMPurify + highlight.js).
//
// Usage:
//   node scripts/build-serve-vendor.mjs
//
// The bundle is committed so `go build` needs no network. To upgrade a
// library, bump the version in a scratch pnpm project (see below), re-run
// this script, and commit the new vendor.min.js + vendor-LICENSES.md.
//
// Reproduce the install:
//   cd $(mktemp -d) && pnpm add marked dompurify highlight.js esbuild
//   node scripts/build-serve-vendor.mjs --node-modules "$PWD/node_modules"
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const here = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(here, '..');
const outFile = join(repoRoot, 'internal', 'serve', 'assets', 'vendor.min.js');
const entry = join(here, 'serve-vendor', 'vendor-entry.mjs');

const arg = process.argv.find((a) => a.startsWith('--node-modules='));
const nodeModules = arg ? resolve(arg.split('=')[1]) : join(repoRoot, 'node_modules');

// Load esbuild from the supplied node_modules (createRequire resolves
// upward from a synthetic path inside it).
const require = createRequire(join(nodeModules, '_vendor-probe.js'));
let esbuild;
try {
  esbuild = require('esbuild');
} catch {
  console.error('esbuild not found under ' + nodeModules);
  console.error('Run: pnpm add marked dompurify highlight.js esbuild');
  process.exit(1);
}

mkdirSync(dirname(outFile), { recursive: true });

await esbuild.build({
  entryPoints: [entry],
  bundle: true,
  minify: true,
  format: 'iife',
  globalName: 'Vendor',
  outfile: outFile,
  nodePaths: [nodeModules],
  logLevel: 'warning',
  target: 'es2020',
});

// Record the exact versions + licenses next to the bundle.
const pkg = (name) => JSON.parse(readFileSync(join(nodeModules, name, 'package.json'), 'utf8'));
const libs = ['marked', 'dompurify', 'highlight.js', 'esbuild'];
const rows = libs.map((n) => {
  const p = pkg(n);
  return `- ${p.name} ${p.version} — ${p.license || 'see package'} (${p.homepage || ''})`;
});

const sizes = Math.round(readFileSync(outFile).length / 1024);
const md = `# serve WebUI vendor bundle

\`vendor.min.js\` is a minified IIFE bundle of the rendering libraries used by
\internal/serve/index.html. It is committed so \`go build\` stays offline.
Rebuild with \`node scripts/build-serve-vendor.mjs\` (needs \`pnpm add marked
dompurify highlight.js esbuild\` in a scratch dir; pass --node-modules).

Bundle size: ${sizes} KiB.

## Versions

${rows.join('\n')}

## Licenses

- marked: MIT — https://github.com/markedjs/marked/blob/master/LICENSE.md
- dompurify: Apache-2.0 (or MPL-2.0) — https://github.com/cure53/DOMPurify/blob/main/LICENSE
- highlight.js: BSD-3-Clause — https://github.com/highlightjs/highlight.js/blob/main/LICENSE
- esbuild (build-time only, not shipped): MIT — https://github.com/evanw/esbuild/blob/main/LICENSE.md
`;
writeFileSync(join(dirname(outFile), 'vendor-LICENSES.md'), md);
console.log(`wrote ${outFile} (${sizes} KiB) + vendor-LICENSES.md`);
