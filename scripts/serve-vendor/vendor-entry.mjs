// Vendor entry for the serve WebUI's embedded rendering libraries.
// Bundled by scripts/build-serve-vendor.mjs into internal/serve/assets/vendor.min.js.
// Exposes the three libraries on window.Vendor for internal/serve/index.html:
//   Vendor.marked    — markdown parsing (GFM)
//   Vendor.DOMPurify — HTML sanitization
//   Vendor.hljs      — code highlighting (common language set)
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import hljs from 'highlight.js/lib/common';

export { marked, DOMPurify, hljs };
