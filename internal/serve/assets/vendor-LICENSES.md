# serve WebUI vendor bundle

`vendor.min.js` is a minified IIFE bundle of the rendering libraries used by
internal/serve/index.html. It is committed so `go build` stays offline.
Rebuild with `node scripts/build-serve-vendor.mjs` (needs `pnpm add marked
dompurify highlight.js esbuild` in a scratch dir; pass --node-modules).

Bundle size: 227 KiB.

## Versions

- marked 18.0.7 — MIT (https://marked.js.org)
- dompurify 3.4.12 — (MPL-2.0 OR Apache-2.0) (https://github.com/cure53/DOMPurify)
- highlight.js 11.11.1 — BSD-3-Clause (https://highlightjs.org/)
- esbuild 0.28.1 — MIT ()

## Licenses

- marked: MIT — https://github.com/markedjs/marked/blob/master/LICENSE.md
- dompurify: Apache-2.0 (or MPL-2.0) — https://github.com/cure53/DOMPurify/blob/main/LICENSE
- highlight.js: BSD-3-Clause — https://github.com/highlightjs/highlight.js/blob/main/LICENSE
- esbuild (build-time only, not shipped): MIT — https://github.com/evanw/esbuild/blob/main/LICENSE.md
