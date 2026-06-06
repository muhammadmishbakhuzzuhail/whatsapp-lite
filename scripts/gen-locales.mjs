// gen-locales.mjs — generate STATIC UI locale JSONs by machine-translating
// en.json → each target language via Google gtx (offline, once). Placeholders
// ({name}/%s/%n) are protected as [[i]] (gtx preserves those; it would translate
// inside {}). Batched (join values with \n, ~70/chunk) → ~few calls per language.
// Resumable: skips locales whose file already exists (use --force to overwrite).
//
// Run:  node scripts/gen-locales.mjs            (only missing)
//       node scripts/gen-locales.mjs --force    (regenerate all)
//
// Source of truth = frontend/src/lib/locales/en.json. id/en/es are hand-authored
// and never overwritten.
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import { TRANSLATE_LANGS } from "../frontend/src/lib/langs.js";

const __dir = path.dirname(fileURLToPath(import.meta.url));
const LDIR = path.join(__dir, "../frontend/src/lib/locales");
const FORCE = process.argv.includes("--force");
const SKIP = new Set(["id", "en", "es"]); // hand-authored
const en = JSON.parse(fs.readFileSync(path.join(LDIR, "en.json"), "utf8"));
const KEYS = Object.keys(en);

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const PH = /\{[^}]+\}|%[a-zA-Z]/g;

function protect(s) {
  const toks = [];
  const masked = s.replace(PH, (m) => { toks.push(m); return `[[${toks.length - 1}]]`; });
  return { masked, toks };
}
function restore(s, toks) {
  return s.replace(/\[\[(\d+)\]\]/g, (_, i) => toks[+i] ?? "");
}

async function gtx(text, tl) {
  const u = "https://translate.googleapis.com/translate_a/single?client=gtx&sl=en&tl=" +
    encodeURIComponent(tl) + "&dt=t&q=" + encodeURIComponent(text);
  for (let a = 0; a < 4; a++) {
    try {
      const r = await fetch(u, { headers: { "User-Agent": "Mozilla/5.0" } });
      if (r.status === 200) { const j = await r.json(); return j[0].map((s) => s[0]).join(""); }
    } catch (e) {}
    await sleep(500 * (a + 1));
  }
  return null; // gagal → pemanggil fallback ke en
}

async function genLang(code) {
  const masks = KEYS.map((k) => protect(en[k]));
  const out = {};
  const CHUNK = 70;
  for (let i = 0; i < KEYS.length; i += CHUNK) {
    const idx = KEYS.slice(i, i + CHUNK);
    const src = masks.slice(i, i + CHUNK);
    const joined = src.map((m) => m.masked).join("\n");
    const res = await gtx(joined, code);
    const lines = res ? res.split("\n") : null;
    idx.forEach((k, j) => {
      if (lines && lines.length === idx.length) out[k] = restore(lines[j], src[j].toks);
      else out[k] = en[k]; // fallback aman (mismatch/gagal)
    });
    await sleep(150);
  }
  return out;
}

const main = async () => {
  for (const { code } of TRANSLATE_LANGS) {
    if (SKIP.has(code)) continue;
    const file = path.join(LDIR, code + ".json");
    if (!FORCE && fs.existsSync(file)) { console.log("skip", code); continue; }
    process.stdout.write("gen " + code + " … ");
    const dict = await genLang(code);
    fs.writeFileSync(file, JSON.stringify(dict, null, 1));
    console.log("ok (" + Object.keys(dict).length + ")");
  }
  console.log("done");
};
main();
