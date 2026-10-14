// Fork-owned Prowlarr advanced search panel: per-request query override (?pq=)
// and category override (?pcat=) for the search toolbar. Loads after app.js
// (FORK:SCRIPTS); app.js calls into prowlarrAdvParams() when building search
// URLs and resetProwlarrAdvForTab() on tab switches. All state here, so
// upstream edits to the toolbar/search flow merge without touching fork lines.

// Tab-default Newznab categories, mirroring categoriesForTab in
// internal/search/prowlarr.go. Used to prefill the panel and as the
// offline fallback when /api/prowlarr/categories is unreachable.
const PROWLARR_ADV_DEFAULT_CATS = { ebooks: '7000,7020', audiobooks: '3030', manga: '7020,7030' };

state.prowlarrAdv = state.prowlarrAdv || { open: false, query: '', cats: '', catsLoaded: false };
try {
  const saved = JSON.parse(localStorage.getItem('prowlarrAdv') || '{}');
  if (typeof saved.open === 'boolean') state.prowlarrAdv.open = saved.open;
  if (typeof saved.cats === 'object' && saved.cats) state.prowlarrAdv.cats = saved.cats;
} catch (e) { /* corrupted prefs: start closed with defaults */ }

function saveProwlarrAdv() {
  try {
    localStorage.setItem('prowlarrAdv', JSON.stringify({ open: state.prowlarrAdv.open, cats: state.prowlarrAdv.cats || {} }));
  } catch (e) { /* private mode: prefs simply don't persist */ }
}

function prowlarrAdvCatsForTab(tab) {
  const perTab = (state.prowlarrAdv.cats || {})[tab];
  if (typeof perTab === 'string' && perTab.trim()) return perTab.trim();
  return PROWLARR_ADV_DEFAULT_CATS[tab] || '';
}

// Called from switchSearchTab (app.js): refresh the panel for the new tab.
function resetProwlarrAdvForTab(tab) {
  const panel = document.getElementById('prowlarr-adv-panel');
  const btn = document.getElementById('prowlarr-adv-toggle');
  const pq = document.getElementById('prowlarr-adv-query');
  const pcat = document.getElementById('prowlarr-adv-cats');
  if (pq) { pq.value = ''; pq.placeholder = t('adv_query_ph'); }
  if (pcat) pcat.value = prowlarrAdvCatsForTab(tab);
  state.prowlarrAdv.query = '';
  if (panel) panel.classList.toggle('hidden', !state.prowlarrAdv.open);
  if (btn) {
    btn.classList.toggle('bg-indigo-600', state.prowlarrAdv.open);
    btn.classList.toggle('text-white', state.prowlarrAdv.open);
  }
  if (state.prowlarrAdv.open) loadProwlarrAdvCats();
}

function toggleProwlarrAdvPanel() {
  state.prowlarrAdv.open = !state.prowlarrAdv.open;
  saveProwlarrAdv();
  resetProwlarrAdvForTab(state.searchTab);
  // Re-run the current search so the toggle takes effect immediately.
  const q = (document.getElementById('search-input').value || '').trim();
  if (q.length >= 1 && typeof doSearch === 'function') doSearch(q);
}

// Flat {id, name} list for the categories datalist. Cached after first load.
let prowlarrAdvCatOptions = null;

async function loadProwlarrAdvCats() {
  const badge = document.getElementById('prowlarr-adv-live');
  const list = document.getElementById('prowlarr-adv-cats-list');
  if (prowlarrAdvCatOptions) {
    renderProwlarrAdvCats(prowlarrAdvCatOptions, true);
    return;
  }
  try {
    const data = await apiJson('/api/prowlarr/categories');
    const flat = [];
    (function walk(cats, prefix) {
      for (const c of cats || []) {
        if (!c || c.id == null) continue;
        flat.push({ id: String(c.id), name: (prefix ? prefix + ' › ' : '') + (c.name || c.id) });
        walk(c.sub, (prefix ? prefix + ' › ' : '') + (c.name || c.id));
      }
    })(data.categories, '');
    prowlarrAdvCatOptions = flat;
    renderProwlarrAdvCats(flat, !!data.live);
  } catch (e) {
    if (badge) {
      badge.textContent = t('adv_offline');
      badge.className = 'text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-500';
    }
  }
}

function renderProwlarrAdvCats(flat, live) {
  const badge = document.getElementById('prowlarr-adv-live');
  const list = document.getElementById('prowlarr-adv-cats-list');
  if (badge) {
    badge.textContent = live ? t('adv_live') : t('adv_fallback');
    badge.className = 'text-xs px-2 py-0.5 rounded ' + (live ? 'bg-emerald-900 text-emerald-300' : 'bg-amber-900 text-amber-300');
  }
  if (list) {
    list.innerHTML = (flat || []).map(c => `<option value="${escapeHtml(c.id)}">${escapeHtml(c.name)}</option>`).join('');
  }
}

// Called from doStreamingSearch/doJsonSearch (app.js). Returns "" normally,
// or "&pq=...&pcat=..." when the panel is open with something to override.
function prowlarrAdvParams(displayQuery) {
  if (!state.prowlarrAdv || !state.prowlarrAdv.open) return '';
  const pqEl = document.getElementById('prowlarr-adv-query');
  const pcatEl = document.getElementById('prowlarr-adv-cats');
  const pq = (pqEl ? pqEl.value : state.prowlarrAdv.query || '').trim();
  const pcat = (pcatEl ? pcatEl.value : '').trim();
  state.prowlarrAdv.query = pq;
  saveProwlarrAdv();
  let out = '';
  if (pq) out += '&pq=' + encodeURIComponent(pq);
  if (pcat) out += '&pcat=' + encodeURIComponent(pcat);
  return out;
}
