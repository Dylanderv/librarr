// Fork-owned metadata editing for the Browse tab: folder propagation panel,
// single-file panel, and online lookup (AniList / Open Library). Uses the
// browse helpers in fork/browse.js (postFiles, browseBaseName) plus the app.js
// globals, so it must load after both — see FORK:SCRIPTS in web/index.html.
// Verbatim move out of app.js.
// ============================================================
// BROWSE — METADATA (folder propagation and single file)
// ============================================================
// Two panels, two different contracts.
//
// The folder panel only offers the fields that describe the *work* — a series,
// its publisher, its genres. Title, volume number and ISBN identify one volume,
// and writing the same one into every file of a folder produces a series whose
// volumes all claim to be the same book. The server refuses them on a folder
// too: a guard that lives only here is not a guard.
//
// Nothing is applied implicitly. The folder panel previews first, the preview
// shows the value each field currently holds next to the one that would replace
// it (decision 4: we overwrite, but never without showing), and Apply is a
// separate click. Editing any input drops the preview, so Apply can never send
// a plan the user is no longer looking at.
//
// The file panel is pre-filled from the file itself, so what would be
// overwritten is already on screen; it saves directly.

// Fields that describe the work and therefore propagate over a folder.
const BROWSE_META_FOLDER_FIELDS = ['series', 'summary', 'publisher', 'language', 'authors', 'genres', 'tags'];
// Everything the single-file panel offers, per-volume fields included.
const BROWSE_META_FILE_FIELDS = ['title', 'series', 'series_index', 'summary', 'publisher', 'language', 'isbn', 'authors', 'genres', 'tags'];
const BROWSE_META_LIST_FIELDS = new Set(['authors', 'genres', 'tags']);
const BROWSE_META_LONG_FIELDS = new Set(['summary']);

// buildMetaField renders one labelled input. withToggle adds the checkbox that
// arms the field: the API distinguishes "leave this alone" (key absent) from
// "set it to empty" (key present, empty), and only an explicit tick can express
// the difference.
function buildMetaField(field, idPrefix, withToggle) {
  const wrap = document.createElement('label');
  wrap.className = 'block';

  const head = document.createElement('span');
  head.className = 'flex items-center gap-2 mb-1';
  if (withToggle) {
    const cb = document.createElement('input');
    cb.type = 'checkbox';
    cb.id = `${idPrefix}-en-${field}`;
    cb.className = 'rounded border-slate-600 bg-slate-800';
    cb.dataset.actionChange = 'browseMetaToggle';
    cb.dataset.field = field;
    head.appendChild(cb);
  }
  const label = document.createElement('span');
  label.className = 'text-xs text-slate-400';
  label.textContent = t(`browse_meta_f_${field}`)
    + (BROWSE_META_LIST_FIELDS.has(field) ? ` (${t('browse_meta_list_hint')})` : '');
  head.appendChild(label);
  wrap.appendChild(head);

  const input = document.createElement(BROWSE_META_LONG_FIELDS.has(field) ? 'textarea' : 'input');
  if (input.tagName === 'TEXTAREA') input.rows = 3;
  else input.type = 'text';
  input.id = `${idPrefix}-${field}`;
  input.className = 'w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed';
  if (withToggle) {
    // Folder panel only: editing a value invalidates the preview it was
    // measured against. The file panel has no preview to drop.
    input.dataset.actionChange = 'browseMetaDirty';
    input.disabled = true;
  }
  wrap.appendChild(input);

  return wrap;
}

function browseMetaValue(idPrefix, field) {
  const el = document.getElementById(`${idPrefix}-${field}`);
  return el ? el.value : '';
}

// browseMetaPayloadFields turns the panel into the API's `fields` object. Only
// the fields it returns are touched; the rest are left exactly as they are.
function browseMetaPayloadFields(idPrefix, fields, isEnabled) {
  const out = {};
  fields.forEach(field => {
    if (!isEnabled(field)) return;
    const raw = browseMetaValue(idPrefix, field);
    if (BROWSE_META_LIST_FIELDS.has(field)) {
      out[field] = raw.split(',').map(v => v.trim()).filter(Boolean);
    } else {
      out[field] = raw;
    }
  });
  return out;
}

// renderMetaNotices shows the informational side of a response. Notices are
// never errors: "12 files are already up to date" is an answer, not a fault,
// and the two must not share a red box.
function renderMetaNotices(el, notices) {
  const lines = (notices || []).map(n => t(`browse_meta_notice_${n.key}`, { count: n.count }));
  el.textContent = lines.join(' · ');
  el.classList.toggle('hidden', lines.length === 0);
}

// --- Folder panel --------------------------------------------------------

function openBrowseMeta() {
  if (!state.browsePath || state.browseIsRoot) return;
  state.metaPath = state.browsePath;
  state.metaPreview = null;

  const host = document.getElementById('browse-meta-fields');
  host.innerHTML = '';
  BROWSE_META_FOLDER_FIELDS.forEach(field => {
    host.appendChild(buildMetaField(field, 'browse-meta', true));
  });
  document.getElementById('browse-meta-recursive').checked = false;
  document.getElementById('browse-meta-append').checked = false;

  // Online lookup: source routed from the library we are browsing, query
  // seeded with the folder name and editable before searching.
  document.getElementById('browse-online-source').value =
    browseOnlineDefaultType(state.browseRootId);
  document.getElementById('browse-online-query').value = browseBaseName(state.browsePath);
  resetBrowseOnline();

  document.getElementById('browse-meta-panel').classList.remove('hidden');
  resetBrowseMetaPreview();
}

function closeBrowseMeta() {
  state.metaPath = null;
  state.metaPreview = null;
  const panel = document.getElementById('browse-meta-panel');
  if (panel) panel.classList.add('hidden');
  hideBrowseMetaError();
  resetBrowseOnline();
}

function browseMetaEnabled(field) {
  const cb = document.getElementById(`browse-meta-en-${field}`);
  return !!(cb && cb.checked);
}

// Any change to the form invalidates the preview: the Apply button must never
// be armed against a plan that is no longer on screen.
function resetBrowseMetaPreview() {
  state.metaPreview = null;
  document.getElementById('browse-meta-preview').classList.add('hidden');
  document.getElementById('browse-meta-rows').innerHTML = '';
  document.getElementById('browse-meta-notices').classList.add('hidden');
  document.getElementById('browse-meta-summary').textContent = '';
  document.getElementById('browse-meta-apply').disabled = true;
}

function browseMetaToggle(el) {
  const input = document.getElementById(`browse-meta-${el.dataset.field}`);
  if (input) {
    input.disabled = !el.checked;
    if (el.checked) input.focus();
  }
  resetBrowseMetaPreview();
}

function browseMetaPayload() {
  return {
    path: state.metaPath,
    recursive: document.getElementById('browse-meta-recursive').checked,
    mode: document.getElementById('browse-meta-append').checked ? 'append' : 'set',
    fields: browseMetaPayloadFields('browse-meta', BROWSE_META_FOLDER_FIELDS, browseMetaEnabled),
  };
}

async function previewBrowseMeta() {
  if (!state.metaPath) return;
  hideBrowseMetaError();
  const payload = browseMetaPayload();
  if (Object.keys(payload.fields).length === 0) {
    showBrowseMetaError(t('browse_meta_no_fields'));
    return;
  }

  let data;
  try {
    data = await postFiles('/api/files/metadata/preview', payload);
  } catch (err) {
    resetBrowseMetaPreview();
    showBrowseMetaError(err.message || t('browse_meta_preview_failed'));
    return;
  }
  state.metaPreview = data;
  renderBrowseMetaPreview(data);
}

function renderBrowseMetaPreview(data) {
  const body = document.getElementById('browse-meta-rows');
  body.innerHTML = '';
  (data.files || []).forEach(file => browseMetaRows(file).forEach(tr => body.appendChild(tr)));
  document.getElementById('browse-meta-preview').classList.remove('hidden');

  renderMetaNotices(document.getElementById('browse-meta-notices'), data.notices);

  const willWrite = (data.files || []).filter(f => f.status === 'edit').length;
  document.getElementById('browse-meta-summary').textContent =
    willWrite === 0 ? t('browse_meta_none') : t('browse_meta_summary', { count: willWrite });
  document.getElementById('browse-meta-apply').disabled = willWrite === 0;
}

// One <tr> per changed field, plus a single explanatory row for files that hold
// no writable metadata — the reason is what makes the row worth showing.
function browseMetaRows(file) {
  const rows = [];
  if (file.status !== 'edit') {
    rows.push(browseMetaStatusRow(file));
    return rows;
  }
  const changes = (file.changes || []).filter(c => c.changed);
  if (changes.length === 0) {
    rows.push(browseMetaStatusRow(file));
    return rows;
  }
  changes.forEach((change, i) => {
    const tr = document.createElement('tr');
    tr.appendChild(browseMetaCell(i === 0 ? file.name : '', 'text-slate-300 break-all'));
    tr.appendChild(browseMetaCell(t(`browse_meta_f_${change.field}`), 'text-slate-400'));
    tr.appendChild(browseMetaCell(change.current || t('browse_meta_empty_value'),
      change.current ? 'text-amber-300/90 break-all' : 'text-slate-600 break-all'));
    tr.appendChild(browseMetaCell(change.target || t('browse_meta_empty_value'),
      change.target ? 'text-emerald-300 break-all' : 'text-slate-600 break-all'));
    rows.push(tr);
  });
  return rows;
}

function browseMetaStatusRow(file) {
  const tr = document.createElement('tr');
  tr.appendChild(browseMetaCell(file.name, 'text-slate-400 break-all'));
  const td = document.createElement('td');
  td.colSpan = 3;
  td.className = 'px-4 py-2 align-middle text-xs text-slate-500';
  let text = t(`browse_meta_status_${file.status}`);
  if (file.reason) text += ' — ' + t(`browse_meta_reason_${file.reason}`);
  if (file.error) text += ' — ' + file.error;
  td.textContent = text;
  tr.appendChild(td);
  return tr;
}

function browseMetaCell(text, cls) {
  const td = document.createElement('td');
  td.className = 'px-4 py-2 align-middle text-sm ' + cls;
  td.textContent = text;
  return td;
}

async function applyBrowseMeta() {
  // Only a preview arms this button, so applying always follows a look.
  if (!state.metaPath || !state.metaPreview) return;
  const btn = document.getElementById('browse-meta-apply');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_meta_applying');

  let data;
  try {
    data = await postFiles('/api/files/metadata/apply', browseMetaPayload());
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseMetaError(err.message || t('browse_meta_failed'));
    return;
  }
  btn.textContent = label;

  const failed = data.errors && data.errors.length;
  if (failed) showBrowseMetaError(data.errors.join(' · '));
  else hideBrowseMetaError();
  renderMetaNotices(document.getElementById('browse-meta-notices'), data.notices);
  showToast(t('browse_meta_done', { count: data.written || 0 }), failed ? 'error' : 'success');

  resetBrowseMetaPreview();
  await openBrowsePath(state.browsePath);
}

function showBrowseMetaError(message) {
  const el = document.getElementById('browse-meta-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseMetaError() {
  const el = document.getElementById('browse-meta-error');
  if (el) el.classList.add('hidden');
}

// --- Online lookup (folder panel only) -----------------------------------
//
// /api/metadata/search is source-agnostic and knows nothing about Kavita, so
// Browse reuses it as-is. The source is routed from the library we are in
// (manga → AniList, ebooks/audiobooks → Open Library) and stays overridable:
// a light novel filed under ebooks is legitimately an AniList entry.
//
// Picking a result fills the fields it carries *and* ticks them, because a
// value nobody armed would silently not be written. Fields the source did not
// return are left blank and unticked — the panel's own contract. Nothing
// reaches disk without Preview then Apply.
//
// Rows are built with DOM APIs: titles and summaries come from a third party
// and may contain quotes, which escapeHtml() does not escape.

// Last candidate list, kept whole so a click prefills from the object rather
// than from what the DOM managed to carry.
let browseOnlineCandidates = [];

// Which source a library defaults to. Anything that is not manga is a book.
function browseOnlineDefaultType(rootId) {
  return rootId === 'manga' ? 'manga' : 'book';
}

// Only http(s) links are rendered: a cover_url or site_url is remote data, and
// a javascript: URL in an href would execute on click.
function browseOnlineSafeURL(raw) {
  if (!raw) return '';
  try {
    const u = new URL(raw, window.location.origin);
    return (u.protocol === 'http:' || u.protocol === 'https:') ? u.href : '';
  } catch (err) {
    return '';
  }
}

function resetBrowseOnline() {
  const results = document.getElementById('browse-online-results');
  if (results) results.innerHTML = '';
  const status = document.getElementById('browse-online-status');
  if (status) {
    status.textContent = '';
    status.classList.add('hidden');
  }
  browseOnlineCandidates = [];
}

function showBrowseOnlineStatus(message) {
  const status = document.getElementById('browse-online-status');
  status.textContent = message;
  status.classList.remove('hidden');
}

async function searchBrowseOnline() {
  const input = document.getElementById('browse-online-query');
  const q = input ? input.value.trim() : '';
  // An empty query would only earn a 400; don't spend a round trip on it.
  if (!q) return;
  const type = document.getElementById('browse-online-source').value;
  const btn = document.getElementById('browse-online-btn');
  const label = btn.textContent;

  resetBrowseOnline();
  showBrowseOnlineStatus(t('browse_online_searching'));
  btn.disabled = true;
  btn.textContent = t('browse_online_searching');

  let data;
  try {
    // api() rather than apiJson(): the status code is the whole message here.
    // 503 means the source is not configured, 502 that it failed to answer —
    // two different problems, and "error" would name neither.
    const resp = await api(`/api/metadata/search?type=${encodeURIComponent(type)}&q=${encodeURIComponent(q)}`);
    if (resp.status === 503) {
      showBrowseOnlineStatus(t('browse_online_unavailable'));
      return;
    }
    if (resp.status === 502) {
      showBrowseOnlineStatus(t('browse_online_upstream'));
      return;
    }
    if (!resp.ok) {
      showBrowseOnlineStatus(t('browse_online_failed'));
      return;
    }
    data = await resp.json();
  } catch (err) {
    // Unauthorized already opened the login modal; anything else is the network.
    if (err.message !== 'Unauthorized') showBrowseOnlineStatus(t('browse_online_failed'));
    return;
  } finally {
    btn.disabled = false;
    btn.textContent = label;
  }

  browseOnlineCandidates = data.candidates || [];
  if (browseOnlineCandidates.length === 0) {
    showBrowseOnlineStatus(t('browse_online_none'));
    return;
  }
  document.getElementById('browse-online-status').classList.add('hidden');
  const results = document.getElementById('browse-online-results');
  browseOnlineCandidates.forEach((c, i) => results.appendChild(browseOnlineRow(c, i)));
}

function browseOnlineRow(c, index) {
  const row = document.createElement('div');
  row.className = 'browse-online-candidate flex items-start gap-3 p-2 rounded-lg border border-slate-700 hover:border-indigo-500 cursor-pointer transition-colors';
  row.dataset.action = 'pickBrowseOnline';
  row.dataset.idx = String(index);

  const cover = browseOnlineSafeURL(c.cover_url);
  if (cover) {
    const img = document.createElement('img');
    img.src = cover;
    img.alt = '';
    img.loading = 'lazy';
    img.className = 'h-14 w-10 object-cover rounded flex-shrink-0';
    row.appendChild(img);
  } else {
    const ph = document.createElement('div');
    ph.className = 'h-14 w-10 rounded bg-slate-700 flex-shrink-0';
    row.appendChild(ph);
  }

  const body = document.createElement('div');
  body.className = 'min-w-0 flex-1';

  const head = document.createElement('div');
  head.className = 'flex items-center gap-1.5';
  const title = document.createElement('span');
  title.className = 'text-sm font-medium text-white line-clamp-1';
  title.textContent = c.title || t('browse_online_unknown_title');
  head.appendChild(title);
  const site = browseOnlineSafeURL(c.site_url);
  if (site) {
    // Two same-named candidates are told apart by opening the source page.
    const link = document.createElement('a');
    link.href = site;
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    link.className = 'text-indigo-400 hover:text-indigo-300 text-xs';
    link.textContent = '↗';
    head.appendChild(link);
  }
  body.appendChild(head);

  const sub = document.createElement('div');
  sub.className = 'text-xs text-slate-400 line-clamp-1';
  sub.textContent = [c.year, (c.authors || []).join(', ')].filter(Boolean).join(' · ');
  body.appendChild(sub);

  const summary = document.createElement('div');
  summary.className = 'text-xs text-slate-500 line-clamp-2 mt-0.5';
  summary.textContent = c.summary || '';
  body.appendChild(summary);

  row.appendChild(body);
  return row;
}

// pickBrowseOnline fills and arms the fields the candidate carries. `title` is
// the *series* here: a folder holds a work, and the server refuses a `title`
// on a folder target anyway. `year` and `cover_url` are deliberately dropped —
// filewriter has nowhere to put them.
function pickBrowseOnline(index) {
  const c = browseOnlineCandidates[index];
  if (!c) return;

  const fill = (field, value) => {
    if (!value) return;
    const input = document.getElementById(`browse-meta-${field}`);
    const cb = document.getElementById(`browse-meta-en-${field}`);
    if (!input || !cb) return;
    cb.checked = true;
    input.disabled = false;
    input.value = value;
  };

  fill('series', c.title || '');
  fill('summary', c.summary || '');
  fill('publisher', c.publisher || '');
  fill('language', c.language || '');
  fill('authors', (c.authors || []).join(', '));
  fill('genres', (c.genres || []).join(', '));
  fill('tags', (c.tags || []).join(', '));

  // Setting values in code fires no change event, so drop the preview here:
  // it was measured against the form as it was a moment ago.
  resetBrowseMetaPreview();

  document.querySelectorAll('.browse-online-candidate').forEach(el => {
    const chosen = +el.dataset.idx === index;
    el.classList.toggle('border-indigo-500', chosen);
    el.classList.toggle('bg-indigo-500/10', chosen);
  });
  showBrowseOnlineStatus(t('browse_online_picked'));
}

// --- Single-file panel ---------------------------------------------------

async function openBrowseFileMeta(path) {
  closeBrowseMeta();
  state.metaFilePath = path;
  state.metaFileEditable = false;

  const panel = document.getElementById('browse-file-panel');
  const host = document.getElementById('browse-file-fields');
  host.innerHTML = '';
  BROWSE_META_FILE_FIELDS.forEach(field => {
    host.appendChild(buildMetaField(field, 'browse-file', false));
  });
  document.getElementById('browse-file-title').textContent =
    t('browse_meta_file_title', { name: browseBaseName(path) });
  panel.classList.remove('hidden');
  document.getElementById('browse-file-convert').classList.add('hidden');
  hideBrowseFileMetaError();

  // Metadata is read here and nowhere else (decision 9): opening one archive on
  // demand is affordable, opening every archive of a listing is not.
  let data;
  try {
    data = await apiJson(`/api/files/metadata?path=${encodeURIComponent(path)}`);
  } catch (err) {
    showBrowseFileMetaError(t('browse_meta_load_failed'));
    lockBrowseFileMeta('');
    return;
  }
  if (state.metaFilePath !== path) return; // a newer panel won the race

  const fields = data.fields || {};
  BROWSE_META_FILE_FIELDS.forEach(field => {
    const el = document.getElementById(`browse-file-${field}`);
    if (!el) return;
    const value = fields[field];
    el.value = Array.isArray(value) ? value.join(', ') : (value || '');
  });

  if (!data.editable) {
    lockBrowseFileMeta(data.reason || 'format_unsupported');
    return;
  }
  state.metaFileEditable = true;
  document.getElementById('browse-file-locked').classList.add('hidden');
  document.getElementById('browse-file-save').disabled = false;
}

// A locked file keeps its panel — greyed out, with the reason spelled out.
// Decision 11 asks for three distinct explanations, not one generic one,
// because what the user should do next differs: rename a PDF, convert a CBR,
// nothing at all for audio.
function lockBrowseFileMeta(reason) {
  state.metaFileEditable = false;
  BROWSE_META_FILE_FIELDS.forEach(field => {
    const el = document.getElementById(`browse-file-${field}`);
    if (el) el.disabled = true;
  });
  document.getElementById('browse-file-save').disabled = true;
  const banner = document.getElementById('browse-file-locked');
  if (!reason) {
    banner.classList.add('hidden');
    return;
  }
  banner.textContent = t('browse_meta_status_unsupported') + ' — ' + t(`browse_meta_reason_${reason}`);
  banner.classList.remove('hidden');
  // CBR and PDF are the locked formats with a way out: repack as CBZ and
  // the panel above becomes editable. Every other format stays convert-less.
  document.getElementById('browse-file-convert').classList
    .toggle('hidden', reason !== 'cbr_rar_archive' && reason !== 'pdf_ignored_by_kavita');
}

function closeBrowseFileMeta() {
  state.metaFilePath = null;
  state.metaFileEditable = false;
  const panel = document.getElementById('browse-file-panel');
  if (panel) panel.classList.add('hidden');
  hideBrowseFileMetaError();
}

async function applyBrowseFileMeta() {
  if (!state.metaFilePath || !state.metaFileEditable) return;
  const btn = document.getElementById('browse-file-save');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_meta_saving');

  // Every field is sent, emptied ones included: the panel shows the file's
  // current values, so what is on screen is what will be stored.
  const payload = {
    path: state.metaFilePath,
    mode: 'set',
    fields: browseMetaPayloadFields('browse-file', BROWSE_META_FILE_FIELDS, () => true),
  };

  let data;
  try {
    data = await postFiles('/api/files/metadata/apply', payload);
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseFileMetaError(err.message || t('browse_meta_failed'));
    return;
  }
  btn.textContent = label;
  btn.disabled = false;

  const failed = data.errors && data.errors.length;
  if (failed) {
    showBrowseFileMetaError(data.errors.join(' · '));
  } else {
    hideBrowseFileMetaError();
    closeBrowseFileMeta();
  }
  showToast(t('browse_meta_done', { count: data.written || 0 }), failed ? 'error' : 'success');
  await openBrowsePath(state.browsePath);
}

// convertBrowseFile repacks the open CBR as CBZ in place (unpack the RAR,
// repack it as ZIP — automated), then lands back in this panel, now editable.
// The source is removed server-side only once the replacement is on disk.
async function convertBrowseFile() {
  if (!state.metaFilePath) return;
  const btn = document.getElementById('browse-file-convert');
  const label = btn.textContent;
  btn.disabled = true;
  btn.textContent = t('browse_file_converting');

  let data;
  try {
    data = await postFiles('/api/files/convert', { path: state.metaFilePath });
  } catch (err) {
    btn.textContent = label;
    btn.disabled = false;
    showBrowseFileMetaError(err.message || t('browse_convert_failed'));
    return;
  }
  btn.textContent = label;
  btn.disabled = false;

  const skipped = (data && data.skipped) || [];
  showToast(skipped.length
    ? t('browse_convert_done_skipped', { count: data.files || 0, name: (data && data.name) || '', skipped: skipped.length })
    : t('browse_convert_done', { count: data.files || 0, name: (data && data.name) || '' }), 'success');
  // The listing must refresh first (it clears this panel), then the new file
  // reopens here — editable this time.
  await openBrowsePath(state.browsePath);
  await openBrowseFileMeta((data && data.path) || state.metaFilePath);
}

function showBrowseFileMetaError(message) {
  const el = document.getElementById('browse-file-error');
  el.textContent = message;
  el.classList.remove('hidden');
}

function hideBrowseFileMetaError() {
  const el = document.getElementById('browse-file-error');
  if (el) el.classList.add('hidden');
}
