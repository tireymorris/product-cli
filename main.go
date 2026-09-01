package main

import (
	"os"

	"product-cli/internal/app"
)

func main() { os.Exit(app.Run(os.Args[1:])) }
