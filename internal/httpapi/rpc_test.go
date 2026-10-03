package httpapi

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"repo-nest/internal/db"
	"repo-nest/internal/service"
)

type fakeBound struct{}

func (fakeBound) Echo(s string) string  { return s }
func (fakeBound) Add(a, b int64) int64  { return a + b }
func (fakeBound) Fail() error           { return errors.New("boom") }
func (fakeBound) Pair() (string, error) { return "ok", nil }
func (fakeBound) List() []int64         { return []int64{1, 2, 3} }

func newRPCHandler(t *testing.T) *handler {
	t.Helper()
	d, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return &handler{svc: service.New(d, "me"), bound: reflect.ValueOf(fakeBound{})}
}

func rpc(t *testing.T, h *handler, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/rpc", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.rpc(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestRPCHappyPaths(t *testing.T) {
	h := newRPCHandler(t)
	if code, out := rpc(t, h, `{"method":"Echo","args":["hi"]}`); code != 200 || out["result"] != "hi" {
		t.Fatalf("Echo: code=%d out=%v", code, out)
	}
	if code, out := rpc(t, h, `{"method":"Add","args":[2,3]}`); code != 200 || out["result"].(float64) != 5 {
		t.Fatalf("Add: code=%d out=%v", code, out)
	}
	if code, out := rpc(t, h, `{"method":"Pair"}`); code != 200 || out["result"] != "ok" {
		t.Fatalf("Pair: code=%d out=%v", code, out)
	}
	if code, out := rpc(t, h, `{"method":"List"}`); code != 200 {
		t.Fatalf("List code=%d", code)
	} else if _, ok := out["result"].([]any); !ok {
		t.Fatalf("List result not array: %v", out["result"])
	}
}

func TestRPCErrorPaths(t *testing.T) {
	h := newRPCHandler(t)
	// method returns error -> 422 with message.
	if code, out := rpc(t, h, `{"method":"Fail"}`); code != 422 || out["error"] != "boom" {
		t.Fatalf("Fail: code=%d out=%v", code, out)
	}
	// unknown method -> 404.
	if code, _ := rpc(t, h, `{"method":"Nope"}`); code != 404 {
		t.Fatalf("unknown method code=%d, want 404", code)
	}
	// bad arg type -> 400.
	if code, _ := rpc(t, h, `{"method":"Add","args":["x",2]}`); code != 400 {
		t.Fatalf("bad arg code=%d, want 400", code)
	}
	// GET not allowed -> 405.
	rec := httptest.NewRecorder()
	h.rpc(rec, httptest.NewRequest("GET", "/api/rpc", nil))
	if rec.Code != 405 {
		t.Fatalf("GET code=%d, want 405", rec.Code)
	}
}
