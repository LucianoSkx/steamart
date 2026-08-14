"use strict";

const $ = (s, r = document) => r.querySelector(s);
const el = (tag, props = {}, kids = []) => {
  const n = document.createElement(tag);
  for (const [k, v] of Object.entries(props)) {
    if (k === "class") n.className = v;
    else if (k === "html") n.innerHTML = v;
    else if (k === "text") n.textContent = v;
    else if (k.startsWith("on") && typeof v === "function") n.addEventListener(k.slice(2), v);
    else if (v !== null && v !== undefined) n.setAttribute(k, v);
  }
  (Array.isArray(kids) ? kids : [kids]).forEach((c) => c && n.appendChild(typeof c === "string" ? document.createTextNode(c) : c));
  return n;
};

async function api(path, opts) {
  const r = await fetch(path, opts);
  if (!r.ok) throw new Error((await r.text()) || r.status);
  return r.json();
}

const ASSETS = [
  { key: "grid", label: "grid" },
  { key: "hero", label: "hero" },
  { key: "logo", label: "logo" },
  { key: "icon", label: "icon" },
  { key: "capsule", label: "capa horiz." },
];
let sgdbKey = "";
let shortcuts = [];
let busy = false;
let lastSig = "";

async function load() {
  try {
    const st = await api("/api/status");
    $("#status").textContent = `Steam: ${st.steam_root ? "ok" : "não encontrado"}`;
  } catch (e) {
    $("#status").textContent = "erro: " + e.message;
  }
  try {
    const sk = await api("/api/sgdb/status");
    sgdbKey = sk.configured ? "***" : "";
    $("#sgdb-status").textContent = sk.configured ? "configurada" : "sem key";
    $("#sgdb-status").className = "badge " + (sk.configured ? "ok" : "");
    if (sk.configured) $("#sgdb-key").value = "********";
  } catch {}
  renderList();
}

async function renderList() {
  const list = $("#list");
  list.innerHTML = "";
  try {
    shortcuts = await api("/api/shortcuts");
  } catch (e) {
    list.appendChild(el("div", { class: "card" }, [el("div", { class: "body", text: "erro: " + e.message })]));
    return;
  }
  if (!shortcuts.length) {
    list.appendChild(el("div", { class: "card" }, [el("div", { class: "body", text: "Nenhum atalho não-Steam encontrado." })]));
    return;
  }
  for (const sc of shortcuts) list.appendChild(card(sc));
  lastSig = sigOf(shortcuts);
  updateAutoBtn();
}

function sigOf(list) {
  return list.map((s) => s.appid + ":" + s.appname + ":" + (s.match ? s.match.steam_appid : 0) + ":" + Object.keys(s.artwork || {}).length).join("|");
}

function updateAutoBtn() {
  const pending = shortcuts.filter((s) => !s.match || Object.keys(s.artwork || {}).length === 0).length;
  const b = $("#auto-all");
  if (b) b.textContent = `Auto-match tudo (${pending})`;
}

function card(sc) {
  const appid = sc.appid;
  const artThumbs = el("div", { class: "art" });
  const meta = sc.match
    ? el("div", { class: "meta-line", text: `Metadados: ${sc.match.name} (app ${sc.match.steam_appid})` })
    : el("div", { class: "meta-line", text: "Sem metadados Steam aplicados" });
  const sgdbPanel = el("div", { class: "sgdb-panel hidden" });
  const c = el("div", { class: "card", "data-appid": appid }, [
    el("div", { class: "body" }, [
      el("h3", { text: sc.appname || "(sem nome)" }),
      el("div", { class: "exe", text: sc.exe || "" }),
      meta,
      artThumbs,
      el("div", { class: "actions" }, [
        el("button", { text: "Auto", onclick: () => autoMatch(appid, sc.appname) }),
        el("button", { class: "alt", text: "Buscar Steam", onclick: (e) => toggleSearch(e.currentTarget, appid) }),
        el("button", { class: "alt", text: "SteamGridDB", onclick: () => toggleSGDB(sgdbPanel, appid) }),
        el("button", { class: "alt", text: "Detalhes", onclick: () => showMeta(appid, sc.match) }),
        el("button", { class: "alt", text: "Remover arte", onclick: () => removeArt(appid) }),
      ]),
      sgdbPanel,
    ]),
  ]);
  fillArt(artThumbs, appid);
  return c;
}

async function fillArt(box, appid) {
  box.innerHTML = "";
  try {
    const arts = await api("/api/art?appid=" + appid);
    if (!arts.length) {
      box.appendChild(el("span", { class: "meta-line", text: "sem arte na grid" }));
      return;
    }
    for (const a of arts) box.appendChild(el("img", { src: a.url, title: a.type }));
  } catch {}
}

function toggleSearch(btn, appid) {
  const body = btn.closest(".body");
  const existing = $(".search-row", body);
  if (existing) { existing.remove(); return; }
  const results = el("div", { class: "results" });
  const row = el("div", { class: "search-row" }, [
    el("input", { placeholder: "nome do jogo…", onkeydown: (e) => { if (e.key === "Enter") runSearch(e.currentTarget.value, appid, results); } }),
    el("button", { text: "Buscar", onclick: () => runSearch($("input", row).value, appid, results) }),
  ]);
  row.after(results);
  body.insertBefore(row, btn);
}

async function runSearch(q, appid, results) {
  results.innerHTML = "buscando…";
  try {
    const res = await api("/api/search?q=" + encodeURIComponent(q));
    results.innerHTML = "";
    if (!res.length) results.appendChild(el("div", { class: "result", text: "nada encontrado" }));
    for (const it of res) {
      results.appendChild(el("div", { class: "result", text: `${it.name} (${it.appid})`, onclick: () => applySteam(appid, it.appid, it.name) }));
    }
  } catch (e) {
    results.textContent = "erro: " + e.message;
  }
}

async function autoMatch(appid, name) {
  try {
    const r = await api("/api/automatch?appid=" + appid);
    if (!r.found) { alert("Auto-match não encontrou '" + name + "'. Use Buscar Steam."); return; }
    await applySteam(appid, r.result.id, r.result.name);
  } catch (e) {
    alert("erro: " + e.message);
  }
}

async function applySteam(appid, steamAppid, name) {
  await api("/api/match", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ shortcut_appid: appid, steam_appid: steamAppid, name, pinned: false }),
  });
  renderList();
}

async function removeArt(appid) {
  if (!confirm("Remover toda a arte (capa/hero/logo/ícone) deste atalho?")) return;
  try {
    const r = await api("/api/remove", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ shortcut_appid: appid }),
    });
    renderList();
    alert((r.removed || 0) + " arquivo(s) removido(s).");
  } catch (e) {
    alert("erro: " + e.message);
  }
}

async function showMeta(appid, match) {
  if (!match) { alert("Faça o match com a Steam primeiro (Auto ou Buscar Steam)."); return; }
  const m = $("#meta-modal");
  const body = $("#meta-body");
  m.classList.remove("hidden");
  body.innerHTML = "carregando…";
  try {
    const d = await api("/api/meta?appid=" + match.steam_appid);
    const shots = (d.screenshots || []).slice(0, 6)
      .map((s) => `<img class="shot" src="${s}" loading="lazy" />`).join("");
    const tags = [...(d.genres || []), ...(d.categories || [])].map((t) => `<span class="tag">${t}</span>`).join(" ");
    const deck = d.deck_text ? `<div class="meta-line">Steam Deck: ${d.deck_text}</div>` : "";
    body.innerHTML = `
      <h2>${d.name}</h2>
      ${d.header_image ? `<img class="shot" src="${d.header_image}" />` : ""}
      <div class="meta-line">${[d.type, (d.developers || []).join(", "), d.release_date].filter(Boolean).join(" · ")}</div>
      <div>${tags}</div>
      ${deck}
      <p>${d.short_desc || ""}</p>
      ${shots}`;
  } catch (e) {
    body.textContent = "erro: " + e.message;
  }
}

function toggleSGDB(panel, appid) {
  if (!panel.classList.contains("hidden")) { panel.classList.add("hidden"); return; }
  panel.classList.remove("hidden");
  panel.innerHTML = "";
  const animated = el("input", { type: "checkbox", id: "sgdb-anim-" + appid });
  panel.appendChild(el("div", { class: "search-row" }, [
    el("input", { id: "sgdb-q-" + appid, placeholder: "buscar jogo na SteamGridDB…", onkeydown: (e) => { if (e.key === "Enter") sgdbSearch(e.currentTarget.value, appid, gameList, animated); } }),
    el("button", { text: "Buscar", onclick: () => sgdbSearch($("#sgdb-q-" + appid).value, appid, gameList, animated) }),
  ]));
  panel.appendChild(el("label", { class: "meta-line" }, [animated, el("span", { text: " só animadas (webp)" })]));
  const gameList = el("div", { class: "results" });
  const tabs = el("div", { class: "tabs" });
  const imgGrid = el("div", { class: "img-grid" });
  panel.append(gameList, tabs, imgGrid);
}

async function sgdbSearch(q, appid, gameList, animated) {
  if (!sgdbKey) { alert("Configure a SteamGridDB API key no topo."); return; }
  gameList.innerHTML = "buscando…";
  try {
    const games = await api("/api/sgdb/search?q=" + encodeURIComponent(q));
    gameList.innerHTML = "";
    if (!games.length) gameList.appendChild(el("div", { class: "result", text: "nada encontrado" }));
    games.forEach((g) => {
      gameList.appendChild(el("div", { class: "result", text: g.name, onclick: () => sgdbSelect(g.id, appid, tabs, imgGrid, animated) }));
    });
  } catch (e) {
    gameList.textContent = "erro: " + e.message;
  }
}

async function sgdbSelect(gameID, appid, tabs, imgGrid, animated) {
  tabs.innerHTML = "";
  const load = async (asset) => {
    [...tabs.children].forEach((b) => b.classList.toggle("active", b.dataset.asset === asset));
    imgGrid.innerHTML = "carregando…";
    try {
      let url = `/api/sgdb/images?game_id=${gameID}&asset=${asset}`;
      if (animated && animated.checked) url += "&animated=true";
      const imgs = await api(url);
      imgGrid.innerHTML = "";
      if (!imgs.length) imgGrid.appendChild(el("div", { class: "result", text: "sem imagens" }));
      imgs.forEach((im) => {
        imgGrid.appendChild(el("img", { src: im.thumb || im.url, title: (im.author && im.author.name) || im.author, onclick: () => sgdbApply(appid, asset === "icon" ? (im.thumb || im.url) : im.url, asset) }));
      });
    } catch (e) {
      imgGrid.textContent = "erro: " + e.message;
    }
  };
  ASSETS.forEach((a) => tabs.appendChild(el("button", { "data-asset": a.key, text: a.label, onclick: () => load(a.key) })));
  await load("grid");
}

async function sgdbApply(appid, url, asset) {
  try {
    await api("/api/sgdb/apply", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ shortcut_appid: appid, url, asset }),
    });
    renderList();
  } catch (e) {
    alert("erro: " + e.message);
  }
}

async function autoMatchAll() {
  if (busy) return;
  busy = true;
  const btn = $("#auto-all");
  const prev = btn.textContent;
  btn.disabled = true;
  btn.textContent = "processando…";
  try {
    const r = await api("/api/automatch-all", { method: "POST" });
    renderList();
    alert(`Auto-match tudo concluído:\nAplicados: ${r.applied}\nPulados (já tinham arte): ${r.skipped}\nSem correspondência: ${r.no_match}\nErros: ${r.errors.length}`);
  } catch (e) {
    alert("erro: " + e.message);
  } finally {
    btn.disabled = false;
    btn.textContent = prev;
    busy = false;
  }
}

$("#sgdb-save").addEventListener("click", async () => {
  const k = $("#sgdb-key").value.trim();
  if (!k || k === "********") { alert("informe a key"); return; }
  try {
    const r = await api("/api/sgdb/key", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ key: k }) });
    sgdbKey = r.configured ? "***" : "";
    $("#sgdb-status").textContent = r.configured ? "configurada" : "sem key";
    $("#sgdb-status").className = "badge " + (r.configured ? "ok" : "");
    $("#sgdb-key").value = "********";
    load();
  } catch (e) { alert("erro: " + e.message); }
});

$("#auto-all").addEventListener("click", autoMatchAll);
$("#refresh").addEventListener("click", renderList);
$("#logs-btn").addEventListener("click", async () => {
  const m = $("#log-modal"); m.classList.remove("hidden");
  try { const l = await api("/api/logs"); $("#log-body").textContent = l.join("\n"); } catch (e) { $("#log-body").textContent = e.message; }
});
document.querySelectorAll("[data-close]").forEach((b) => b.addEventListener("click", (e) => e.target.closest(".modal").classList.add("hidden")));

// Auto-refresh: re-lê os atalhos periodicamente sem atrapalhar painéis abertos.
setInterval(async () => {
  if (busy) return;
  if ($(".sgdb-panel:not(.hidden)") || !$("#meta-modal").classList.contains("hidden")) return;
  try {
    const list = await api("/api/shortcuts");
    const s = sigOf(list);
    if (s !== lastSig) { shortcuts = list; renderList(); }
  } catch {}
}, 15000);

load();
