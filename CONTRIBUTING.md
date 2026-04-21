# Contributing to dockpose

Thanks for your interest. This doc is minimal for pre-v0.1.0; it will grow.

## Dev loop

Requirements: Go 1.25+, Docker, `golangci-lint`, `make`.

```sh
git clone https://github.com/cwklurks/dockpose.git
cd dockpose
make check   # fmt + vet + lint + test + build
```

Run locally:

```sh
make build
./dockpose --version
```

## Running tests

```sh
go test -race ./...
```

## Commit message format

[Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>
```

Common types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`.

Do not use em dashes in commit messages. Prefer hyphens, commas, or semicolons.

## Pull requests

1. Fork, branch, commit.
2. `make check` passes locally.
3. Open a PR, fill the template, link the related issue.
4. CI must be green before merge.

## Scope

v0.1.0 is focused. Please read `DOCKPOSE_SPEC.md` (once public) or open a discussion before proposing large features. Explicit non-goals for v1: web UI, Kubernetes/Swarm, compose-file editor, plugin system.
