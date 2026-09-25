package httpapi

import "net/http"

// Health returns the liveness handler for one node.
//
// Liveness is a deliberately narrow claim: this says the process is answering
// and which node it is. It does not say that any other node is reachable, that
// a transfer would be safe, or that everything the node will ever do is ready.
// Keeping the claim small is what stops a passing health check being mistaken
// for proof that the whole system works.
//
// The handler takes the node id rather than reading it from a global, so a
// test can build one for any identity and assert what comes back.
func Health(nodeID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// A health endpoint answers one question, so it accepts one method.
		// Allow is set so a client can discover which, rather than having to
		// guess from a bare 405.
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			ErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed: use GET")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"node_id": nodeID,
			"status":  "ok",
		})
	}
}
