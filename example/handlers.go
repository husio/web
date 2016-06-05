package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/husio/web"

	"golang.org/x/net/context"
)

func handleList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	db := DB(ctx)
	resp := struct {
		Entries []Entry
	}{
		Entries: db.List(),
	}
	web.JSONResp(w, resp, http.StatusOK)
}

func handleCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.JSONErr(w, fmt.Sprintf("cannot decode: %s", err), http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		web.JSONErr(w, `"Content" is required`, http.StatusBadRequest)
		return
	}

	db := DB(ctx)
	entry := db.Create(input.Content)
	web.JSONResp(w, entry, http.StatusCreated)
}

func handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	db := DB(ctx)
	entry, ok := db.Get(web.Args(ctx).ByIndex(0))
	if !ok {
		web.StdJSONResp(w, http.StatusNotFound)
		return
	}
	web.JSONResp(w, entry, http.StatusOK)
}

func handleSet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.JSONErr(w, fmt.Sprintf("cannot decode: %s", err), http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		web.JSONErr(w, `"Content" is required`, http.StatusBadRequest)
		return
	}

	db := DB(ctx)
	entry, ok := db.Set(web.Args(ctx).ByIndex(0), input.Content)
	if !ok {
		web.StdJSONResp(w, http.StatusNotFound)
		return
	}

	web.JSONResp(w, entry, http.StatusOK)
}
