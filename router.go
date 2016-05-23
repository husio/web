package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/context"
)

type Routes []Route

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fn(context.Background(), w, r)
}

func (fn HandlerFunc) ServeCtxHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fn(ctx, w, r)
}

// Route binds together HTTP method, path and handler function.
type Route struct {
	// Path defines regexp-like pattern match used to determine if route should
	// handle request.
	Path string
	// Func defines HTTP handler that is used to serve request when route is
	// matching.
	//
	// Handler can be one following interfaces:
	// * http.Handler: interface with ServeHTTP(http.ResponseWriter, *http.Request)
	// * http.HandlerFunc: func(http.ResponseWriter, *http.Request)
	// * HandlerFunc: func(context.Context, http.ResponseWriter, *http.Request)
	// * interface with ServeCtxHTTP(context.Context, http.ResponseWriter, *http.Request)
	Handler interface{}

	// Method is string that can represent one or more, coma separated HTTP
	// methods that this route should match.
	// Method can be set to "*" to match any request no matter method.
	Methods string
}

// NewRouter create and return immutable router instance.
func NewRouter(routes Routes) *Router {
	var handlers []handler
	builder := regexp.MustCompile(`\(.*?\)`)

	for _, r := range routes {
		var names []string
		raw := builder.ReplaceAllStringFunc(r.Path, func(s string) string {
			s = s[1 : len(s)-1]
			// every {<name>} can be optionally contain separate regexp
			// definition using notation {<name>:<regexp>}
			chunks := strings.SplitN(s, ":", 2)
			if len(chunks) == 1 {
				names = append(names, s)
				return `([^/]+)`
			}
			names = append(names, chunks[0])
			return `(` + chunks[1] + `)`
		})
		// replace {} with regular expressions syntax
		rx, err := regexp.Compile(`^` + raw + `$`)
		if err != nil {
			panic(fmt.Sprintf("invalid routing path %q: %s", r.Path, err))
		}

		methods := make(map[string]struct{})
		for _, method := range strings.Split(r.Methods, ",") {
			methods[strings.TrimSpace(method)] = struct{}{}
		}

		var fn HandlerFunc
		switch h := r.Handler.(type) {
		case func(context.Context, http.ResponseWriter, *http.Request):
			fn = h
		case HandlerFunc:
			fn = h
		case ctxhandler:
			fn = func(ctx context.Context, w http.ResponseWriter, r *http.Request) { h.ServeCtxHTTP(ctx, w, r) }

		case http.Handler:
			fn = func(ctx context.Context, w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) }
		case http.HandlerFunc:
			fn = func(ctx context.Context, w http.ResponseWriter, r *http.Request) { h(w, r) }
		case func(http.ResponseWriter, *http.Request):
			fn = func(ctx context.Context, w http.ResponseWriter, r *http.Request) { h(w, r) }

		default:
			panic(fmt.Sprintf("invalid handler for %s %s: %T", r.Methods, r.Path, r.Handler))
		}

		handlers = append(handlers, handler{
			methods: methods,
			rx:      rx,
			names:   names,
			fn:      fn,
		})
	}

	return &Router{
		handlers:         handlers,
		NotFound:         StdTextHandler(http.StatusNotFound),
		MethodNotAllowed: StdTextHandler(http.StatusMethodNotAllowed),
	}
}

type ctxhandler interface {
	ServeCtxHTTP(context.Context, http.ResponseWriter, *http.Request)
}

type Router struct {
	handlers []handler

	// NotFound is called when none of defined handlers match current route.
	NotFound HandlerFunc

	// MethodNotAllowed is called when there is path match but not for current
	// method.
	MethodNotAllowed HandlerFunc
}

// ServeHTTP handle HTTP request using empty context.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	rt.ServeCtxHTTP(ctx, w, r)
}

// ServeCtxHTTP handle HTTP request using given context.
func (rt *Router) ServeCtxHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var pathMatch bool

	for _, h := range rt.handlers {
		match := h.rx.FindAllStringSubmatch(r.URL.Path, 1)
		if len(match) == 0 {
			continue
		}

		pathMatch = true

		_, ok := h.methods[r.Method]
		if !ok {
			_, ok = h.methods["*"]
		}
		if !ok {
			continue
		}

		values := match[0]
		ctx = context.WithValue(ctx, "router:args", &args{
			names:  h.names,
			values: values[1:],
		})
		h.fn(ctx, w, r)
		return
	}

	if pathMatch {
		rt.MethodNotAllowed(ctx, w, r)
	} else {
		rt.NotFound(ctx, w, r)
	}
}

type args struct {
	names  []string
	values []string
}

// WithArgs return context with HTTP args set to given list of pairs.
//
// Use for testing only.
func WithArgs(ctx context.Context, pairs ...string) context.Context {
	if len(pairs)%2 != 0 {
		panic("Invalid args list: pairs are not even")
	}
	a := args{}
	for i := 0; i < len(pairs); i += 2 {
		a.names = append(a.names, pairs[i])
		a.values = append(a.values, pairs[i+1])
	}
	return context.WithValue(ctx, "router:args", &a)
}

// Len return number of arguments.
func (a *args) Len() int {
	return len(a.values)
}

// ByName return URL mached value using name assigned to it. Returns empty
// string if does not exist.
func (a *args) ByName(name string) string {
	for i, n := range a.names {
		if n == name {
			return a.values[i]
		}
	}
	return ""
}

// ByIndex return URL mached value using definition position. Returns empty
// string if does not exist.
func (a *args) ByIndex(n int) string {
	if len(a.values) < n {
		return ""
	}
	return a.values[n]
}

type handler struct {
	methods map[string]struct{}
	rx      *regexp.Regexp
	names   []string
	fn      HandlerFunc
}

// Args return PathArgs carried by given context.
func Args(ctx context.Context) PathArgs {
	return ctx.Value("router:args").(*args)
}

type PathArgs interface {
	ByName(string) string
	ByIndex(int) string
}
