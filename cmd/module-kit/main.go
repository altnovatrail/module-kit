// Command module-kit is the authors' CLI of NovaTrail's module format. In time it checks a
// module and builds its package and the module index (architecture 3.7 in spec); for now it
// only tells its version, so that the shared Go checks of .github have code to run on (R04).
package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

// version is the release a build comes from: the release's own version, which release-files
// sets at link time (R07), or dev until something says otherwise.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run carries out a command line and returns its exit status: 0 when it did what was asked,
// 2 when it was asked for something it does not do.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintf(stdout, "module-kit %s\n", currentVersion())
		return 0
	}
	fmt.Fprintln(stderr, "usage: module-kit --version")
	return 2
}

// currentVersion is version when a release set it. Otherwise it is the version of the module the
// binary was built from, as go install module-kit/cmd/module-kit@v0.1.0 records it, and dev when
// there is none: a build from a checkout without git, or one that records (devel).
func currentVersion() string {
	if version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return version
	}
	return info.Main.Version
}
