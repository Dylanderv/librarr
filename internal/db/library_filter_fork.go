package db

import (
	"strings"
)

// libraryFilter builds the WHERE clause shared by the list and count queries.
// The query term matches the fields a user can plausibly search on, source_id
// included: it is what the duplicate check in the download path keys on, so
// being able to look an item up by its info hash or MD5 is what makes a stale
// entry findable at all.
func libraryFilter(mediaType, query string) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if mediaType != "" {
		clauses = append(clauses, "media_type = ?")
		args = append(args, mediaType)
	}
	if q := strings.TrimSpace(query); q != "" {
		like := "%" + escapeLikePattern(q) + "%"
		clauses = append(clauses,
			`(title LIKE ? ESCAPE '\' OR author LIKE ? ESCAPE '\' OR file_path LIKE ? ESCAPE '\' OR source_id LIKE ? ESCAPE '\')`)
		args = append(args, like, like, like, like)
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

// escapeLikePattern neutralises the LIKE wildcards so a literal "%" or "_" in
// a title doesn't turn into a match-anything.
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	return strings.ReplaceAll(s, "_", `\_`)
}
