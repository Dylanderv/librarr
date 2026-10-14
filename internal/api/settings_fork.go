package api

// forkSettingsMap returns defaults for the settings keys introduced by the
// fork. They are merged into the defaults map built by handleGetSettings (see
// settings.go) so upstream additions to that map never touch fork lines.
func (s *Server) forkSettingsMap() map[string]any {
	return map[string]any{
		"kavita_path_map": s.cfg.KavitaPathMap,
	}
}
