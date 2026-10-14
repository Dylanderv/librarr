// Fork-owned Prowlarr advanced search panel: per-request query override (?pq=)
// and category override (?pcat=) for the search toolbar. Loads after app.js
// (FORK:SCRIPTS); app.js calls into prowlarrAdvParams() when building search
// URLs and resetProwlarrAdvForTab() on tab switches. All state here, so
// upstream edits to the toolbar/search flow merge without touching fork lines.

// Tab-default Newznab categories, mirroring categoriesForTab in
// internal/search/prowlarr.go. Used to pre-tick the panel.
const PROWLARR_ADV_DEFAULT_CATS = { ebooks: ['7000', '7020'], audiobooks: ['3030'], manga: ['7020', '7030'] };

state.prowlarrAdv = state.prowlarrAdv || { open: false, query: '', checked: null };
try {
  const saved = JSON.parse(localStorage.getItem('prowlarrAdv') || '{}');
  if (typeof saved.open === 'boolean') state.prowlarrAdv.open = saved.open;
  if (saved.checked && typeof saved.checked === 'object') state.prowlarrAdv.checked = saved.checked;
} catch (e) { /* corrupted prefs: start closed with defaults */ }

function saveProwlarrAdv() {
  try {
    localStorage.setItem('prowlarrAdv', JSON.stringify({ open: state.prowlarrAdv.open, checked: state.prowlarrAdv.checked }));
  } catch (e) { /* private mode: prefs simply don't persist */ }
}

// Checked category IDs for a tab: saved selection, else the tab defaults.
function prowlarrAdvCheckedForTab(tab) {
  const saved = state.prowlarrAdv.checked && state.prowlarrAdv.checked[tab];
  if (Array.isArray(saved)) return saved;
  return (PROWLARR_ADV_DEFAULT_CATS[tab] || []).slice();
}

function setProwlarrAdvCheckedForTab(tab, ids) {
  state.prowlarrAdv.checked = state.prowlarrAdv.checked || {};
  state.prowlarrAdv.checked[tab] = ids;
  saveProwlarrAdv();
}

// Called from switchSearchTab (app.js): refresh the panel for the new tab.
function resetProwlarrAdvForTab(tab) {
  const panel = document.getElementById('prowlarr-adv-panel');
  const btn = document.getElementById('prowlarr-adv-toggle');
  const pq = document.getElementById('prowlarr-adv-query');
  if (pq) { pq.value = ''; pq.placeholder = t('adv_query_ph'); }
  state.prowlarrAdv.query = '';
  renderProwlarrAdvCatsBox(prowlarrAdvCheckedForTab(tab));
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
  rerunProwlarrAdvSearch();
}

function applyProwlarrAdv() {
  const pq = document.getElementById('prowlarr-adv-query');
  state.prowlarrAdv.query = pq ? pq.value.trim() : '';
  setProwlarrAdvCheckedForTab(state.searchTab, readProwlarrAdvChecked());
  rerunProwlarrAdvSearch();
}

function resetProwlarrAdvCats() {
  setProwlarrAdvCheckedForTab(state.searchTab, (PROWLARR_ADV_DEFAULT_CATS[state.searchTab] || []).slice());
  renderProwlarrAdvCatsBox(prowlarrAdvCheckedForTab(state.searchTab));
}

function rerunProwlarrAdvSearch() {
  const q = (document.getElementById('search-input').value || '').trim();
  if (q.length >= 1 && typeof doSearch === 'function') doSearch(q);
}

function readProwlarrAdvChecked() {
  const box = document.getElementById('prowlarr-adv-cats-box');
  if (!box) return prowlarrAdvCheckedForTab(state.searchTab);
  return Array.from(box.querySelectorAll('input[data-cat-id]:checked')).map(el => el.dataset.catId);
}

// Flat cache of the live tree for rendering (id -> {name, depth}).
let prowlarrAdvTree = null;

async function loadProwlarrAdvCats() {
  const badge = document.getElementById('prowlarr-adv-live');
  if (prowlarrAdvTree) {
    renderProwlarrAdvCatsBox(prowlarrAdvCheckedForTab(state.searchTab));
    if (badge) setProwlarrAdvBadge(badge, true);
    return;
  }
  try {
    const data = await apiJson('/api/prowlarr/categories');
    prowlarrAdvTree = data.categories || [];
    renderProwlarrAdvCatsBox(prowlarrAdvCheckedForTab(state.searchTab));
    if (badge) setProwlarrAdvBadge(badge, !!data.live);
  } catch (e) {
    if (badge) {
      badge.textContent = t('adv_offline');
      badge.className = 'text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-500';
    }
  }
}

function setProwlarrAdvBadge(badge, live) {
  badge.textContent = live ? t('adv_live') : t('adv_fallback');
  badge.className = 'text-xs px-2 py-0.5 rounded ' + (live ? 'bg-emerald-900 text-emerald-300' : 'bg-amber-900 text-amber-300');
}

function renderProwlarrAdvCatsBox(checked) {
  const box = document.getElementById('prowlarr-adv-cats-box');
  if (!box) return;
  const tree = prowlarrAdvTree && prowlarrAdvTree.length ? prowlarrAdvTree : fallbackProwlarrAdvTree();
  const on = new Set(checked || []);
  let html = '';
  (function walk(cats, depth) {
    for (const c of cats || []) {
      if (!c || c.id == null) continue;
      const id = String(c.id);
      const pad = depth * 16;
      html += `<label class="flex items-center gap-2 py-0.5 hover:text-white" style="padding-left:${pad}px">` +
        `<input type="checkbox" data-cat-id="${escapeHtml(id)}" class="accent-indigo-600"${on.has(id) ? ' checked' : ''}>` +
        `<span>${escapeHtml(c.name || id)}</span>` +
        `<span class="text-slate-600 text-xs">${escapeHtml(id)}</span></label>`;
      walk(c.sub, depth + 1);
    }
  })(tree, 0);
  box.innerHTML = html || `<span class="text-slate-500 text-xs">${escapeHtml(t('adv_offline'))}</span>`;
  // Persist tick changes immediately so a re-render keeps them.
  box.querySelectorAll('input[data-cat-id]').forEach(el => {
    el.addEventListener('change', () => setProwlarrAdvCheckedForTab(state.searchTab, readProwlarrAdvChecked()));
  });
}

// Static fallback mirroring DefaultProwlarrCategories (backend) for when the
// live tree hasn't loaded yet.
function fallbackProwlarrAdvTree() {
  return [
    { id: '7000', name: 'Books', sub: [{ id: '7020', name: 'EBook' }, { id: '7030', name: 'Comics & Manga' }] },
    { id: '3000', name: 'Audio', sub: [{ id: '3030', name: 'Audiobook' }] },
  ];
}

// Called from doStreamingSearch/doJsonSearch (app.js). Returns "" normally.
// When the panel is open, always sends pcat (checked IDs; empty = no filter)
// plus pq when set.
function prowlarrAdvParams(displayQuery) {
  if (!state.prowlarrAdv || !state.prowlarrAdv.open) return '';
  const pqEl = document.getElementById('prowlarr-adv-query');
  const pq = (pqEl ? pqEl.value : state.prowlarrAdv.query || '').trim();
  state.prowlarrAdv.query = pq;
  const cats = readProwlarrAdvChecked();
  setProwlarrAdvCheckedForTab(state.searchTab, cats);
  let out = '&pcat=' + encodeURIComponent(cats.join(','));
  if (pq) out += '&pq=' + encodeURIComponent(pq);
  return out;
}

// Enter in the override field applies immediately.
document.getElementById('prowlarr-adv-query').addEventListener('keydown', e => {
  if (e.key === 'Enter') { e.preventDefault(); applyProwlarrAdv(); }
});
