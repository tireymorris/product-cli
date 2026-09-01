# Product CLI

Product CLI turns a product goal into an outcome-focused `prd.json` document. It is the planning half of Ralph extracted into a standalone tool.

## Usage

```bash
product-cli "let travelers share saved itineraries"
product-cli --headless "let travelers share saved itineraries"
product-cli --resume
product-cli status
product-cli clean
```

The AI runner is selected with `PRODUCT_RUNNER` (`claude`, `cursor`, `pi`, `opencode`, or `copilot`; default `claude`). `PRODUCT_YOLO=1` skips clarification. `--runner` and `--timeout` provide command-line overrides.

Product CLI asks the runner to write only product-level outcomes. The resulting document contains stories and observable behaviors, but no implementation plans, tests, code paths, or refactoring instructions. Product CLI never implements source code.

The current artifact is `prd.json` with `"mode": "product"`, which remains compatible with Ralph's product-document format. Older `product.json` documents are migrated by `--resume`.

## Development

```bash
go test ./...
go build .
```
