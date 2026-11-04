// Command module-kit is the authors' CLI of NovaTrail's module format. In time it checks a
// module and builds its package and the module index (architecture 3.7 in spec); for now it
// only tells its version, so that the shared Go checks of .github have code to run on (R04).
package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"unicode"
)

// version is the release a build comes from: the release's own version, which release-files
// sets at link time (R07), or dev until something says otherwise.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run carries out a command line and returns its exit status: 0 when it did what was asked,
// 2 when it was asked for something it does not do. --version tells the release, and the version of
// Unicode that the normalization of the check kinds follows, which the format leaves to the release
// (format/README.md, section 15; R27). That is verdict.UnicodeVersion, the tests hold it equal to
// unicode.Version, and the CLI names the latter until it checks modules: so its binary carries no
// code of golang.org/x/text before the licenses of that code go with a release.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintf(stdout, "module-kit %s\nUnicode %s\n", currentVersion(), unicode.Version)
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
