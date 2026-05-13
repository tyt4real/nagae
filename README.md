# nagae - a great, astounding and tremendous webmail client
![iku nagae](static/img/ikunagae.png)

The project iz named after Iku Nagae, the messenger of the Lunatic Kingdom

## Why though
I was scared that whatever happened to Roundcube at pissmail might happen to me so I decided to make my own webmail client, also I've never worked with email so it is a great learning experience I guess.

## Getting started

### Prerequisites

```bash
go install github.com/a-h/templ/cmd/templ@latest
# for hot reload if you want it i guess:
go install github.com/air-verse/air@latest
```

### Install dependencies

```bash
go mod tidy
```

### Run (production)

```bash
make run
```

### Run (hot reload dev)

```bash
make dev
```

Open http://localhost:8080 and log in with your email credentials.

## Environment variables

| Variable         | Default                   | Description              |
|-----------------|---------------------------|--------------------------|
| `PORT`           | `8080`                    | HTTP listen port         |
| `SESSION_SECRET` | `change-me-boss` | Cookie signing secret    |
