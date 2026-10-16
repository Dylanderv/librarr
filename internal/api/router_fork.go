package api

// registerForkRoutes wires the fork-specific endpoints: the Kavita metadata
// editor and the filesystem-backed library browser.
//
// Keeping these registrations in their own file (called once from the tail of
// registerLibraryRoutes) means upstream changes to the library, wishlist, or
// files sections of router.go merge without touching fork lines.
func (s *Server) registerForkRoutes() {
	// Kavita metadata editing (writes to the media files, then rescans).
	s.mux.HandleFunc("GET /api/library/manga/series/{id}/metadata", requireAdmin(s.handleSeriesMetadata))
	s.mux.HandleFunc("POST /api/library/metadata/bulk", requireAdmin(s.handleBulkMetadata))
	s.mux.HandleFunc("POST /api/library/metadata/rename-preview", requireAdmin(s.handleRenamePreview))
	// Online metadata lookup to pre-fill the bulk editor (AniList/Open Library).
	s.mux.HandleFunc("GET /api/metadata/search", requireAdmin(s.handleMetadataSearch))

	// Filesystem-backed library browser. Admin-only, like
	// /api/library/metadata/*: these endpoints expose the server's directory
	// tree, which is strictly more revealing than the metadata editor they
	// will feed, so they cannot be looser.
	s.mux.HandleFunc("GET /api/files/roots", requireAdmin(s.handleFilesRoots))
	s.mux.HandleFunc("GET /api/files/list", requireAdmin(s.handleFilesList))
	s.mux.HandleFunc("POST /api/files/rename/preview", requireAdmin(s.handleFilesRenamePreview))
	s.mux.HandleFunc("POST /api/files/rename/apply", requireAdmin(s.handleFilesRenameApply))
	s.mux.HandleFunc("POST /api/files/rename/file", requireAdmin(s.handleFilesRenameFile))
	s.mux.HandleFunc("POST /api/files/mkdir", requireAdmin(s.handleFilesMkdir))
	s.mux.HandleFunc("POST /api/files/convert", requireAdmin(s.handleFilesConvert))
	s.mux.HandleFunc("POST /api/files/move", requireAdmin(s.handleFilesMove))
	s.mux.HandleFunc("POST /api/files/merge/preview", requireAdmin(s.handleFilesMergePreview))
	s.mux.HandleFunc("POST /api/files/merge/apply", requireAdmin(s.handleFilesMergeApply))
	s.mux.HandleFunc("GET /api/files/metadata", requireAdmin(s.handleFilesMetadata))
	s.mux.HandleFunc("POST /api/files/metadata/preview", requireAdmin(s.handleFilesMetadataPreview))
	s.mux.HandleFunc("POST /api/files/metadata/apply", requireAdmin(s.handleFilesMetadataApply))
	s.mux.HandleFunc("POST /api/files/kavita/scan", requireAdmin(s.handleFilesKavitaScan))

	// Prowlarr advanced search: live category tree for the query/category
	// override panel. Same auth as /api/search (no indexer secrets leak: only
	// taxonomy IDs and names are exposed).
	s.mux.HandleFunc("GET /api/prowlarr/categories", s.handleProwlarrCategories)
}
