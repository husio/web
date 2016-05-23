package web

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// JSONResp write content as JSON encoded response.
func JSONResp(w http.ResponseWriter, content interface{}, code int) {
	b, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		code = http.StatusInternalServerError
		b = []byte(`{"errors":["Internal Server Errror"]}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	w.Write(b)
}

// JSONErr write single error as JSON encoded response.
func JSONErr(w http.ResponseWriter, errText string, code int) {
	JSONErrs(w, []string{errText}, code)
}

// JSONErrs write multiple errors as JSON encoded response.
func JSONErrs(w http.ResponseWriter, errs []string, code int) {
	resp := struct {
		Code   int
		Errors []string `json:"errors"`
	}{
		Code:   code,
		Errors: errs,
	}
	JSONResp(w, resp, code)
}

// StdJSONResp write JSON encoded, standard HTTP response text for given status
// code. Depending on status, either error or successful response format is
// used.
func StdJSONResp(w http.ResponseWriter, code int) {
	if code >= 400 {
		JSONErr(w, http.StatusText(code), code)
	} else {
		JSONResp(w, http.StatusText(code), code)
	}
}

// JSONRedirect return redirect response, but with JSON formatted body.
func JSONRedirect(w http.ResponseWriter, urlStr string, code int) {
	w.Header().Set("Location", urlStr)
	var content = struct {
		Code     int
		Location string
	}{
		Code:     code,
		Location: urlStr,
	}
	JSONResp(w, content, code)
}

// Modified check given request for If-Modified-Since header and if present,
// compares it with given modification time. Returns true and set Last-Modified
// header value if modification was made, otherwise write NotModified response
// to the client.
func Modified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	// https://golang.org/src/net/http/fs.go#L273
	ms, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since"))
	if err == nil && modtime.Before(ms.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return false
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return true
}

// RedirectHandler return HandlerFunc that always redirect to given url.
func RedirectHandler(url string, code int) HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	}
}

// StdJSONHandler return HandlerFunc that always response with JSON encoded,
// standard for given status code text message.
func StdJSONHandler(code int) HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		StdJSONResp(w, code)
	}
}

// StaticHandler return HandlerFunc that serve content of given directory.
func StaticHandler(root string) HandlerFunc {
	h := http.StripPrefix("/"+root, http.FileServer(http.Dir(root)))
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
