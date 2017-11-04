# supervisor

### note

Use package [retry](https://github.com/dc0d/retry) instead, which has a simpler and more strict semantics. This package contains code smells that are remnants of it's history in trying to implement supervisor trees in Erlang.

A try in implementing supervision trees (as in Erlang; look into git history), that ended up in a function!

See [documents](https://godoc.org/github.com/dc0d/supervisor).

Get it via:

```bash
$ go get -u gopkg.in/dc0d/supervisor.v1
```

Or for development version:

```bash
$ go get -u github.com/dc0d/supervisor
```
