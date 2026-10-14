package api

import (
	"github.com/JeremiahM37/librarr/internal/kavita"
)

// kavitaClient builds a Kavita client from the current config. A fresh client
// per request keeps runtime settings changes honored while still caching the
// JWT across the multiple calls a single request may make.
func (s *Server) kavitaClient() *kavita.Client {
	return kavita.New(s.cfg.KavitaURL, s.cfg.KavitaUser, s.cfg.KavitaPass)
}
