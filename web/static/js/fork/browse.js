// Fork-owned Browse tab (filesystem library browser, /api/files/*): listing,
// rename with preview, single-file rename, folder merge. Extends the globals
// in app.js (state, t(), apiJson, escapeHtml, ...) and must load after it —
// see FORK:SCRIPTS in web/index.html. Verbatim move out of app.js; app.js
// keeps only the state fields and the CLICK/CHANGE action hooks.
// ============================================================
// BROWSE TAB (filesystem library browser)
// ============================================================
// Backed by /api/files/*, which reads the disk directly — unlike the Library
// tab, nothing here goes through Kavita. One directory per request, so the UI
// never waits on a recursive walk of an NFS share.
//
// Rows are built with DOM APIs rather than HTML strings: file names come from
// the filesystem and may contain quotes, which escapeHtml() does not escape and
// which would break out of an attribute.

const BROWSE_FOLDER_ICON = '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>';
const BROWSE_FILE_ICON = '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 3h7l5 5v13a1 1 0 01-1 1H7a1 1 0 01-1-1V4a1 1 0 011-1z"/></svg>';
const BROWSE_PENCIL_ICON = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>';
const BROWSE_TAG_ICON = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5a1.99 1.99 0 011.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.99 1.99 0 013 12V7a4 4 0 014-4z"/></svg>';
const BROWSE_UP_ICON ='<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12"/></svg>';
const BROWSE_MOVE_ICON = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 11v6m0 0l-3-3m3 3l3-3"/></svg>';

// Kinds this build knows how to label; anything else falls back to "other".
const BROWSE_KINDS = new Set(['cbz', 'epub', 'pdf', 'cbr', 'audio', 'other']);

// Library names are translated from the stable root id; the API's label is only
// a fallback for an id this build does not know about.
function browseRootLabel(root) {
  const key = `browse_root_${root.id}`;
  const label = t(key);
  return label === key ? (root.label || root.id) : label;
}

async function loadBrowse() {
  hideBrowseError();
  try {
    const data = await apiJson('/api/files/roots');
    state.browseRoots = data.roots || [];
    state.browseKavita = !!data.kavita;
  } catch (err) {
    state.browseRoots = [];
    state.browseKavita = false;
    showBrowseError(t('browse_roots_failed'));
  }

  renderBrowseRoots();
  // No Kavita configured means the endpoint can only answer 503, so the button
  // is not offered at all rather than shown and failing.
  document.getElementById('browse-scan-btn').classList
    .toggle('hidden', !state.browseKavita || state.browseRoots.length === 0);
  document.getElementById('browse-no-roots').classList.toggle('hidden', state.browseRoots.length > 0);
  if (state.browseRoots.length === 0) {
    clearBrowseListing();
    return;
  }

  // Keep the folder we were in across tab switches, as long as its library is
  // still configured.
  const known = state.browseRoots.some(r => r.id === state.browseRootId);
  if (!known || !state.browsePath) {
    state.browseRootId = state.browseRoots[0].id;
    state.browsePath = state.browseRoots[0].path;
  }
  await openBrowsePath(state.browsePath);
}

function selectBrowseRoot(id) {
  const root = state.browseRoots.find(r => r.id === id);
  if (!root) return;
  state.browseRootId = root.id;
  openBrowsePath(root.path);
}

async function openBrowsePath(path) {
  hideBrowseError();
  let data;
  try {
    data = await apiJson(`/api/files/list?path=${encodeURIComponent(path)}`);
  } catch (err) {
    // A folder that vanished (renamed, deleted) must not leave the tab stuck on
    // a dead path: fall back to the library root.
    const root = state.browseRoots.find(r => r.id === state.browseRootId);
    if (root && path !== root.path) {
      state.browsePath = root.path;
      await openBrowsePath(root.path);
    }
    showBrowseError(t('browse_load_failed'));
    return;
  }

  state.browseRootId = data.root;
  state.browsePath = data.path;
  state.browseIsRoot = !!data.is_root;
  state.browseParent = data.parent || null;
  // Leaving the folder invalidates the plan that was previewed on it.
  if (state.renamePath && state.renamePath !== data.path) closeBrowseRename();
  if (state.metaPath && state.metaPath !== data.path) closeBrowseMeta();
  // Selection is per-folder — the paths it holds are not in the new listing —
  // and so is the merge panel built from it. The previous report described the
  // folder we are leaving, so it goes too; applyBrowseMerge re-renders its own
  // after this call.
  state.browseSelection = [];
  closeBrowseMerge();
  closeBrowseMove();
  closeBrowseConvert();
  hideBrowseMergeReport();
  renderBrowseSelectBar();
  // The file panel is anchored to one file, which the new listing no longer shows.
  if (state.metaFilePath) closeBrowseFileMeta();
  // A library root is never renamable, so the button has nothing to offer there.
  document.getElementById('browse-rename-btn').classList.toggle('hidden', state.browseIsRoot);
  // Same for a folder-wide metadata edit: at the root it would cover every series.
  document.getElementById('browse-meta-btn').classList.toggle('hidden', state.browseIsRoot);
  // Creating a child of the current folder is always allowed — including under
  // a root, which is the main use case. A navigation drops the open panel, like
  // the rename and metadata ones above.
  closeBrowseMkdir();
  document.getElementById('browse-mkdir-btn').classList.remove('hidden');
  renderBrowseRoots();
  renderBrowseBreadcrumb(data);
  renderBrowseEntries(data);
}

function renderBrowseRoots() {
  const wrap = document.getElementById('browse-roots');
  wrap.innerHTML = '';
  state.browseRoots.forEach(root => {
    const btn = document.createElement('button');
    btn.className = 'px-4 py-1.5 text-sm font-medium rounded-lg transition-colors whitespace-nowrap ' +
      (root.id === state.browseRootId ? 'bg-indigo-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white');
    btn.dataset.action = 'browseRoot';
    btn.dataset.arg = root.id;
    btn.textContent = browseRootLabel(root);
    wrap.appendChild(btn);
  });
}

function renderBrowseBreadcrumb(data) {
  const nav = document.getElementById('browse-breadcrumb');
  nav.innerHTML = '';
  const segments = data.segments || [];
  const rootMeta = state.browseRoots.find(r => r.id === data.root)
    || { id: data.root, label: data.root_label };
  nav.appendChild(browseCrumb(browseRootLabel(rootMeta), data.root_path, segments.length === 0));
  segments.forEach((seg, i) => {
    const sep = document.createElement('span');
    sep.className = 'text-slate-600';
    sep.textContent = '/';
    nav.appendChild(sep);
    nav.appendChild(browseCrumb(seg.name, seg.path, i === segments.length - 1));
  });
}

function browseCrumb(label, path, isCurrent) {
  if (isCurrent) {
    const span = document.createElement('span');
    span.className = 'text-slate-200 font-medium';
    span.textContent = label;
    return span;
  }
  const btn = document.createElement('button');
  btn.className = 'text-indigo-400 hover:text-indigo-300 transition-colors';
  btn.dataset.action = 'browseOpen';
  btn.dataset.path = path;
  btn.textContent = label;
  return btn;
}

function clearBrowseListing() {
  closeBrowseRename();
  closeBrowseMeta();
  closeBrowseFileMeta();
  state.browseSelection = [];
  closeBrowseMerge();
  closeBrowseMove();
  closeBrowseConvert();
  hideBrowseMergeReport();
  renderBrowseSelectBar();
  document.getElementById('browse-rename-btn').classList.add('hidden');
  document.getElementById('browse-meta-btn').classList.add('hidden');
  document.getElementById('browse-mkdir-btn').classList.add('hidden');
  closeBrowseMkdir();
  document.getElementById('browse-breadcrumb').innerHTML = '';
  document.getElementById('browse-list').innerHTML = '';
  document.getElementById('browse-empty').classList.add('hidden');
}

function renderBrowseEntries(data) {
  const list = document.getElementById('browse-list');
  list.innerHTML = '';
  const entries = data.entries || [];
  document.getElementById('browse-empty').classList.toggle('hidden', entries.length > 0);

  if (data.parent) {
    list.appendChild(browseRow({ name: '..', is_dir: true }, data.parent, true));
  }
  entries.forEach(entry => {
    list.appendChild(browseRow(entry, browseJoin(data.path, entry.name), false));
  });
}

// The API returns POSIX paths (the server runs on Linux); a child path is the
// directory plus its entry name.
function browseJoin(dir, name) {
  return dir.endsWith('/') ? dir + name : `${dir}/${name}`;
}

function browseRow(entry, targetPath, isParent) {
  const clickable = entry.is_dir;
  const row = document.createElement(clickable ? 'button' : 'div');
  row.className = 'w-full flex items-center gap-3 px-4 py-2.5 text-left'
    + (clickable ? ' hover:bg-slate-800/60 transition-colors' : '');
  if (clickable) {
    row.dataset.action = 'browseOpen';
    row.dataset.path = targetPath;
  }
  if (!isParent && !entry.is_dir) {
    // Anchors the inline rename editor to this row.
    row.dataset.filePath = targetPath;
    row.dataset.fileName = entry.name;
  }

  const icon = document.createElement('span');
  icon.className = 'shrink-0 ' + (entry.is_dir ? 'text-amber-400/80' : 'text-slate-500');
  icon.innerHTML = isParent ? BROWSE_UP_ICON : (entry.is_dir ? BROWSE_FOLDER_ICON : BROWSE_FILE_ICON);
  row.appendChild(icon);

  const name = document.createElement('span');
  name.className = 'browse-row-name flex-1 min-w-0 truncate text-sm ' + (entry.is_dir ? 'text-slate-100' : 'text-slate-300');
  name.textContent = isParent ? t('browse_parent') : entry.name;
  row.appendChild(name);

  if (isParent) return row;

  const meta = document.createElement('span');
  meta.className = 'shrink-0 flex items-center gap-3 text-xs text-slate-500';
  if (!entry.is_dir) {
    meta.appendChild(browseFileMetaButton(targetPath));
    meta.appendChild(browseRenameFileButton());
    meta.appendChild(browseMoveFileButton());
    meta.appendChild(browseKindBadge(entry));
    const size = document.createElement('span');
    size.className = 'hidden sm:inline tabular-nums';
    size.textContent = formatSize(entry.size || 0);
    meta.appendChild(size);
  }
  const modified = document.createElement('span');
  modified.className = 'hidden md:inline tabular-nums';
  modified.textContent = formatBrowseDate(entry.modified);
  meta.appendChild(modified);
  row.appendChild(meta);

  // Folders and files are tickable: folders feed the merge panel, files the
  // move panel. The bar and its buttons sort out which applies.
  return browseSelectableRow(row, targetPath, entry.is_dir, entry.kind || '');
}

// browseSelectableRow puts the tick-box beside a listing row rather than
// inside it: a folder row is a <button> that opens the folder, and a checkbox
// nested in a button would toggle on every navigation click.
function browseSelectableRow(row, path, isDir, kind) {
  const wrap = document.createElement('div');
  wrap.className = 'flex items-center gap-2 pl-4';

  const cb = document.createElement('input');
  cb.type = 'checkbox';
  cb.className = 'shrink-0 rounded border-slate-600 bg-slate-800';
  cb.dataset.actionChange = 'browseSelect';
  cb.dataset.path = path;
  cb.dataset.isdir = isDir ? '1' : '0';
  cb.dataset.kind = kind || '';
  cb.checked = state.browseSelection.includes(path);
  const hint = isDir ? t('browse_select_folder') : t('browse_select_file');
  cb.title = hint;
  cb.setAttribute('aria-label', hint);
  wrap.appendChild(cb);

  row.classList.remove('w-full');
  row.classList.add('flex-1', 'min-w-0');
  wrap.appendChild(row);
  return wrap;
}

// Per-file rename trigger: free-form naming for the one file, next to the
// folder-wide plan the Rename panel offers.
function browseRenameFileButton() {
  const btn = document.createElement('button');
  btn.className = 'text-slate-500 hover:text-indigo-300 transition-colors';
  btn.dataset.action = 'startBrowseFileRename';
  btn.title = t('browse_rename_file');
  btn.setAttribute('aria-label', t('browse_rename_file'));
  btn.innerHTML = BROWSE_PENCIL_ICON;
  return btn;
}

// Per-file move trigger: relocates the file into another folder, keeping its
// name. Folders are excluded — they relocate through merge.
function browseMoveFileButton() {
  const btn = document.createElement('button');
  btn.className = 'text-slate-500 hover:text-indigo-300 transition-colors';
  btn.dataset.action = 'startBrowseFileMove';
  btn.title = t('browse_move_file');
  btn.setAttribute('aria-label', t('browse_move_file'));
  btn.innerHTML = BROWSE_MOVE_ICON;
  return btn;
}

// Opens the per-file metadata panel. Present on every file, editable or not:
// decision 11 wants a locked format to explain itself, which it can only do
// once the panel is open.
function browseFileMetaButton(path) {
  const btn = document.createElement('button');
  btn.className = 'text-slate-500 hover:text-indigo-300 transition-colors';
  btn.dataset.action = 'openBrowseFileMeta';
  btn.dataset.path = path;
  btn.title = t('browse_meta_file');
  btn.setAttribute('aria-label', t('browse_meta_file'));
  btn.innerHTML = BROWSE_TAG_ICON;
  return btn;
}

function browseKindBadge(entry) {
  const badge = document.createElement('span');
  const kind = BROWSE_KINDS.has(entry.kind) ? entry.kind : 'other';
  // Editable formats (CBZ, EPUB) are highlighted; the others carry the
  // rename-only badge, since renaming is the only lever they leave (decision 11).
  badge.className = 'px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wide '
    + (entry.editable
      ? 'bg-indigo-500/15 text-indigo-300 border border-indigo-500/30'
      : 'bg-slate-700/40 text-slate-400 border border-slate-600/40');
  badge.textContent = t(`browse_kind_${kind}`);
  if (!entry.editable) badge.title = t('browse_meta_status_unsupported');
  return badge;
}

function formatBrowseDate(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  return isNaN(d.getTime()) ? '' : d.toLocaleDateString();
}

function showBrowseError(message) {
  const el = document.getElementById('browse-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseError() {
  document.getElementById('browse-error').classList.add('hidden');
}

// scanBrowseKavita asks Kavita to re-read the files. Manual by design: renaming
// and metadata edits never trigger it, so a batch of fixes costs one scan
// instead of one per edit.
async function scanBrowseKavita() {
  if (!state.browseRootId) return;
  const btn = document.getElementById('browse-scan-btn');
  const label = btn.querySelector('span');
  const previous = label.textContent;
  btn.disabled = true;
  label.textContent = t('browse_scan_running');
  try {
    const res = await postFiles('/api/files/kavita/scan', { root: state.browseRootId });
    // Name the libraries actually scanned: the scan is targeted now, and the
    // user needs to see that it hit the one they expected.
    const names = (res && res.libraries || []).map((l) => l.name).filter(Boolean);
    showToast(names.length
      ? `${t('browse_scan_done')} — ${names.join(', ')}`
      : t('browse_scan_done'), 'success');
  } catch (err) {
    showToast(err.message || t('browse_scan_failed'), 'error');
  } finally {
    btn.disabled = false;
    label.textContent = previous;
  }
}

// ============================================================
// BROWSE — RENAME (folder with preview, and single files)
// ============================================================
// Nothing is ever applied implicitly: the panel previews, the user reads, and
// only the Apply button writes. Every input that feeds the plan (series name,
// scope, per-file volume) re-previews, so what is on screen is always what the
// server would do — the plan is recomputed server-side on apply anyway.

// postFiles POSTs JSON and surfaces the server's error message, which the
// generic apiJson swallows in favour of the status code.
async function postFiles(path, body) {
  const resp = await api(path, { method: 'POST', body: JSON.stringify(body) });
  let data = null;
  try {
    data = await resp.json();
  } catch (err) {
    data = null;
  }
  if (!resp.ok) {
    // `detail` is the server's user-facing explanation (e.g. what Kavita
    // answered, or which paths failed to match). It is vetted server-side to
    // exclude URLs and credentials, so it is safe to show as-is, and showing
    // it is the whole point: "kavita scan failed" alone is undiagnosable.
    const base = (data && data.error) || `API error: ${resp.status}`;
    throw new Error(data && data.detail ? `${base}: ${data.detail}` : base);
  }
  return data;
}

function browseBaseName(path) {
  const parts = String(path || '').split('/').filter(Boolean);
  return parts.length ? parts[parts.length - 1] : '';
}

function openBrowseRename() {
  if (!state.browsePath || state.browseIsRoot) return;
  state.renamePath = state.browsePath;
  state.renameVolumes = {};
  state.renamePreview = null;
  // Decision 5: the folder's own name is the starting point, and it is editable.
  document.getElementById('browse-rename-series').value = browseBaseName(state.browsePath);
  document.getElementById('browse-rename-mode').value = 'folder_and_files';
  document.getElementById('browse-rename-panel').classList.remove('hidden');
  refreshBrowseRenamePreview();
}

function closeBrowseRename() {
  state.renamePath = null;
  state.renamePreview = null;
  state.renameVolumes = {};
  document.getElementById('browse-rename-panel').classList.add('hidden');
  hideBrowseRenameError();
}

async function refreshBrowseRenamePreview() {
  if (!state.renamePath) return;
  hideBrowseRenameError();
  let data;
  try {
    data = await postFiles('/api/files/rename/preview', browseRenamePayload());
  } catch (err) {
    state.renamePreview = null;
    showBrowseRenameError(err.message || t('browse_rename_preview_failed'));
    document.getElementById('browse-rename-apply').disabled = true;
    return;
  }
  state.renamePreview = data;
  renderBrowseRenamePreview(data);
}

function browseRenamePayload() {
  return {
    path: state.renamePath,
    series_name: document.getElementById('browse-rename-series').value,
    mode: document.getElementById('browse-rename-mode').value,
    volumes: state.renameVolumes,
  };
}

function renderBrowseRenamePreview(data) {
  const folder = document.getElementById('browse-rename-folder');
  folder.textContent = data.folder_to
    ? t('browse_rename_folder_to', {
      from: browseBaseName(data.folder_from),
      to: browseBaseName(data.folder_to),
    })
    : t('browse_rename_folder_same');

  const warn = document.getElementById('browse-rename-warnings');
  const warnings = data.warnings || [];
  warn.textContent = warnings.join(' · ');
  warn.classList.toggle('hidden', warnings.length === 0);

  const body = document.getElementById('browse-rename-rows');
  body.innerHTML = '';
  const files = data.files || [];
  files.forEach(row => body.appendChild(browseRenameRow(row)));
  if (files.length === 0) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan = 3;
    td.className = 'px-4 py-3 text-sm text-slate-500';
    td.textContent = t('browse_rename_empty');
    tr.appendChild(td);
    body.appendChild(tr);
  }

  const planned = files.filter(f => f.status === 'rename').length;
  const total = planned + (data.folder_to ? 1 : 0);
  document.getElementById('browse-rename-summary').textContent =
    total === 0 ? t('browse_rename_nothing') : t('browse_rename_summary', { count: planned });
  document.getElementById('browse-rename-apply').disabled = total === 0;
}

function browseRenameRow(row) {
  const tr = document.createElement('tr');

  const current = document.createElement('td');
  current.className = 'px-4 py-2 align-middle';
  const name = document.createElement('div');
  name.className = 'text-slate-300 break-all';
  name.textContent = row.name;
  current.appendChild(name);
  tr.appendChild(current);

  // The detected volume is pre-filled and editable: detection both misses
  // volumes and invents them, so the user must be able to correct it. An empty
  // field means "leave this file alone" (decision 6).
  const volCell = document.createElement('td');
  volCell.className = 'px-4 py-2 align-middle';
  const input = document.createElement('input');
  input.type = 'number';
  input.min = '1';
  input.className = 'w-20 px-2 py-1 bg-slate-800 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-indigo-500';
  input.value = row.volume ? String(row.volume) : '';
  input.placeholder = t('browse_rename_volume_placeholder');
  input.dataset.actionChange = 'browseRenameVolume';
  input.dataset.name = row.name;
  volCell.appendChild(input);
  tr.appendChild(volCell);

  const target = document.createElement('td');
  target.className = 'px-4 py-2 align-middle';
  if (row.status === 'rename') {
    const to = document.createElement('div');
    to.className = 'text-emerald-300 break-all';
    to.textContent = row.target;
    target.appendChild(to);
  } else {
    const why = document.createElement('div');
    why.className = 'text-xs text-slate-500';
    why.textContent = t(`browse_rename_status_${row.status}`);
    target.appendChild(why);
  }
  tr.appendChild(target);

  return tr;
}

async function applyBrowseRename() {
  if (!state.renamePath) return;
  const btn = document.getElementById('browse-rename-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_rename_applying');

  let data;
  try {
    data = await postFiles('/api/files/rename/apply', browseRenamePayload());
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseRenameError(err.message || t('browse_rename_failed'));
    return;
  }
  btn.textContent = label;

  if (data.errors && data.errors.length) {
    showBrowseRenameError(data.errors.join(' · '));
  }
  showToast(t('browse_rename_done', { count: data.renamed || 0 }),
    data.errors && data.errors.length ? 'error' : 'success');

  closeBrowseRename();
  // The folder may have moved: re-open where it actually is now.
  await openBrowsePath(data.path || state.browsePath);
}

function browseRenameVolume(el) {
  const raw = el.value.trim();
  const n = parseInt(raw, 10);
  if (raw === '' || !(n > 0)) {
    delete state.renameVolumes[el.dataset.name];
  } else {
    state.renameVolumes[el.dataset.name] = n;
  }
  refreshBrowseRenamePreview();
}

// --- Single file rename -------------------------------------------------
// An inline editor inside the listing row, so renaming one file needs no modal
// and no plan. The extension is not shown: the server always keeps the original
// one, and offering to edit it would suggest otherwise.
function startBrowseFileRename(el) {
  const row = el.closest('[data-file-path]');
  if (!row || row.querySelector('.browse-rename-inline')) return;
  const path = row.dataset.filePath;
  const fullName = row.dataset.fileName || browseBaseName(path);
  const dot = fullName.lastIndexOf('.');
  const stem = dot > 0 ? fullName.slice(0, dot) : fullName;

  const editor = document.createElement('div');
  editor.className = 'browse-rename-inline flex items-center gap-2 flex-1 min-w-0';

  const input = document.createElement('input');
  input.type = 'text';
  input.value = stem;
  input.className = 'flex-1 min-w-0 px-2 py-1 bg-slate-800 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-indigo-500';
  input.addEventListener('keydown', e => {
    if (e.key === 'Enter') commitBrowseFileRename(path, input.value);
    if (e.key === 'Escape') openBrowsePath(state.browsePath);
  });

  const save = document.createElement('button');
  save.className = 'px-2 py-1 text-xs bg-indigo-600 hover:bg-indigo-500 text-white rounded';
  save.textContent = t('browse_rename_file_save');
  save.dataset.action = 'commitBrowseFileRename';
  save.dataset.path = path;

  const cancel = document.createElement('button');
  cancel.className = 'px-2 py-1 text-xs bg-slate-800 hover:bg-slate-700 text-slate-300 rounded';
  cancel.textContent = t('browse_rename_file_cancel');
  cancel.dataset.action = 'cancelBrowseFileRename';

  editor.appendChild(input);
  editor.appendChild(save);
  editor.appendChild(cancel);

  // Replace the name cell for the duration of the edit.
  const nameEl = row.querySelector('.browse-row-name');
  if (nameEl) nameEl.classList.add('hidden');
  row.insertBefore(editor, nameEl ? nameEl.nextSibling : null);
  input.focus();
  input.select();
}

async function commitBrowseFileRename(path, newName) {
  try {
    await postFiles('/api/files/rename/file', { path, new_name: newName });
  } catch (err) {
    showToast(err.message || t('browse_rename_file_failed'), 'error');
    return;
  }
  showToast(t('browse_rename_file_done'), 'success');
  await openBrowsePath(state.browsePath);
}

// --- Single file move ---------------------------------------------------
// An inline destination picker inside the listing row: the parent folder (when
// there is one) plus every subfolder of the current one, which are exactly the
// tick-boxes the merge selection already renders. No new folder option here —
// creating one is the New folder button's job, one step above.
function browseMoveDestinations() {
  const dests = [];
  if (state.browseParent) dests.push({ path: state.browseParent, label: t('browse_parent') });
  document.querySelectorAll('#browse-list input[data-action-change="browseSelect"]')
    .forEach(cb => {
      if (cb.dataset.isdir === '1') dests.push({ path: cb.dataset.path, label: browseBaseName(cb.dataset.path) });
    });
  return dests;
}

function startBrowseFileMove(el) {
  const row = el.closest('[data-file-path]');
  if (!row || row.querySelector('.browse-move-inline, .browse-rename-inline')) return;
  const dests = browseMoveDestinations();
  if (!dests.length) {
    showToast(t('browse_move_no_dest'), 'error');
    return;
  }

  const editor = document.createElement('div');
  editor.className = 'browse-move-inline flex items-center gap-2 flex-1 min-w-0';

  const select = document.createElement('select');
  select.className = 'flex-1 min-w-0 px-2 py-1 bg-slate-800 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-indigo-500';
  dests.forEach(d => {
    const opt = document.createElement('option');
    opt.value = d.path;
    opt.textContent = d.label;
    select.appendChild(opt);
  });

  const save = document.createElement('button');
  save.className = 'px-2 py-1 text-xs bg-indigo-600 hover:bg-indigo-500 text-white rounded';
  save.textContent = t('browse_move_file_save');
  save.dataset.action = 'commitBrowseFileMove';
  save.dataset.path = row.dataset.filePath;

  const cancel = document.createElement('button');
  cancel.className = 'px-2 py-1 text-xs bg-slate-800 hover:bg-slate-700 text-slate-300 rounded';
  cancel.textContent = t('browse_move_file_cancel');
  cancel.dataset.action = 'cancelBrowseFileMove';

  editor.appendChild(select);
  editor.appendChild(save);
  editor.appendChild(cancel);

  // Replace the name cell for the duration of the pick.
  const nameEl = row.querySelector('.browse-row-name');
  if (nameEl) nameEl.classList.add('hidden');
  row.insertBefore(editor, nameEl ? nameEl.nextSibling : null);
  select.focus();
}

async function commitBrowseFileMove(path, dest, btn) {
  const label = btn ? btn.textContent : null;
  if (btn) {
    btn.disabled = true;
    btn.textContent = t('browse_move_file_saving');
  }
  try {
    await postFiles('/api/files/move', { path, dest });
  } catch (err) {
    if (btn) {
      btn.textContent = label;
      btn.disabled = false;
    }
    showToast(err.message || t('browse_move_file_failed'), 'error');
    return;
  }
  showToast(t('browse_move_file_done', { name: browseBaseName(dest) }), 'success');
  await openBrowsePath(state.browsePath);
}

function showBrowseRenameError(message) {
  const el = document.getElementById('browse-rename-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseRenameError() {
  document.getElementById('browse-rename-error').classList.add('hidden');
}

// ============================================================
// BROWSE — MERGE (several volume folders into one series folder)
// ============================================================
// Torrents hand out one folder per volume; Kavita wants one folder per series.
// The listing grows a tick-box on every folder, and two or more ticks unlock the
// Merge panel, which behaves like every other write here: preview, read, then an
// explicit Apply. The server re-plans from the payload on apply, so the panel is
// a view of the decision rather than the decision itself.

// browseSelected splits the ticked rows by kind: folders merge, files move.
// Read from the DOM (authoritative post-render) rather than a second state
// copy that could drift from the tick-boxes.
function browseSelected() {
  const folders = [], files = [];
  document.querySelectorAll('#browse-list input[data-action-change="browseSelect"]:checked')
    .forEach(cb => (cb.dataset.isdir === '1' ? folders : files).push(cb.dataset.path));
  return { folders, files };
}

// browseSelectedFiles lists the ticked files with their format kind, so the
// move and convert panels know what they were opened on.
function browseSelectedFiles() {
  const out = [];
  document.querySelectorAll('#browse-list input[data-action-change="browseSelect"]:checked')
    .forEach(cb => {
      if (cb.dataset.isdir !== '1') out.push({ path: cb.dataset.path, kind: cb.dataset.kind || 'other' });
    });
  return out;
}

function renderBrowseSelectBar() {
  const n = state.browseSelection.length;
  // The bar shows from the first tick so Clear is always reachable, but merging
  // needs two folders — the button says so by being inert rather than by
  // answering an error after the click.
  document.getElementById('browse-select-bar').classList.toggle('hidden', n === 0);
  const { folders, files } = browseSelected();
  document.getElementById('browse-select-count').textContent = files.length === 0
    ? t('browse_select_count', { count: n })
    : t('browse_select_count_mixed', { folders: folders.length, files: files.length });
  document.getElementById('browse-merge-btn').disabled = folders.length < 2;
  // Moving needs at least one file; folders never move (they merge).
  document.getElementById('browse-move-btn').classList.toggle('hidden', files.length === 0);
  // Converting needs at least one ticked CBR — the per-file button stays for
  // the one-at-a-time case.
  const cbrs = browseSelectedFiles().filter(f => f.kind === 'cbr');
  document.getElementById('browse-convert-btn').classList.toggle('hidden', cbrs.length === 0);
}

function browseSelect(el) {
  const path = el.dataset.path;
  const i = state.browseSelection.indexOf(path);
  if (el.checked && i < 0) state.browseSelection.push(path);
  else if (!el.checked && i >= 0) state.browseSelection.splice(i, 1);
  // An open preview was computed for the previous selection.
  if (state.mergeSources.length) closeBrowseMerge();
  if (state.moveSources.length) closeBrowseMove();
  if (state.convertSources.length) closeBrowseConvert();
  renderBrowseSelectBar();
}

function clearBrowseSelection() {
  state.browseSelection = [];
  closeBrowseMerge();
  closeBrowseMove();
  closeBrowseConvert();
  document.querySelectorAll('#browse-list input[data-action-change="browseSelect"]')
    .forEach(cb => { cb.checked = false; });
  renderBrowseSelectBar();
}

// Tokens that belong to the volume rather than to the series, so a selection of
// "Serie.Vol.01" and "Serie.Vol.02" suggests "Serie" and not "Serie Vol".
const BROWSE_MERGE_VOLUME_TOKENS = /^(?:vol|volume|tome|tomes|t|v|\d+)$/i;

// browseMergeSuggestedName pre-fills the target name with what the selected
// folder names have in common. It compares *tokens*, not characters: a raw
// character prefix of "…plage.T01…" and "…plage.T02…" stops mid-token on "T0",
// which is not a name anyone would want offered.
function browseMergeSuggestedName(paths) {
  if (!paths.length) return '';
  const tokenized = paths.map(p => browseBaseName(p).split(/[.\s_-]+/).filter(Boolean));
  let common = tokenized[0].slice();
  tokenized.slice(1).forEach(tokens => {
    let i = 0;
    while (i < common.length && i < tokens.length
      && common[i].toLowerCase() === tokens[i].toLowerCase()) i++;
    common = common.slice(0, i);
  });
  while (common.length && BROWSE_MERGE_VOLUME_TOKENS.test(common[common.length - 1])) common.pop();
  // Nothing in common (or nothing left of it) is not a reason to offer an empty
  // field: the first folder's name is at least a starting point to edit.
  return common.join(' ') || browseBaseName(paths[0]);
}

function openBrowseMerge() {
  const { folders } = browseSelected();
  if (folders.length < 2) return;
  state.mergeSources = folders;
  state.mergeVolumes = {};
  state.mergePreview = null;
  hideBrowseMergeReport();
  document.getElementById('browse-merge-series').value = browseMergeSuggestedName(state.mergeSources);
  document.getElementById('browse-merge-panel').classList.remove('hidden');
  refreshBrowseMergePreview();
}

function closeBrowseMerge() {
  state.mergeSources = [];
  state.mergeVolumes = {};
  state.mergePreview = null;
  document.getElementById('browse-merge-panel').classList.add('hidden');
  hideBrowseMergeError();
}

function browseMergePayload() {
  return {
    sources: state.mergeSources,
    series_name: document.getElementById('browse-merge-series').value,
    volumes: state.mergeVolumes,
  };
}

async function refreshBrowseMergePreview() {
  if (state.mergeSources.length < 2) return;
  hideBrowseMergeError();
  let data;
  try {
    data = await postFiles('/api/files/merge/preview', browseMergePayload());
  } catch (err) {
    state.mergePreview = null;
    showBrowseMergeError(err.message || t('browse_merge_preview_failed'));
    document.getElementById('browse-merge-apply').disabled = true;
    return;
  }
  state.mergePreview = data;
  renderBrowseMergePreview(data);
}

function renderBrowseMergePreview(data) {
  const target = document.getElementById('browse-merge-target');
  target.textContent = t(data.target_exists ? 'browse_merge_target_existing' : 'browse_merge_target_new',
    { name: browseBaseName(data.target) });

  // One line per selected folder, so what will survive the merge — and why — is
  // readable before the click, not only in the report afterwards.
  const folders = document.getElementById('browse-merge-folders');
  folders.textContent = (data.folders || []).map(f => {
    if (f.subfolders && f.subfolders.length) {
      return t('browse_merge_folder_subfolders',
        { name: f.name, count: f.files, sub: f.subfolders.join(', ') });
    }
    return t(f.is_target ? 'browse_merge_folder_target' : 'browse_merge_folder',
      { name: f.name, count: f.files });
  }).join(' · ');

  const warn = document.getElementById('browse-merge-warnings');
  const warnings = data.warnings || [];
  warn.textContent = warnings.join(' · ');
  warn.classList.toggle('hidden', warnings.length === 0);

  const body = document.getElementById('browse-merge-rows');
  body.innerHTML = '';
  const files = data.files || [];
  files.forEach(row => body.appendChild(browseMergeRow(row)));
  if (files.length === 0) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan = 3;
    td.className = 'px-4 py-3 text-sm text-slate-500';
    td.textContent = t('browse_merge_empty');
    tr.appendChild(td);
    body.appendChild(tr);
  }

  const moving = files.filter(f => f.status === 'move').length;
  document.getElementById('browse-merge-summary').textContent =
    moving === 0 ? t('browse_merge_nothing') : t('browse_merge_summary', { count: moving });
  document.getElementById('browse-merge-apply').disabled = moving === 0;
}

function browseMergeRow(row) {
  const tr = document.createElement('tr');

  const source = document.createElement('td');
  source.className = 'px-4 py-2 align-middle';
  const folder = document.createElement('div');
  folder.className = 'text-xs text-slate-500 break-all';
  folder.textContent = row.folder;
  source.appendChild(folder);
  const name = document.createElement('div');
  name.className = 'text-slate-300 break-all';
  name.textContent = row.name;
  source.appendChild(name);
  tr.appendChild(source);

  // Editable for the same reason as in the rename panel: detection both misses
  // volumes and invents them, and an empty field means "leave this file alone".
  const volCell = document.createElement('td');
  volCell.className = 'px-4 py-2 align-middle';
  const input = document.createElement('input');
  input.type = 'number';
  input.min = '1';
  input.className = 'w-20 px-2 py-1 bg-slate-800 border border-slate-700 rounded text-sm text-white focus:outline-none focus:border-indigo-500';
  input.value = row.volume ? String(row.volume) : '';
  input.placeholder = t('browse_rename_volume_placeholder');
  input.dataset.actionChange = 'browseMergeVolume';
  input.dataset.key = row.key;
  volCell.appendChild(input);
  tr.appendChild(volCell);

  const target = document.createElement('td');
  target.className = 'px-4 py-2 align-middle';
  if (row.status === 'move') {
    const to = document.createElement('div');
    to.className = 'text-emerald-300 break-all';
    to.textContent = row.target;
    target.appendChild(to);
  } else {
    if (row.target) {
      const to = document.createElement('div');
      to.className = 'text-slate-500 break-all line-through';
      to.textContent = row.target;
      target.appendChild(to);
    }
    const why = document.createElement('div');
    // A collision or a taken name is a refusal, not a detail.
    why.className = 'text-xs ' + (row.status === 'in_place' ? 'text-slate-500' : 'text-amber-300/90');
    why.textContent = t(`browse_merge_status_${row.status}`);
    target.appendChild(why);
  }
  tr.appendChild(target);

  return tr;
}

function browseMergeVolume(el) {
  const raw = el.value.trim();
  const n = parseInt(raw, 10);
  if (raw === '' || !(n > 0)) {
    delete state.mergeVolumes[el.dataset.key];
  } else {
    state.mergeVolumes[el.dataset.key] = n;
  }
  refreshBrowseMergePreview();
}

async function applyBrowseMerge() {
  if (state.mergeSources.length < 2) return;
  const btn = document.getElementById('browse-merge-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_merge_applying');

  let data;
  try {
    data = await postFiles('/api/files/merge/apply', browseMergePayload());
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseMergeError(err.message || t('browse_merge_failed'));
    return;
  }
  btn.textContent = label;

  showToast(t('browse_merge_done', { count: data.moved || 0 }),
    data.errors && data.errors.length ? 'error' : 'success');

  // The listing must show the new folder before the report describes it —
  // openBrowsePath clears the previous report, so the new one comes after.
  await openBrowsePath(data.path || state.browsePath);
  renderBrowseMergeReport(data);
}

// renderBrowseMergeReport says what moved, what was left alone, and which
// folders survived. A folder kept because of a leftover the listing hides
// (@eaDir, a dotfile) is the case this exists for: without naming the leftover,
// "the folder is still there" reads as a bug.
function renderBrowseMergeReport(data) {
  const el = document.getElementById('browse-merge-report');
  el.innerHTML = '';
  const lines = [];

  lines.push({ text: t('browse_merge_done', { count: data.moved || 0 }), tone: 'text-slate-200' });
  if (data.skipped && data.skipped.length) {
    lines.push({ text: t('browse_merge_report_skipped', { count: data.skipped.length }), tone: 'text-amber-300/90' });
    data.skipped.forEach(row => {
      lines.push({
        text: `${row.folder}/${row.name} — ${t(`browse_merge_status_${row.status}`)}`,
        tone: 'text-slate-400 pl-3',
      });
    });
  }
  if (data.removed_folders && data.removed_folders.length) {
    lines.push({
      text: t('browse_merge_report_removed', { names: data.removed_folders.join(', ') }),
      tone: 'text-slate-400',
    });
  }
  (data.kept_folders || []).forEach(k => {
    lines.push({
      text: k.remaining && k.remaining.length
        ? t('browse_merge_report_kept', { name: k.name, items: k.remaining.join(', ') })
        : t('browse_merge_report_kept_failed', { name: k.name }),
      tone: 'text-amber-300/90',
    });
  });
  (data.errors || []).forEach(e => lines.push({ text: e, tone: 'text-red-300' }));
  (data.warnings || []).forEach(w => lines.push({ text: w, tone: 'text-slate-400' }));

  // Built with DOM APIs: every name in here comes from the filesystem.
  lines.forEach(line => {
    const div = document.createElement('div');
    div.className = 'text-xs break-all ' + line.tone;
    div.textContent = line.text;
    el.appendChild(div);
  });
  el.classList.remove('hidden');
}

function hideBrowseMergeReport() {
  const el = document.getElementById('browse-merge-report');
  el.innerHTML = '';
  el.classList.add('hidden');
}

function showBrowseMergeError(message) {
  const el = document.getElementById('browse-merge-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseMergeError() {
  document.getElementById('browse-merge-error').classList.add('hidden');
}

// ============================================================
// BROWSE — MOVE (several files into one folder)
// ============================================================
// The tick-boxes grew past folders: ticked files feed this panel instead of
// the merge one. Destinations are the parent plus the current subfolders —
// same list the single-file move editor offers. Collisions are previewed
// client-side from the destination listing (an existing endpoint, no new
// round-trip type), and the server re-checks on apply, reporting per file
// instead of failing the batch.
function openBrowseMove() {
  const { files } = browseSelected();
  if (!files.length) return;
  state.moveSources = files;
  state.moveDest = null;
  hideBrowseMergeReport();
  hideBrowseMoveError();
  const dests = browseMoveDestinations();
  if (!dests.length) {
    showToast(t('browse_move_no_dest'), 'error');
    return;
  }
  const select = document.getElementById('browse-move-dest');
  select.innerHTML = '';
  dests.forEach(d => {
    const opt = document.createElement('option');
    opt.value = d.path;
    opt.textContent = d.label;
    select.appendChild(opt);
  });
  state.moveDest = select.value || null;
  document.getElementById('browse-move-panel').classList.remove('hidden');
  refreshBrowseMovePreview();
}

function closeBrowseMove() {
  state.moveSources = [];
  state.moveDest = null;
  const panel = document.getElementById('browse-move-panel');
  if (panel) panel.classList.add('hidden');
  hideBrowseMoveError();
}

function hideBrowseMoveError() {
  const el = document.getElementById('browse-move-error');
  if (el) el.classList.add('hidden');
}

function showBrowseMoveError(message) {
  const el = document.getElementById('browse-move-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

async function refreshBrowseMovePreview() {
  if (!state.moveSources.length || !state.moveDest) return;
  hideBrowseMoveError();
  let listing;
  try {
    listing = await apiJson(`/api/files/list?path=${encodeURIComponent(state.moveDest)}`);
  } catch (err) {
    showBrowseMoveError(t('browse_move_preview_failed'));
    document.getElementById('browse-move-apply').disabled = true;
    return;
  }
  document.getElementById('browse-move-target').textContent =
    t('browse_move_target', { name: browseBaseName(state.moveDest) });
  const present = new Set((listing.entries || []).filter(e => !e.is_dir).map(e => e.name));
  const body = document.getElementById('browse-move-rows');
  body.innerHTML = '';
  let movable = 0;
  state.moveSources.forEach(src => {
    const name = browseBaseName(src);
    const taken = present.has(name);
    if (!taken) movable++;
    body.appendChild(browseMoveRow(name, taken));
  });
  document.getElementById('browse-move-summary').textContent = movable === 0
    ? t('browse_move_nothing')
    : t('browse_move_summary', { count: movable });
  document.getElementById('browse-move-apply').disabled = movable === 0;
}

function browseMoveRow(name, taken) {
  const tr = document.createElement('tr');
  const from = document.createElement('td');
  from.className = 'px-4 py-2 align-middle';
  const nameEl = document.createElement('div');
  nameEl.className = 'text-slate-300 break-all';
  nameEl.textContent = name;
  from.appendChild(nameEl);
  tr.appendChild(from);
  const to = document.createElement('td');
  to.className = 'px-4 py-2 align-middle';
  const dest = document.createElement('div');
  dest.className = taken ? 'text-slate-500 break-all line-through' : 'text-emerald-300 break-all';
  dest.textContent = name;
  to.appendChild(dest);
  if (taken) {
    const why = document.createElement('div');
    why.className = 'text-xs text-amber-300/90';
    why.textContent = t('browse_move_status_exists');
    to.appendChild(why);
  }
  tr.appendChild(to);
  return tr;
}

async function applyBrowseMove() {
  if (!state.moveSources.length || !state.moveDest) return;
  const btn = document.getElementById('browse-move-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_move_applying');

  let data;
  try {
    data = await postFiles('/api/files/move', { sources: state.moveSources, dest: state.moveDest });
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseMoveError(err.message || t('browse_move_failed'));
    return;
  }
  btn.textContent = label;
  btn.disabled = false;

  showToast(t('browse_move_done', { count: data.moved ? data.moved.length : 0, name: browseBaseName(data.path) }),
    data.skipped && data.skipped.length ? 'error' : 'success');

  // The listing must refresh (and drop the selection) before the report
  // describes the new state — openBrowsePath clears the shared report area.
  await openBrowsePath(state.browsePath);
  renderBrowseMoveReport(data);
}

// renderBrowseMoveReport reuses the merge report area: one write report at a
// time, cleared on the next navigation like every other report here.
function renderBrowseMoveReport(data) {
  const el = document.getElementById('browse-merge-report');
  el.innerHTML = '';
  const lines = [];
  lines.push({ text: t('browse_move_done', { count: data.moved ? data.moved.length : 0, name: browseBaseName(data.path) }), tone: 'text-slate-200' });
  (data.skipped || []).forEach(sk => {
    lines.push({ text: `${sk.name} — ${t(`browse_move_status_${sk.reason}`)}`, tone: 'text-amber-300/90' });
  });
  lines.forEach(line => {
    const div = document.createElement('div');
    div.className = 'text-xs break-all ' + line.tone;
    div.textContent = line.text;
    el.appendChild(div);
  });
  el.classList.remove('hidden');
}

// ============================================================
// BROWSE — CONVERT (several CBRs to CBZ at once)
// ============================================================
// Same shape as the move panel: ticked CBRs, a preview of what each becomes,
// then one batch request the server reports per file. The preview needs no
// server round-trip — every target lands next to its source, so the current
// listing (already on screen) says which names are taken.
function openBrowseConvert() {
  const files = browseSelectedFiles().filter(f => f.kind === 'cbr');
  if (!files.length) return;
  state.convertSources = files.map(f => f.path);
  hideBrowseMergeReport();
  hideBrowseConvertError();

  const present = new Set();
  document.querySelectorAll('#browse-list [data-file-name]')
    .forEach(el => { if (el.dataset.fileName) present.add(el.dataset.fileName); });

  const body = document.getElementById('browse-convert-rows');
  body.innerHTML = '';
  let ready = 0;
  files.forEach(f => {
    const target = browseConvertTarget(f.path);
    const taken = present.has(target);
    if (!taken) ready++;
    body.appendChild(browseConvertRow(f.path, target, taken));
  });
  document.getElementById('browse-convert-summary').textContent = ready === 0
    ? t('browse_convert_nothing')
    : t('browse_convert_summary', { count: ready });
  document.getElementById('browse-convert-apply').disabled = ready === 0;
  document.getElementById('browse-convert-panel').classList.remove('hidden');
}

// browseConvertTarget mirrors the server's naming: same stem, .cbz extension.
function browseConvertTarget(path) {
  return browseBaseName(path).replace(/\.[^.]*$/, '') + '.cbz';
}

function closeBrowseConvert() {
  state.convertSources = [];
  const panel = document.getElementById('browse-convert-panel');
  if (panel) panel.classList.add('hidden');
  hideBrowseConvertError();
}

function hideBrowseConvertError() {
  const el = document.getElementById('browse-convert-error');
  if (el) el.classList.add('hidden');
}

function showBrowseConvertError(message) {
  const el = document.getElementById('browse-convert-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function browseConvertRow(path, target, taken) {
  const tr = document.createElement('tr');
  const from = document.createElement('td');
  from.className = 'px-4 py-2 align-middle';
  const nameEl = document.createElement('div');
  nameEl.className = 'text-slate-300 break-all';
  nameEl.textContent = browseBaseName(path);
  from.appendChild(nameEl);
  tr.appendChild(from);
  const to = document.createElement('td');
  to.className = 'px-4 py-2 align-middle';
  const dest = document.createElement('div');
  dest.className = taken ? 'text-slate-500 break-all line-through' : 'text-emerald-300 break-all';
  dest.textContent = target;
  to.appendChild(dest);
  if (taken) {
    const why = document.createElement('div');
    why.className = 'text-xs text-amber-300/90';
    why.textContent = t('browse_convert_status_exists');
    to.appendChild(why);
  }
  tr.appendChild(to);
  return tr;
}

async function applyBrowseConvert() {
  if (!state.convertSources.length) return;
  const btn = document.getElementById('browse-convert-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_convert_applying');

  let data;
  try {
    data = await postFiles('/api/files/convert', { sources: state.convertSources });
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseConvertError(err.message || t('browse_convert_failed'));
    return;
  }
  btn.textContent = label;
  btn.disabled = false;

  const done = data.converted ? data.converted.length : 0;
  const failed = data.failed ? data.failed.length : 0;
  showToast(t('browse_convert_batch_done', { count: done }), failed ? 'error' : 'success');

  await openBrowsePath(state.browsePath);
  renderBrowseConvertReport(data);
}

// renderBrowseConvertReport reuses the merge report area like the move report.
function renderBrowseConvertReport(data) {
  const el = document.getElementById('browse-merge-report');
  el.innerHTML = '';
  const lines = [];
  lines.push({ text: t('browse_convert_batch_done', { count: data.converted ? data.converted.length : 0 }), tone: 'text-slate-200' });
  (data.converted || []).forEach(c => {
    lines.push({ text: `${c.name} — ${t('browse_convert_report_converted', { count: c.files })}`, tone: 'text-slate-400 pl-3' });
  });
  (data.failed || []).forEach(f => {
    const reason = t(`browse_convert_status_${f.reason}`);
    lines.push({
      text: f.detail ? `${f.name} — ${reason}: ${f.detail}` : `${f.name} — ${reason}`,
      tone: 'text-amber-300/90',
    });
  });
  lines.forEach(line => {
    const div = document.createElement('div');
    div.className = 'text-xs break-all ' + line.tone;
    div.textContent = line.text;
    el.appendChild(div);
  });
  el.classList.remove('hidden');
}

// ============================================================
// BROWSE — MKDIR (one empty folder inside the current one)
// ============================================================
// No preview: there is nothing to plan. The toolbar button opens a small panel
// with a name field, and Create writes exactly one directory. The name is
// sanitized server-side (same SanitizeSegment as rename), so what the server
// kept is what the toast names — not necessarily what was typed.
function openBrowseMkdir() {
  if (!state.browsePath) return;
  document.getElementById('browse-mkdir-error').classList.add('hidden');
  const input = document.getElementById('browse-mkdir-name');
  input.value = '';
  document.getElementById('browse-mkdir-panel').classList.remove('hidden');
  input.focus();
}

function closeBrowseMkdir() {
  const panel = document.getElementById('browse-mkdir-panel');
  if (panel) panel.classList.add('hidden');
  const err = document.getElementById('browse-mkdir-error');
  if (err) err.classList.add('hidden');
}

function showBrowseMkdirError(message) {
  const el = document.getElementById('browse-mkdir-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

async function applyBrowseMkdir() {
  const input = document.getElementById('browse-mkdir-name');
  const name = (input.value || '').trim();
  if (!name) {
    showBrowseMkdirError(t('browse_mkdir_name_required'));
    input.focus();
    return;
  }
  const btn = document.getElementById('browse-mkdir-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_mkdir_creating');

  let data;
  try {
    data = await postFiles('/api/files/mkdir', { path: state.browsePath, name });
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseMkdirError(err.message || t('browse_mkdir_failed'));
    return;
  }
  btn.textContent = label;
  btn.disabled = false;

  showToast(t('browse_mkdir_done', { name: (data && data.name) || name }), 'success');
  closeBrowseMkdir();
  // Stay where we are: the new folder appears in the refreshed listing, and
  // the user clicks into it if that is where they wanted to go.
  await openBrowsePath(state.browsePath);
}

// Typing the series name in the Browse rename panel re-previews, debounced so it
// doesn't hammer the endpoint. Only the rows are re-rendered, so the series field
// keeps its focus and caret while the preview refreshes under it.
let browseRenameTimer = null;
document.addEventListener('input', e => {
  if (e.target.id !== 'browse-rename-series') return;
  clearTimeout(browseRenameTimer);
  browseRenameTimer = setTimeout(refreshBrowseRenamePreview, 400);
});

document.addEventListener('keydown', e => {
  if (e.target.id !== 'browse-mkdir-name') return;
  if (e.key === 'Enter') applyBrowseMkdir();
  if (e.key === 'Escape') closeBrowseMkdir();
});
