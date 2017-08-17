# Midway [![Documentations][shield-godoc]][godoc] [![Travis CI results][shield-travis]][travis]

[godoc]: https://godoc.org/github.com/go-midway/midway
[shield-godoc]: https://img.shields.io/badge/godoc-reference-5272B4.svg
[travis]: https://travis-ci.org/go-midway/midway
[shield-travis]: https://api.travis-ci.org/go-midway/midway.svg?branch=master

A simple middleware collection to work with [http.Handler][http.Handler]
implementations, and maybe [go-kit][go-kit].

Inspired by go-kit's middleware chaining.

[http.Handler]: https://golang.org/pkg/net/http/#Handler
[go-kit]: https://github.com/go-kit/kit


## Basic Design

Like other similar efforts that inspired by go1.7 release and the changes
in [net/http and context](https://golang.org/doc/go1.7#context), this is
a collection of opiniated middlewares that fulfill this signature:

```go

type Middleware func(http.Handler) http.Handler

```

Along with some utils to [chain middleware][middleware.Chain] and to
[rewrite functions][funconv], this collection of middleware passes variables
into http.Request.Context, or rewrites request for inner http.Handler
implmentations, or both.

The current collection includes:
* [logcontext]: put go-kit's [Logger][kitlog.Logger] into context.
* [gormcontext]: put [*gorm.DB][gorm.DB] into context.

[middleware.Chain]: https://godoc.org/github.com/go-midway/midway#Chain
[funconv]: https://godoc.org/github.com/go-midway/midway/funconv
[logcontext]: https://godoc.org/github.com/go-midway/midway/logcontext
[gormcontext]: https://godoc.org/github.com/go-midway/midway/db/gormcontext
[kitlog.Logger]: https://godoc.org/github.com/go-kit/kit/log#Logger
[gorm.DB]: https://godoc.org/github.com/jinzhu/gorm#DB


## Similar and Interoperable Projects

* [alice][justinas/alice]
* [interpose][carbocation/interpose]
* [chi][go-chi/chi]

[justinas/alice]: https://github.com/justinas/alice
[carbocation/interpose]: https://github.com/carbocation/interpose
[go-chi/chi]: https://github.com/go-chi/chi
