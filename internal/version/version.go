package version

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
	Ref     = "unknown"
)

func Info() string {
	return fmt.Sprintf("product-cli version=%s commit=%s ref=%s", Version, Commit, Ref)
}
