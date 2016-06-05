package main

import (
	"net/http"

	"github.com/husio/web"
	"golang.org/x/net/context"
)

func NewApp(ctx context.Context) http.Handler {
	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{
			{`/`, handleList, "GET"},
			{`/`, handleCreate, "POST"},
			{`/(entry-id)`, handleGet, "GET"},
			{`/(entry-id)`, handleSet, "PUT"},
		}),
	}
}

type application struct {
	ctx context.Context
	rt  *web.Router
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.rt.ServeCtxHTTP(app.ctx, w, r)
}
