// i18n berbasis store. Bahasa UI = file statis di ./locales/<code>.json.
// id/en/es ditulis tangan (kualitas); sisanya di-generate dari en.json
// (lihat scripts/gen-locales.mjs) → ~70 bahasa. Locale non-base di-LAZY-load
// saat dipilih; kunci yg hilang fallback ke en.
// Pakai di komponen: import { t } from "../i18n.js"; lalu {$t('key')} atau {$t('key',{n:2})}.
import { writable, derived, get } from "svelte/store";
import en from "./locales/en.json";
import id from "./locales/id.json";
import es from "./locales/es.json";
import { TRANSLATE_LANGS } from "./langs.js";

// Base hand-authored — selalu termuat + sumber fallback.
const base = { en, id, es };
// Importer lazy utk SEMUA locale JSON (Vite glob; eager=false → di-load on demand).
const loaders = import.meta.glob("./locales/*.json");
const codeOf = (p) => p.slice(p.lastIndexOf("/") + 1, -5); // ./locales/zh-CN.json → zh-CN
const available = new Set(Object.keys(loaders).map(codeOf));

// Daftar bahasa UI = tiap locale yg punya file, label = nama native (dari langs).
const labelOf = (c) => (TRANSLATE_LANGS.find((l) => l.code === c) || {}).name || c;
export const languages = TRANSLATE_LANGS
  .filter((l) => available.has(l.code) || base[l.code])
  .map((l) => ({ code: l.code, label: l.name || labelOf(l.code) }));

// Dict termuat (base + locale lazy yg sudah di-fetch).
const loaded = writable({ ...base });

async function ensure(code) {
  if (base[code]) return;                 // sudah ada
  if (get(loaded)[code]) return;          // sudah di-load
  const imp = loaders["./locales/" + code + ".json"];
  if (!imp) return;                       // tak ada file → biarkan fallback en
  try {
    const mod = await imp();
    loaded.update((d) => ({ ...d, [code]: mod.default || mod }));
  } catch (e) {}
}

const params = new URLSearchParams(location.search);
function detect() {
  let stored = null;
  try { stored = localStorage.getItem("wa-lang"); } catch (e) {}
  const c = params.get("lang") || stored || "id";
  return (base[c] || available.has(c)) ? c : "id";
}

export const locale = writable(detect());
locale.subscribe((v) => {
  try { localStorage.setItem("wa-lang", v); } catch (e) {}
  ensure(v); // muat locale aktif (kalau non-base)
});

// $t('key', {vars}) — tabel locale aktif → fallback en → key mentah.
export const t = derived([locale, loaded], ([$l, $d]) => {
  const table = $d[$l] || base.en;
  return (key, vars) => {
    let s = table[key] ?? base.en[key] ?? key;
    if (vars) for (const k in vars) s = s.split(`{${k}}`).join(vars[k]);
    return s;
  };
});
