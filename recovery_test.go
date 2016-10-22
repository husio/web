package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panic" {
			panic("as expected")
		}
		w.WriteHeader(http.StatusTeapot)
	})
	hn := Recovery(fn)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/panic", nil)

	hn(w, r)

	if want, got := http.StatusInternalServerError, w.Code; want != got {
		t.Fatalf("when panic, want %d, got %d", want, got)
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/teapot?", nil)

	hn(w, r)

	if want, got := http.StatusTeapot, w.Code; want != got {
		t.Fatalf("when panic, want %d, got %d", want, got)
	}
}
