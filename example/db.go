package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"
)

type Entry struct {
	ID      string
	Content string
	Created time.Time
	Updated time.Time
}

type Database struct {
	mu  sync.Mutex
	idx map[string]Entry
}

func (db *Database) List() []Entry {
	var entries []Entry

	db.mu.Lock()
	defer db.mu.Unlock()

	for _, e := range db.idx {
		entries = append(entries, e)
	}
	return entries
}

func (db *Database) Get(id string) (Entry, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entry, ok := db.idx[id]
	return entry, ok
}

func (db *Database) Create(content string) Entry {
	db.mu.Lock()
	defer db.mu.Unlock()

	now := time.Now()
	entry := Entry{
		ID:      fmt.Sprintf("entry-%d", now.UnixNano()),
		Content: content,
		Created: now,
		Updated: now,
	}
	db.idx[entry.ID] = entry
	return entry
}

func (db *Database) Set(id, content string) (Entry, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	entry, ok := db.idx[id]
	if !ok {
		return Entry{}, false
	}

	entry.Updated = time.Now()
	entry.Content = content
	db.idx[id] = entry

	return entry, true
}

func WithDB(ctx context.Context) context.Context {
	db := &Database{
		idx: make(map[string]Entry),
	}
	return context.WithValue(ctx, "database", db)
}

func DB(ctx context.Context) *Database {
	v := ctx.Value("database")
	if v == nil {
		panic("database not present in context")
	}
	return v.(*Database)
}
