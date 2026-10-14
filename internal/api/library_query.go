package api

import (
	"net/http"
	"strings"
)

// libraryQueryParam returns the trimmed ?q= free-text filter shared by the
// library tabs. The UI's single search box feeds every tab, so each handler
// honors it the same way.
func libraryQueryParam(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("q"))
}
