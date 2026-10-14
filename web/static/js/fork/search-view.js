// Fork-owned search result views: per-source swimlanes, table layout, view
// preferences, and the keyed result index. Extends the globals defined in
// app.js (state, t(), renderBookCard, ...) and must load after it — see
// FORK:SCRIPTS in web/index.html. Blocks below are verbatim moves out of
// app.js; app.js keeps the shared entry points (renderSearchResults,
// sortResults, setSortMode, doSearch, ...) which call into these helpers.
// Some sources emit a result.source that differs from their SourceMeta name
// (e.g. Prowlarr ebooks → "torrent", AudioBookBay → "audiobook"). Map the
// result source onto the lane/meta name so swimlanes reconcile correctly.
const RESULT_SOURCE_ALIASES = {
  torrent: 'prowlarr',
  audiobook: 'audiobookbay',
};
function laneKeyFor(source) {
  return RESULT_SOURCE_ALIASES[source] || source || 'unknown';
}

// Maps the frontend search tab onto the backend SearchTab() value used in
// SourceMeta so we can filter sources to the current tab.
const SEARCH_TAB_META = { ebooks: 'main', audiobooks: 'audiobook', manga: 'manga' };

// Filters SourceMeta to the sources relevant to the current search tab.
function filterSourceMeta(meta) {
  const wantTab = SEARCH_TAB_META[state.searchTab] || 'main';
  return (meta || []).filter(s => s && s.enabled && s.search_tab === wantTab);
}

// Builds pseudo-source lanes from a flat result list (JSON fallback).
function deriveSourcesFromResults(results) {
  const seen = new Set();
  const out = [];
  for (const r of results || []) {
    const key = laneKeyFor(r.source);
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ name: key, label: laneLabel(key), enabled: true });
  }
  return out;
}

// SSE `started`: seed the lane skeleton from SourceMeta.
function handleSearchStarted(data) {
  state.sources = filterSourceMeta(data.sources);
  state.sourceStatus = new Map();
  for (const s of state.sources) {
    state.sourceStatus.set(s.name, { status: 'searching', count: 0, timeMs: null, error: '' });
  }
  applySearchResults(data.results || [], true, { rebuild: true });
}

// SSE `results`: one source responded. Update its lane status + counts.
function handleSearchResults(data) {
  // A frame may reference sources not in the initial meta (rare); make sure
  // every referenced source has a lane so nothing is dropped.
  ensureSource(data.source);
  const status = state.sourceStatus.get(laneKeyFor(data.source)) || state.sourceStatus.get(data.source);
  if (status) {
    status.timeMs = data.search_time_ms ?? status.timeMs;
    if (data.error) { status.status = 'error'; status.error = data.error; }
    else if (data.circuit_open) status.status = 'circuit';
    else if (data.source_done) status.status = 'done';
    else status.status = 'searching';
  }
  applySearchResults(data.results || [], true, { changedSource: data.source });
}

// SSE `complete`: mark every remaining lane done and do a final full render.
function handleSearchComplete(data) {
  for (const status of state.sourceStatus.values()) {
    if (status.status === 'searching') status.status = 'done';
  }
  applySearchResults(data.results || [], false, { rebuild: true });
}

// Ensures a lane/status entry exists for a source seen mid-stream.
function ensureSource(source) {
  if (!source) return;
  const key = laneKeyFor(source);
  if (state.sourceStatus.has(key) || state.sourceStatus.has(source)) return;
  state.sources.push({ name: key, label: laneLabel(key), enabled: true });
  state.sourceStatus.set(key, { status: 'searching', count: 0, timeMs: null, error: '' });
}

// Central entry point for a new batch of results (streaming frame, complete,
// or JSON fallback). Rebuilds the keyed index and renders — either fully or,
// during streaming swimlane mode, just the lane that changed.
function applySearchResults(results, searching, opts = {}) {
  state.searchResults = results || [];
  buildResultsByKey();

  const hasLanes = !!document.querySelector('#search-results .search-lane');
  const canTargetLane = opts.changedSource && !opts.rebuild &&
    state.viewGroup === 'source' && hasLanes;

  if (canTargetLane) {
    applyLaneFrame(opts.changedSource);
    refreshLaneCounts();
  } else {
    renderSearchResults();
  }
  updateSearchChrome(searching);
}

// Builds state.resultsByKey (stable key -> result) and stamps result.__key,
// replacing the old data-idx scheme. Collisions get a numeric suffix so every
// clickable element resolves to exactly one result.
function buildResultsByKey() {
  const map = new Map();
  for (const r of state.searchResults) {
    let key = getDownloadKey(r);
    if (map.has(key)) {
      let n = 2;
      while (map.has(`${key}#${n}`)) n++;
      key = `${key}#${n}`;
    }
    r.__key = key;
    map.set(key, r);
  }
  state.resultsByKey = map;
}

function updateSearchChrome(searching) {
  const statuses = [...state.sourceStatus.values()];
  const anySearching = searching && statuses.some(s => s.status === 'searching');
  document.getElementById('search-spinner').classList.toggle('hidden', !anySearching);

  const total = state.searchResults.length;
  const started = state.sources.length > 0 || total > 0;
  document.getElementById('search-toolbar').classList.toggle('hidden', !started);
  document.getElementById('search-result-count').textContent = t('n_results', {n: total});
  if (started) document.getElementById('search-empty').classList.add('hidden');

  const settled = !anySearching && !statuses.some(s => s.status === 'searching');
  const noResults = settled && total === 0;
  document.getElementById('search-no-results').classList.toggle('hidden', !noResults);
}

function resultSizeBytes(r) {
  if (typeof r.size === 'number' && r.size > 0) return r.size;
  return parseSize(r.size_human || r.sizeHuman || r.size || '');
}

function resultSizeText(r) {
  return (r.size && r.size > 0) ? formatSize(r.size) : (r.size_human || r.sizeHuman || '');
}

// ---- View preference setters (persisted) ----
function setViewLayout(layout) {
  if (layout !== 'grid' && layout !== 'list') return;
  state.viewLayout = layout;
  localStorage.setItem('librarr.search.viewLayout', layout);
  renderSearchResults();
}

function setViewGroup(group) {
  if (group !== 'flat' && group !== 'source') return;
  state.viewGroup = group;
  localStorage.setItem('librarr.search.viewGroup', group);
  renderSearchResults();
}

function toggleLane(source) {
  if (state.collapsedLanes.has(source)) state.collapsedLanes.delete(source);
  else state.collapsedLanes.add(source);
  const lane = document.querySelector(`#search-results .search-lane[data-lane="${cssEscape(source)}"]`);
  if (lane) {
    const body = lane.querySelector('[data-lane-body]');
    const chevron = lane.querySelector('[data-lane-chevron]');
    const collapsed = state.collapsedLanes.has(source);
    if (body) body.classList.toggle('hidden', collapsed);
    if (chevron) chevron.classList.toggle('rotate-180', collapsed);
  }
}

function cssEscape(s) {
  return String(s).replace(/["\\]/g, '\\$&');
}

// baseName returns the last path segment of a file path (works for POSIX
// mounts), trimming a trailing slash first. Used to show a filename when no
// title metadata is available.
function baseName(p) {
  if (!p) return '';
  return String(p).replace(/\/+$/, '').split('/').pop();
}

// ---- Lane helpers ----
function laneLabel(name) {
  const meta = state.sources.find(s => s.name === name);
  if (meta && meta.label) return meta.label;
  if (SOURCE_COLORS[name] && SOURCE_COLORS[name].label) return SOURCE_COLORS[name].label;
  return name || 'Unknown';
}

// Ordered list of lane keys: SourceMeta order first, then any extra sources
// that showed up in results but weren't in the meta.
function laneOrder() {
  const names = state.sources.map(s => s.name);
  const set = new Set(names);
  for (const r of state.searchResults) {
    const k = laneKeyFor(r.source);
    if (!set.has(k)) { set.add(k); names.push(k); }
  }
  return names;
}

function laneResults(name) {
  return sortResults(state.searchResults.filter(r => laneKeyFor(r.source) === name));
}

function laneStatus(name) {
  return state.sourceStatus.get(name) || { status: 'done', count: 0, timeMs: null };
}

// Recomputes each lane's result count into sourceStatus.
function computeLaneCounts() {
  const counts = new Map();
  for (const r of state.searchResults) {
    const k = laneKeyFor(r.source);
    counts.set(k, (counts.get(k) || 0) + 1);
  }
  for (const [name, status] of state.sourceStatus.entries()) {
    status.count = counts.get(name) || 0;
  }
}

// ---- Render router ----
function renderSearchResults() {
  hideSearchSkeleton();
  computeLaneCounts();
  document.getElementById('search-no-results').classList.add('hidden');
  syncToolbarButtons();
  const container = document.getElementById('search-results');
  if (state.viewGroup === 'source') {
    renderSwimlanes(container);
  } else {
    renderFlat(container);
  }
}

function renderFlat(container) {
  const sorted = sortResults(state.searchResults);
  container.className = 'mb-6';
  if (state.viewLayout === 'list') {
    container.innerHTML = renderResultsTable(sorted);
  } else {
    container.innerHTML = `<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">${sorted.map((r, i) => renderBookCard(r, i)).join('')}</div>`;
  }
}

function renderSwimlanes(container) {
  container.className = 'space-y-3 mb-6';
  container.innerHTML = laneOrder().map(name => renderLane(name)).join('');
}

function renderLane(name) {
  const results = laneResults(name);
  const collapsed = state.collapsedLanes.has(name) || (laneStatus(name).status !== 'searching' && results.length === 0);
  // Persist the auto-collapse decision so the chevron/state stays consistent.
  if (collapsed) state.collapsedLanes.add(name);
  return `
    <div class="search-lane bg-slate-900/40 rounded-xl border border-slate-800 overflow-hidden" data-lane="${escapeHtml(name)}">
      ${renderLaneHeader(name, results.length, collapsed)}
      <div data-lane-body class="p-3 ${collapsed ? 'hidden' : ''}">${renderLaneBody(name, results)}</div>
    </div>
  `;
}

function renderLaneHeader(name, count, collapsed) {
  const status = laneStatus(name);
  const src = SOURCE_COLORS[name] || { bg: '#475569', text: 'white' };
  const time = (status.timeMs != null && status.timeMs > 0)
    ? `<span class="text-xs text-slate-500">${(status.timeMs / 1000).toFixed(1)}s</span>` : '';
  return `
    <button data-action="toggleLane" data-arg="${escapeHtml(name)}"
      class="w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-slate-800/50 transition-colors">
      <svg data-lane-chevron class="w-4 h-4 text-slate-500 transition-transform ${collapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
      <span class="px-2 py-0.5 rounded text-xs font-medium" style="background:${src.bg};color:${src.text || 'white'}">${escapeHtml(laneLabel(name))}</span>
      ${renderLaneBadge(status, count)}
      <span class="ml-auto flex items-center gap-2">${time}</span>
    </button>
  `;
}

function renderLaneBadge(status, count) {
  switch (status.status) {
    case 'searching':
      return `<span class="text-xs text-indigo-400 flex items-center gap-1"><svg class="w-3 h-3 spin" viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="3"></circle><path class="opacity-90" fill="currentColor" d="M12 3a9 9 0 0 1 9 9h-2.5A6.5 6.5 0 0 0 12 5.5V3z"></path></svg>${t('lane_searching')}</span>`;
    case 'circuit':
      return `<span class="text-xs text-amber-400">⚠ ${t('lane_circuit')}</span>`;
    case 'error':
      return `<span class="text-xs text-rose-400" title="${escapeHtml(status.error || '')}">✗ ${t('lane_error')}</span>`;
    default:
      return `<span class="text-xs text-slate-400">${t('lane_done', {n: count})}</span>`;
  }
}

function renderLaneBody(name, results) {
  const status = laneStatus(name);
  if (results.length === 0) {
    if (status.status === 'searching') return laneSkeletonHtml();
    if (status.status === 'error') return `<p class="text-xs text-rose-400/80">${escapeHtml(status.error || t('lane_error'))}</p>`;
    return `<p class="text-xs text-slate-500">${t('lane_done', {n: 0})}</p>`;
  }
  if (state.viewLayout === 'list') return renderResultsTable(results);
  return `<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">${results.map((r, i) => renderBookCard(r, i)).join('')}</div>`;
}

function laneSkeletonHtml() {
  let html = '<div class="space-y-2">';
  for (let i = 0; i < 3; i++) {
    html += `<div class="flex items-center gap-3"><div class="skeleton h-10 w-8 rounded"></div><div class="flex-1 space-y-1.5"><div class="skeleton h-3 w-2/3 rounded"></div><div class="skeleton h-3 w-1/3 rounded"></div></div></div>`;
  }
  return html + '</div>';
}

// Targeted update of a single lane during streaming — preserves the scroll
// position and collapse state of every other lane.
function applyLaneFrame(source) {
  const name = laneKeyFor(source);
  const lane = document.querySelector(`#search-results .search-lane[data-lane="${cssEscape(name)}"]`);
  if (!lane) { renderSearchResults(); return; }
  computeLaneCounts();
  const results = laneResults(name);
  // Auto-collapse a lane once it settles with zero results (transparency: keep
  // it visible but folded). Mirrors renderLane so streaming matches the final render.
  if (laneStatus(name).status !== 'searching' && results.length === 0) {
    state.collapsedLanes.add(name);
  }
  const collapsed = state.collapsedLanes.has(name);
  const tmp = document.createElement('div');
  tmp.innerHTML = renderLaneHeader(name, results.length, collapsed);
  lane.querySelector('button[data-action="toggleLane"]').replaceWith(tmp.firstElementChild);
  const body = lane.querySelector('[data-lane-body]');
  body.classList.toggle('hidden', collapsed);
  body.innerHTML = renderLaneBody(name, results);
}

// Refreshes the count badges + response times on every lane (cross-source
// dedup can shift counts on lanes other than the one targeted this frame).
function refreshLaneCounts() {
  computeLaneCounts();
  for (const name of laneOrder()) {
    const lane = document.querySelector(`#search-results .search-lane[data-lane="${cssEscape(name)}"]`);
    if (!lane) { renderSearchResults(); return; }
    const collapsed = state.collapsedLanes.has(name);
    const tmp = document.createElement('div');
    tmp.innerHTML = renderLaneHeader(name, laneStatus(name).count, collapsed);
    lane.querySelector('button[data-action="toggleLane"]').replaceWith(tmp.firstElementChild);
  }
}

// Reflects the current view prefs + sort in the toolbar: active toggle
// highlighting, and showing sort buttons only in grid layout (list sorts by
// clicking column headers instead).
function syncToolbarButtons() {
  document.querySelectorAll('#search-toolbar [data-view-layout]').forEach(el => {
    const active = el.dataset.viewLayout === state.viewLayout;
    el.className = `seg-btn text-xs px-3 py-1 rounded-md transition-colors ${active ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-white'}`;
  });
  document.querySelectorAll('#search-toolbar [data-view-group]').forEach(el => {
    const active = el.dataset.viewGroup === state.viewGroup;
    el.className = `seg-btn text-xs px-3 py-1 rounded-md transition-colors ${active ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-white'}`;
  });
  const sortButtons = document.getElementById('search-sort-buttons');
  if (sortButtons) sortButtons.classList.toggle('hidden', state.viewLayout !== 'grid');
  document.querySelectorAll('#search-sort-buttons .sort-btn').forEach(el => {
    const active = el.dataset.sort === state.sortMode;
    el.className = `sort-btn text-xs px-3 py-1 rounded-md transition-colors ${active ? 'bg-indigo-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white'}`;
  });
}

// ---- Table rendering ----
function renderResultsTable(results) {
  return `
    <div class="overflow-x-auto">
      <table class="w-full text-sm border-collapse">
        ${renderTableHead()}
        <tbody>${results.map((r, i) => renderBookRow(r, i)).join('')}</tbody>
      </table>
    </div>
  `;
}

function renderTableHead() {
  const cols = [
    { key: 'title', label: t('col_title') },
    { key: 'author', label: t('col_author') },
    { key: 'source', label: t('col_source') },
    { key: 'format', label: t('col_format') },
    { key: 'size', label: t('col_size') },
    { key: 'seeders', label: t('col_seeders') },
    { key: 'indexer', label: t('col_indexer') },
  ];
  const arrow = key => state.sortMode === key ? (state.sortDir === 'asc' ? ' ▲' : ' ▼') : '';
  const th = cols.map(c => `
    <th data-action="setSortMode" data-arg="${c.key}"
      class="text-left font-medium text-slate-400 px-2 py-1.5 cursor-pointer hover:text-white whitespace-nowrap select-none ${state.sortMode === c.key ? 'text-white' : ''}">
      ${escapeHtml(c.label)}${arrow(c.key)}
    </th>`).join('');
  return `<thead><tr class="border-b border-slate-800">
    <th class="px-2 py-1.5"></th>${th}
    <th class="px-2 py-1.5 text-right font-medium text-slate-400 whitespace-nowrap">${escapeHtml(t('col_action'))}</th>
  </tr></thead>`;
}

function renderBookRow(result, index) {
  const coverHtml = result.cover_url
    ? `<img src="${escapeHtml(result.cover_url)}" alt="" class="h-12 w-9 object-cover rounded" loading="lazy" data-ph-title="${escapeHtml(result.title || '')}" data-ph-idx="${index}" data-ph-size="sm">`
    : makeSmallPlaceholderHtml(result.title || '?', index);
  const seeders = result.seeders ? `<span class="text-emerald-400">${escapeHtml(String(result.seeders))}</span>` : '';
  const leechers = result.leechers ? `<span class="text-amber-400"> / ${escapeHtml(String(result.leechers))}</span>` : '';
  const swarm = (seeders || leechers) ? `${seeders}${leechers}` : '<span class="text-slate-600">—</span>';
  const sizeText = resultSizeText(result);
  return `
    <tr class="border-b border-slate-800/60 hover:bg-slate-800/30">
      <td class="px-2 py-1.5">${coverHtml}</td>
      <td class="px-2 py-1.5 max-w-xs"><span class="text-white line-clamp-2" title="${escapeHtml(result.title || '')}">${escapeHtml(result.title || 'Unknown')}</span></td>
      <td class="px-2 py-1.5 text-slate-400 max-w-[10rem]"><span class="line-clamp-1">${escapeHtml(result.author || '')}</span></td>
      <td class="px-2 py-1.5 text-slate-400 whitespace-nowrap">${escapeHtml(laneLabel(laneKeyFor(result.source)))}</td>
      <td class="px-2 py-1.5 text-slate-500 uppercase whitespace-nowrap">${escapeHtml(result.format || '')}</td>
      <td class="px-2 py-1.5 text-slate-400 whitespace-nowrap">${escapeHtml(sizeText)}</td>
      <td class="px-2 py-1.5 whitespace-nowrap">${swarm}</td>
      <td class="px-2 py-1.5 text-slate-500 whitespace-nowrap">${escapeHtml(result.indexer || '')}</td>
      <td class="px-2 py-1.5 text-right">${renderDownloadButton(result, true)}</td>
    </tr>
  `;
}

// ---- Download button (shared by cards + rows) ----
function downloadButtonMeta(result) {
  const downloadKey = getDownloadKey(result);
  const isDownloading = state.pendingDownloads.has(downloadKey);
  const isTrackedAnnaDownload = hasTrackedAnnaDownload(downloadKey);
  const isTrackedDirectDownload = hasTrackedDirectDownload(downloadKey);
  const trackedJob = getTrackedDownloadJob(downloadKey);
  const downloadOutcome = state.downloadOutcomes.get(downloadKey);
  const buttonState = (isDownloading || isTrackedAnnaDownload || isTrackedDirectDownload)
    ? 'loading' : (downloadOutcome ? downloadOutcome.status : 'idle');
  const text = {
    idle: t('download'), loading: t('loading'),
    success: t('download_added'), error: t('download_failed_state'),
  };
  const displayText = buttonState === 'loading' && trackedJob?.detail
    ? escapeHtml(trackedJob.detail) : (text[buttonState] || text.idle);
  return { buttonState, displayText };
}

function renderDownloadButton(result, compact = false) {
  const { buttonState, displayText } = downloadButtonMeta(result);
  const styles = {
    idle: 'bg-indigo-600 hover:bg-indigo-500 text-white',
    loading: 'bg-indigo-500/70 text-white cursor-not-allowed',
    success: 'bg-emerald-600 hover:bg-emerald-500 text-white cursor-default',
    error: 'bg-rose-600 hover:bg-rose-500 text-white cursor-pointer',
  };
  const icon = {
    idle: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>`,
    loading: `<svg class="w-4 h-4 spin" viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2"></circle><path class="opacity-90" fill="currentColor" d="M12 3a9 9 0 0 1 9 9h-2.5A6.5 6.5 0 0 0 12 5.5V3z"></path></svg>`,
    success: `<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/></svg>`,
    error: `<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 8v4m0 4h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z"/></svg>`,
  };
  const disabled = (buttonState === 'idle' || buttonState === 'error') ? '' : 'disabled aria-busy="true"';
  const sizing = compact
    ? 'text-xs px-3 py-1 gap-1'
    : 'w-full text-sm py-1.5 gap-1.5';
  return `
    <button data-action="startDownload" data-key="${escapeHtml(result.__key || '')}" ${disabled}
      class="${sizing} ${styles[buttonState] || styles.idle} font-medium rounded-lg transition-colors inline-flex items-center justify-center disabled:opacity-100">
      ${icon[buttonState] || icon.idle}
      <span class="truncate">${displayText}</span>
    </button>`;
}

// Refreshes only the download buttons in place (no re-render) so downloads
// don't reset scroll or lane state.
function syncDownloadButtons() {
  const buttons = document.querySelectorAll('#search-results button[data-action="startDownload"]');
  buttons.forEach(btn => {
    const result = state.resultsByKey.get(btn.dataset.key);
    if (!result) return;
    const compact = btn.classList.contains('text-xs');
    const tmp = document.createElement('div');
    tmp.innerHTML = renderDownloadButton(result, compact);
    btn.replaceWith(tmp.firstElementChild);
  });
}

function makeSmallPlaceholderHtml(title, index) {
  const gradient = COVER_GRADIENTS[index % COVER_GRADIENTS.length];
  const letter = (title || '?').charAt(0).toUpperCase();
  return `<div class="h-12 w-9 rounded bg-gradient-to-br ${gradient} flex items-center justify-center text-white text-xs font-bold">${escapeHtml(letter)}</div>`;
}
