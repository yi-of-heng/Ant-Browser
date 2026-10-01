package launchcode

import (
	"net/http"
)

// handleListProxies exposes read-only proxy metadata for CLI/MCP callers.
// Mutations remain in the desktop API until their confirmation semantics are
// also available to non-GUI clients.
func (s *LaunchServer) handleListProxies(w http.ResponseWriter, _ *http.Request) {
	if s.browserMgr == nil || s.browserMgr.ProxyDAO == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "proxy service unavailable"})
		return
	}
	items, err := s.browserMgr.ProxyDAO.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "count": len(items), "items": items})
}
