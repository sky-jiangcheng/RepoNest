package httpapi

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
)

// /api/rpc is a generic JSON-RPC bridge to the SAME bound object the desktop
// Wails App exposes (internal/app.App). The web frontend's transport calls it
// in browser/standalone mode, so every Wails binding works over HTTP without a
// hand-written REST route per method — and without changing the existing
// REST endpoints the dsh-plugin/agents rely on.
//
// It is loopback-only (wrapped by loopbackGuard like the rest of this server),
// so it carries the same trust boundary as the direct bindings: it can do
// anything the desktop UI can.

var errorType = reflect.TypeOf((*error)(nil)).Elem()

type rpcRequest struct {
	Method string            `json:"method"`
	Args   []json.RawMessage `json:"args"`
}

func (h *handler) rpc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad request: " + err.Error()})
		return
	}
	if !h.bound.IsValid() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "rpc not available"})
		return
	}
	m := h.bound.MethodByName(req.Method)
	if !m.IsValid() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown method: " + req.Method})
		return
	}
	mt := m.Type()
	if mt.IsVariadic() {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "variadic methods not supported"})
		return
	}
	in := make([]reflect.Value, mt.NumIn())
	for i := 0; i < mt.NumIn(); i++ {
		p := reflect.New(mt.In(i))
		if i < len(req.Args) && len(req.Args[i]) > 0 && string(req.Args[i]) != "null" {
			if err := json.Unmarshal(req.Args[i], p.Interface()); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "arg " + strconv.Itoa(i) + ": " + err.Error()})
				return
			}
		}
		in[i] = p.Elem()
	}

	out := m.Call(in)
	var result any
	switch {
	case len(out) == 1 && out[0].Type() == errorType:
		if !out[0].IsNil() {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": out[0].Interface().(error).Error()})
			return
		}
	case len(out) == 2: // (T, error)
		if !out[1].IsNil() {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": out[1].Interface().(error).Error()})
			return
		}
		result = out[0].Interface()
	case len(out) == 1:
		result = out[0].Interface()
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}
