package api

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/rs/zerolog/log"
)

// registerDebugPprofRoutes mounts Go's pprof handlers under /debug/pprof for
// operators, protected by admin/instance_admin auth. The endpoints are
// disabled by default (server.debug_pprof) because profiles expose internals
// (goroutine stacks, heap contents) and sampling can be expensive; see
// docs/src/content/docs/guides/operational-runbook.md.
func (s *Server) registerDebugPprofRoutes() {
	if !s.config.Server.DebugPprof {
		return
	}

	s.app.Get("/debug/pprof", s.requireAuth, RequireRole("admin", "instance_admin"), adaptor.HTTPHandlerFunc(pprof.Index))
	s.app.Get("/debug/pprof/*", s.requireAuth, RequireRole("admin", "instance_admin"), adaptor.HTTPHandlerFunc(debugPprofDispatch))

	log.Info().Msg("Debug pprof endpoints enabled at /debug/pprof (admin/instance_admin auth required)")
}

// debugPprofDispatch routes /debug/pprof/<name> to the matching pprof
// profile handler, mirroring net/http/pprof's own mux.
func debugPprofDispatch(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/debug/pprof/")
	switch name {
	case "cmdline":
		pprof.Cmdline(w, r)
	case "profile":
		pprof.Profile(w, r)
	case "symbol":
		pprof.Symbol(w, r)
	case "trace":
		pprof.Trace(w, r)
	default:
		// heap, goroutine, allocs, block, mutex, threadcreate, and the
		// bare profile index
		pprof.Handler(name).ServeHTTP(w, r)
	}
}
