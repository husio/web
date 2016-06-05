[![Build Status](https://travis-ci.org/husio/web.svg?branch=master)](https://travis-ci.org/husio/web)
[![GoDoc](https://godoc.org/github.com/husio/web?status.png)](https://godoc.org/github.com/husio/web)


# Web

`web` contains declarative, regexp based router and helper functions to write
web application in Go.


## Example application

See [example application](https://github.com/husio/web/tree/master/example),
which is split into several parts:

* `main.go` contains `main` function and setup required to run application but
   nothing relevant for the tests.
* `app.go` defines `application` that binds context with router. This is where
   you define how your HTTP application looks like. Define all the routes,
   install all middlewares.
* `handlers.go` contains all handler definitions. Those must be never public
   and accessed only via router instance.
* `db.go` contains `Entity` persistance layer implementation and two helper
   functions:
  * `WithDB` to attach selected database into context. This should be used both
     in `main` function and in tests,
  * `DB` return database from the context.
