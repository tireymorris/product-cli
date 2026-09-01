# Product CLI

Product CLI turns a product goal into a concise, outcome-focused Markdown plan and prints it to stdout. It does not write product files or implement code.

## Install

```bash
go install github.com/tireymorris/product-cli@latest
```

Requires Go 1.26+, Git, and one supported AI runner on `PATH`.

From a checkout:

```bash
go install .
# or build a binary at a specific path:
scripts/build.sh -o "$(go env GOPATH)/bin/product-cli"
```

## Usage

```bash
product-cli "let teams track customer feedback"
product-cli --headless "let teams track customer feedback"
```

The runner is selected with `PRODUCT_RUNNER` (`claude`, `cursor`, `pi`, `opencode`, or `copilot`; default `claude`). Use `--runner` to override it and `--timeout` to limit its runtime.

The generated plan contains:

- the user problem and value
- prioritized, user-facing outcomes
- observable success signals
- open questions

## Development

```bash
go test ./...
go build .
```
