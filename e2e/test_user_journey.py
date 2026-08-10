"""End-to-end user-journey tests for the librarr web UI.

Runs the real binary + real Chromium against a stubbed Gutenberg source (see
conftest.py) — no external network. Covers the full scope of a user:

  boot → strict CSP → tab navigation → search → sort → download → import →
  library → wishlist CRUD → cover fallback → language toggle

Every test also implicitly asserts a zero-JS-error journey (ui fixture).
"""
import json
import time

import pytest

from conftest import BROWSE_FIXTURE_DIR, BROWSE_FIXTURE_FILE


def active_tab(page):
    return page.evaluate("() => document.querySelector('.tab-content.active')?.id")


def api(page, path):
    return page.evaluate(f"() => fetch('{path}').then(r => r.json())")


# ── Boot, security posture, assets ─────────────────────────────────────────

def test_boot_serves_ui_with_strict_csp(ui):
    page = ui["page"]
    resp = page.request.get(ui["base"] + "/")
    assert resp.status == 200
    csp = resp.headers.get("content-security-policy", "")
    assert "script-src 'self'" in csp, f"strict CSP missing: {csp}"

    html = resp.text()
    for attr in ("onclick=", "onerror=", "onchange=", "onsubmit="):
        assert attr not in html, f"inline handler {attr} leaked back into index.html"
    assert "<script>" not in html, "inline script block leaked back into index.html"

    for asset in ("/static/js/app.js", "/static/css/app.css",
                  "/static/js/vendor/tailwind.js", "/static/fonts/inter-latin.woff2"):
        r = page.request.get(ui["base"] + asset)
        assert r.status == 200, f"{asset} -> {r.status}"


def test_tailwind_and_font_render_offline(ui):
    page = ui["page"]
    bg = page.evaluate(
        "() => getComputedStyle(document.querySelector('[class*=bg-slate]')).backgroundColor")
    assert bg not in ("", "rgba(0, 0, 0, 0)"), "Tailwind runtime produced no styles"
    font = page.evaluate("() => getComputedStyle(document.body).fontFamily")
    assert "Inter" in font


# ── Navigation via event delegation ─────────────────────────────────────────

def test_all_tabs_switch(ui):
    page = ui["page"]
    for tab in ("library", "browse", "downloads", "wishlist", "settings", "search"):
        page.click(f'[data-action="switchTab"][data-arg="{tab}"]')
        page.wait_for_timeout(200)
        assert active_tab(page) == f"tab-{tab}"


# ── Search / sort / the data-key contract ───────────────────────────────────

@pytest.fixture()
def searched(ui):
    page = ui["page"]
    page.fill("#search-input", "test adventure")
    page.press("#search-input", "Enter")
    page.wait_for_selector('[data-action="startDownload"]', timeout=30000)
    return ui


def test_search_renders_stub_results(searched):
    page = searched["page"]
    cards = page.locator('[data-action="startDownload"]').count()
    assert cards == 3, f"expected the 3 stub books, got {cards} cards"
    body = page.inner_text("#tab-search")
    assert "Test Adventure" in body


# The download button carries a stable data-key resolved against
# state.resultsByKey. This walks every rendered button — card (grid) or table
# row (list) — and checks the title it shows is the title of the result its key
# maps to.
_MAPPING_PROBE = """() =>
    [...document.querySelectorAll('[data-action="startDownload"]')].map(btn => {
        const card = btn.closest('.book-card');
        const row = btn.closest('tr');
        const shown = card
            ? card.querySelector('h3')?.textContent.trim()
            : row?.children[1]?.textContent.trim();
        const mapped = state.resultsByKey.get(btn.dataset.key)?.title;
        return shown === mapped ? null : `${shown} != ${mapped}`;
    }).filter(Boolean)"""


def test_download_key_maps_to_rendered_result_in_every_sort_mode(searched):
    """The download button carries data-key into the *rendered* (sorted) list.
    A mismatch here would download the wrong book — the most dangerous
    possible regression of the inline-handler removal."""
    page = searched["page"]

    # Grid layout: the toolbar sort buttons are only shown here (in list layout
    # they are hidden and sorting goes through the column headers instead).
    page.click('[data-action="setViewLayout"][data-arg="grid"]')
    page.wait_for_timeout(200)
    for mode in ("size", "seeders", "relevance"):
        page.click(f'#search-sort-buttons [data-action="setSortMode"][data-arg="{mode}"]')
        page.wait_for_timeout(300)
        assert page.evaluate(_MAPPING_PROBE) == [], f"grid sort={mode}"

    # List layout: sorting is driven by the table column headers. One <thead>
    # per swimlane, so scope to the first — setSortMode is global anyway.
    page.click('[data-action="setViewLayout"][data-arg="list"]')
    page.wait_for_timeout(200)
    for col in ("title", "size", "seeders"):
        page.locator(f'#search-results th[data-action="setSortMode"][data-arg="{col}"]').first.click()
        page.wait_for_timeout(300)
        assert page.evaluate(_MAPPING_PROBE) == [], f"list sort={col}"


def test_retry_wait_shows_progress_in_search_and_downloads(searched):
    page = searched["page"]
    retry_text = page.evaluate("""() => {
        const probe = document.querySelector('[data-action="startDownload"]');
        const result = state.resultsByKey.get(probe.dataset.key);
        const key = getDownloadKey(result);
        state.trackedDownloadJobs.set('retry-probe', {
            key, title: result.title, source: 'annas', url: result.url || ''
        });
        state.downloadJobs = [{
            job_id: 'retry-probe', title: result.title, source: 'annas',
            status: 'retry_wait', detail: 'Retry 1/2 scheduled',
            retry_count: 1, max_retries: 2, error: 'download HTTP 504'
        }];
        renderSearchResults();
        renderDownloadList();
        return {
            button: document.querySelector('[data-action="startDownload"]')?.innerText,
            downloads: document.querySelector('#downloads-list')?.innerText,
        };
    }""")
    assert "Retry 1/2 scheduled" in retry_text["button"]
    assert "Retry 1/2 scheduled" in retry_text["downloads"]
    assert "Attempt 2/3" in retry_text["downloads"]

    page.evaluate("""() => {
        state.trackedDownloadJobs.delete('retry-probe');
        state.downloadJobs = [];
        renderSearchResults();
        renderDownloadList();
    }""")


def test_download_completes_and_lands_in_library(searched):
    page = searched["page"]
    # Download the first rendered card; remember which book it claims to be.
    title = page.evaluate("""() => {
        const btn = document.querySelector('[data-action="startDownload"]');
        return state.resultsByKey.get(btn.dataset.key).title;
    }""")
    page.click('[data-action="startDownload"]')

    # Poll the API until the job finishes (direct download from the stub).
    deadline = time.time() + 60
    last = None
    while time.time() < deadline:
        jobs = api(page, "/api/downloads")
        rows = jobs.get("downloads") or []
        ours = [j for j in rows if j.get("title") == title]
        if ours:
            last = ours[0]
            if last.get("status") == "completed":
                break
            assert last.get("status") not in ("error", "dead_letter"), \
                f"download failed: {json.dumps(last)}"
        time.sleep(2)  # gentle poll — 1/s trips the API rate limiter
    assert last and last.get("status") == "completed", f"job never completed: {last}"

    # The epub must exist on disk under the library dir.
    files = list(searched["books_dir"].rglob("*.epub"))
    assert files, "no .epub imported into the library directory"

    # And the Downloads tab renders the job row.
    page.click('[data-action="switchTab"][data-arg="downloads"]')
    page.click('[data-action="refreshDownloads"]')
    page.wait_for_timeout(500)
    assert title in page.inner_text("#tab-downloads")


# ── Wishlist CRUD through delegated row buttons ─────────────────────────────

def test_wishlist_add_search_delete(ui):
    page = ui["page"]
    page.click('[data-action="switchTab"][data-arg="wishlist"]')
    page.click('[data-action="showWishlistForm"]')
    page.wait_for_timeout(200)
    first_input = page.evaluate(
        "() => [...document.querySelectorAll('#wishlist-form input')]"
        ".find(i => i.type !== 'checkbox')?.id")
    assert first_input, "wishlist form has no text input"
    page.fill(f"#{first_input}", "The Hobbit")
    page.click('[data-action="addWishlistItem"]')
    page.wait_for_selector('[data-action="deleteWishlistItem"]', timeout=5000)

    # Row's search button jumps to the search tab with the query.
    page.click('[data-action="searchWishlistItem"]')
    page.wait_for_timeout(400)
    assert active_tab(page) == "tab-search"

    # Delete removes the row.
    page.click('[data-action="switchTab"][data-arg="wishlist"]')
    page.wait_for_timeout(200)
    page.click('[data-action="deleteWishlistItem"]')
    page.wait_for_timeout(500)
    assert page.locator('[data-action="deleteWishlistItem"]').count() == 0


# ── Browse tab: filesystem-backed listing ───────────────────────────────────

def test_browse_lists_roots_and_navigates_into_a_folder(ui):
    """Smoke test of the Browse tab: the configured libraries show up, and
    walking into a folder lists the files it really holds on disk."""
    page = ui["page"]
    page.click('[data-action="switchTab"][data-arg="browse"]')
    page.wait_for_selector('#browse-roots button[data-action="browseRoot"]')
    roots = page.locator('#browse-roots button[data-action="browseRoot"]').count()
    assert roots >= 1, "no library root rendered in Browse"
    assert page.locator("#browse-no-roots").is_hidden()
    assert page.locator("#browse-error").is_hidden()

    # The seeded folder is listed, and opening it shows the file inside.
    folder = page.locator('#browse-list [data-action="browseOpen"]',
                          has_text=BROWSE_FIXTURE_DIR)
    folder.first.click()
    page.wait_for_timeout(500)
    assert BROWSE_FIXTURE_FILE in page.inner_text("#browse-list")
    assert BROWSE_FIXTURE_DIR in page.inner_text("#browse-breadcrumb")
    assert page.locator("#browse-error").is_hidden()


def _open_browse_meta_panel(page):
    """Navigate to the seeded folder and open its metadata panel."""
    page.click('[data-action="switchTab"][data-arg="browse"]')
    page.wait_for_selector('#browse-roots button[data-action="browseRoot"]')
    page.locator('#browse-list [data-action="browseOpen"]',
                 has_text=BROWSE_FIXTURE_DIR).first.click()
    page.wait_for_timeout(500)
    page.click("#browse-meta-btn")
    page.wait_for_selector("#browse-online")


def test_browse_online_lookup_prefills_and_arms_fields(ui):
    """Online lookup in the Browse folder panel. The upstream call is stubbed at
    the browser, so the test never reaches AniList or Open Library."""
    page = ui["page"]
    _open_browse_meta_panel(page)

    # The library we are in routes the source (the e2e root is EBOOK_DIR), and
    # the query starts from the folder name.
    assert page.input_value("#browse-online-source") == "book"
    assert page.input_value("#browse-online-query") == BROWSE_FIXTURE_DIR

    page.route(
        "**/api/metadata/search*",
        lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body=json.dumps({"candidates": [{
                "source": "openlibrary",
                "title": "Stubbed Series",
                "summary": "A stubbed summary.",
                "authors": ["Stub Author"],
                "genres": ["Stub Genre"],
                "year": "1999",
                "language": "en",
            }]}),
        ),
    )
    page.click("#browse-online-btn")
    page.wait_for_selector(".browse-online-candidate")
    results = page.inner_text("#browse-online-results")
    assert "Stubbed Series" in results
    # Year is shown to help identify the result, never written.
    assert "1999" in results

    page.click(".browse-online-candidate")
    page.wait_for_timeout(200)
    assert page.input_value("#browse-meta-series") == "Stubbed Series"
    assert page.is_checked("#browse-meta-en-series")
    assert page.input_value("#browse-meta-authors") == "Stub Author"
    assert page.is_checked("#browse-meta-en-authors")
    assert page.input_value("#browse-meta-genres") == "Stub Genre"
    # A field the source did not return stays empty and unarmed.
    assert page.input_value("#browse-meta-publisher") == ""
    assert not page.is_checked("#browse-meta-en-publisher")
    # Nothing is written without a preview: Apply is still locked.
    assert page.locator("#browse-meta-apply").is_disabled()
    page.unroute("**/api/metadata/search*")


def test_browse_online_lookup_reports_an_unconfigured_source(ui):
    """A 503 means the source is not configured — its own message, not a
    generic failure."""
    page = ui["page"]
    _open_browse_meta_panel(page)
    page.route(
        "**/api/metadata/search*",
        lambda route: route.fulfill(
            status=503,
            content_type="application/json",
            body=json.dumps({"error": "Open Library client not available"}),
        ),
    )
    page.click("#browse-online-btn")
    page.wait_for_timeout(500)
    status = page.inner_text("#browse-online-status")
    assert "not configured" in status
    assert page.locator(".browse-online-candidate").count() == 0
    page.unroute("**/api/metadata/search*")


# ── Cover fallback via capture-phase error delegation ───────────────────────

def test_broken_cover_falls_back_to_placeholder(ui):
    page = ui["page"]
    ok = page.evaluate("""() => new Promise(res => {
        const img = document.createElement('img');
        img.dataset.phTitle = 'Fallback Probe';
        img.dataset.phIdx = '0';
        img.src = '/static/definitely-missing.png';
        document.body.appendChild(img);
        setTimeout(() => res(!!document.querySelector('.cover-placeholder')), 800);
    })""")
    assert ok, "broken cover did not swap to the gradient placeholder"


# ── i18n toggle (re-renders DOM incl. converted anchor templates) ───────────

def test_language_toggle_rerenders(ui):
    page = ui["page"]
    before = page.evaluate("() => document.body.innerText.slice(0, 300)")
    page.click('[data-action="toggleLanguage"]')
    page.wait_for_timeout(400)
    after = page.evaluate("() => document.body.innerText.slice(0, 300)")
    assert before != after, "language toggle changed nothing"
    page.click('[data-action="toggleLanguage"]')  # restore EN for later tests
    page.wait_for_timeout(300)
