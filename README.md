# ez-mig

A UX wrapper around [golang-migrate](https://github.com/golang-migrate/migrate) for lazy people.

## Why

`go migrate` works, but it's too verbose for everyday use. `ez-mig` wraps it with shorter commands and a session system so you're not copy-pasting connection strings every time.

```bash
# the old way
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/app?sslmode=disable" up

# the ez way
ez-mig up --db app
```

## Features

- **Short commands** — `up`, `down`, `goto`, `force`, `version`
- **Session management** — save your DB connection + migration path under a name, reuse it forever
- **Use it anywhere** — install globally and run migrations outside the project directory

## Prerequisites

- Linux, macOS, or Windows
- PostgreSQL or MySQL *(more databases coming soon)*
- Go *(only if building from source)*

## Installation

**From release** *(recommended)*

Download the latest binary from the [releases page](https://github.com/AkhileshThykkat/ez-mig/releases).

**From source**

```bash
git clone https://github.com/AkhileshThykkat/ez-mig.git
cd ez-mig
go build
```

## Usage

See [Commands.md](./Commands.md) for the full reference. Quick start:

```bash
# save a session
ez-mig config add local \
  --uri "postgres://user:pass@localhost:5432/app?sslmode=disable" \
  --path "./migrations"

# run migrations
ez-mig up --db local
```
