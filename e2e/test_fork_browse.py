"""Fork-only end-to-end tests for the Browse tab.

Filesystem-backed library listing, folder merge preview, and online metadata
lookup — coverage added by the fork's Browse commits, isolated here so the
shared user-journey file stays close to upstream.

Runs against the same hermetic harness as test_user_journey.py (see
conftest.py): real binary + real Chromium, stubbed sources, no external
network. Fixture seeds (BROWSE_FIXTURE_*, MERGE_FIXTURE_*) live in conftest.py
and are imported, not duplicated. Every test also implicitly asserts a
zero-JS-error journey (ui fixture).
"""

import json

from conftest import (
    BROWSE_FIXTURE_DIR,
    BROWSE_FIXTURE_FILE,
    MERGE_FIXTURE_DIRS,
    MERGE_FIXTURE_SERIES,
)


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
    folder = page.locator(
        '#browse-list [data-action="browseOpen"]', has_text=BROWSE_FIXTURE_DIR
    )
    folder.first.click()
    page.wait_for_timeout(500)
    assert BROWSE_FIXTURE_FILE in page.inner_text("#browse-list")
    assert BROWSE_FIXTURE_DIR in page.inner_text("#browse-breadcrumb")
    assert page.locator("#browse-error").is_hidden()


def test_browse_merge_preview_lists_destinations(ui):
    """Smoke test of the folder merge: ticking two volume folders unlocks the
    action bar, and the panel previews one destination per file. Preview only —
    nothing is applied, so the fixture survives for the other tests."""
    page = ui["page"]
    page.click('[data-action="switchTab"][data-arg="browse"]')
    page.wait_for_selector('#browse-roots button[data-action="browseRoot"]')

    # One tick is not a merge: the bar shows, the button stays inert.
    for name in MERGE_FIXTURE_DIRS:
        page.locator(
            f'#browse-list input[data-action-change="browseSelect"]'
            f'[data-path$="{name}"]'
        ).check()
        page.wait_for_timeout(100)
        if name == MERGE_FIXTURE_DIRS[0]:
            assert page.locator("#browse-merge-btn").is_disabled()

    assert page.locator("#browse-select-bar").is_visible()
    page.click("#browse-merge-btn")
    page.wait_for_selector("#browse-merge-rows tr")

    # The target name is derived from the folder names' common tokens, with the
    # volume marker dropped.
    assert page.input_value("#browse-merge-series") == MERGE_FIXTURE_SERIES
    rows = page.inner_text("#browse-merge-rows")
    assert f"{MERGE_FIXTURE_SERIES} Vol.01.cbz" in rows
    assert f"{MERGE_FIXTURE_SERIES} Vol.02.cbz" in rows
    assert page.locator("#browse-merge-apply").is_enabled()
    assert page.locator("#browse-merge-error").is_hidden()

    # Leaving the folder must drop both the selection and the panel.
    page.click('[data-action="closeBrowseMerge"]')
    page.click('[data-action="clearBrowseSelection"]')
    assert page.locator("#browse-select-bar").is_hidden()


def _open_browse_meta_panel(page):
    """Navigate to the seeded folder and open its metadata panel."""
    page.click('[data-action="switchTab"][data-arg="browse"]')
    page.wait_for_selector('#browse-roots button[data-action="browseRoot"]')
    page.locator(
        '#browse-list [data-action="browseOpen"]', has_text=BROWSE_FIXTURE_DIR
    ).first.click()
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
            body=json.dumps(
                {
                    "candidates": [
                        {
                            "source": "openlibrary",
                            "title": "Stubbed Series",
                            "summary": "A stubbed summary.",
                            "authors": ["Stub Author"],
                            "genres": ["Stub Genre"],
                            "year": "1999",
                            "language": "en",
                        }
                    ]
                }
            ),
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
