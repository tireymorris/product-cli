# Product CLI

Product CLI turns a product goal into a concise, outcome-focused Markdown plan and prints it to stdout. It is the product-planning half of Ralph extracted into a standalone tool.

## Usage

```bash
product-cli "let travelers share saved itineraries"
product-cli --headless "let travelers share saved itineraries"
```

The AI runner is selected with `PRODUCT_RUNNER` (`claude`, `cursor`, `pi`, `opencode`, or `copilot`; default `claude`). `--runner` and `--timeout` provide command-line overrides.

The runner receives a product-planning prompt and returns Markdown containing:

- the user problem and value
- prioritized, user-facing outcomes
- observable success signals
- open questions

Product CLI does not write `prd.json`, `product.json`, or any other product artifact. It does not implement code.

## Development

```bash
go test ./...
go build .
```
