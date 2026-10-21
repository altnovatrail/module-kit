// Command module-kit is the authors' CLI of NovaTrail's module format. In time it checks a
// module and builds its package and the module index (architecture 3.7 in spec); for now it
// only tells its version, so that the shared Go checks of .github have code to run on (R04).
package main

import (
	"fmt"
	"io"
	"os"
)

// version is the release a build comes from: dev in a build from a checkout, and the release's
// own version once a release sets it at link time.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run carries out a command line and returns its exit status: 0 when it did what was asked,
// 2 when it was asked for something it does not do.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintf(stdout, "module-kit %s\n", version)
		return 0
	}
	fmt.Fprintln(stderr, "usage: module-kit --version")
	return 2
}
