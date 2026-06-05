// chatThemes.js — tema latar chat ala WhatsApp (kurasi, BUKAN wallpaper bebas).
// Tiap tema punya warna terang & gelap agar tetap enak di kedua mode app.
// `doodle:true` → tampilkan pola doodle WhatsApp di atas warna.
//
// Dipakai stores.applyChatTheme() → set CSS var --chat-bg & --chat-doodle.

export const CHAT_THEMES = [
  { id: "default", label: "Default", doodle: true,  light: "#eef1f6", dark: "#0a0f14" },
  { id: "plain",   label: "Polos",   doodle: false, light: "#eef1f6", dark: "#0a0f14" },
  { id: "sage",    label: "Sage",    doodle: false, light: "#dde7da", dark: "#11201a" },
  { id: "ocean",   label: "Ocean",   doodle: false, light: "#d7e6ee", dark: "#0d1b24" },
  { id: "lilac",   label: "Lilac",   doodle: false, light: "#e4ddee", dark: "#1a1424" },
  { id: "sand",    label: "Sand",    doodle: false, light: "#ece2d2", dark: "#221c12" },
  { id: "graphite",label: "Graphite",doodle: false, light: "#e3e5e8", dark: "#15191c" },
  { id: "dusk",    label: "Dusk",    doodle: false,
    light: "linear-gradient(135deg,#dfe9f3,#e7d8c9)",
    dark:  "linear-gradient(135deg,#0b141a,#10322a)" },
];

export function chatThemeById(id) {
  return CHAT_THEMES.find((t) => t.id === id) || CHAT_THEMES[0];
}

// Warna swatch utk pratinjau tombol (sesuai mode aktif).
export function chatThemeSwatch(id, dark) {
  const t = chatThemeById(id);
  return dark ? t.dark : t.light;
}
