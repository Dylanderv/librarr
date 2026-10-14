package api

import (
	"fmt"
	"net/http"
	"strconv"
)

// handleDeleteManga removes a manga row from librarr's own library. Unlike books
// and audiobooks there is no external system to mirror the delete to: manga rows
// only ever come from the local DB, so the id is always an internal one and a
// non-numeric id is a caller bug rather than a foreign identifier.
//
// The file on disk is deliberately left alone. Removing the row is what releases
// the duplicate check in the download path, which is the whole point of exposing
// this; deleting media is a separate, destructive concern.
func (s *Server) handleDeleteManga(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid library item id",
		})
		return
	}
	if err := s.db.DeleteItem(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	username, _ := r.Context().Value(ctxUsername).(string)
	s.db.LogActivity(username, "library_remove", idStr, fmt.Sprintf("Removed library item %s", idStr))
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}
