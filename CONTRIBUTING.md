# Contributing to alfred

Thanks for your interest in alfred. This document explains how to set up
the project locally, run the test suite, and submit a pull request.

## Reporting issues

Open an issue at
[`sachncs/alfred/issues`](https://github.com/sachncs/alfred/issues)
using the appropriate template (`bug`, `feature_request`, or `question` via
Discussions if enabled). For security issues, follow
[`SECURITY.md`](./SECURITY.md).

## Development setup

(no extra setup — Go toolchain required)

## Tests

```
go test ./...
```

## Lint / format

```
go vet ./...
```

## Pull request flow

1. Fork the repository.
2. Create a topic branch off `main` (use linear history).
3. Make focused commits with clear messages.
4. Ensure `tests`, `lint`, and `format` all pass.
5. Use the [PR template](./.github/PULL_REQUEST_TEMPLATE.md).
6. Push the branch and open a pull request targeting `main`.

By submitting a pull request, you agree to follow the
[Code of Conduct](./CODE_OF_CONDUCT.md).
