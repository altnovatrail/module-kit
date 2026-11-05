// Command module-kit is the authors' CLI of NovaTrail's module format (architecture 3.7 in spec;
// architecture 3.6): check checks a module and build builds its package as files, by the library of
// this module. The index of modules, the comparison of versions and the open views come with their
// tasks of RUN.md.
//
// Every command exits with 0 when it did what was asked, 1 when the format refuses its input, with
// the ID of each rule broken, and 2 when it could not do its work: a call it does not take, a file it
// could not read or write, or a stop by a signal. It reads and writes only the files it is given, and
// never the network.
package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/verdict"
)

// version is the release a build comes from: the release's own version, which release-files
// sets at link time (R07), or dev until something says otherwise.
var version = "dev"

// The exit statuses, one meaning for every command (architecture 3.6).
const (
	statusAccepted = 0
	statusRefused  = 1
	statusFailed   = 2
)

const usage = `usage: module-kit check <module>
       module-kit build <module> <folder>
       module-kit --version
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	status := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	os.Exit(status)
}

// run carries out a command line with ctx, which a signal makes done, and returns its exit status.
// --version tells the release, and the version of Unicode that the normalization of the check kinds
// follows, which the format leaves to the release (format/README.md, section 15; R27).
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	switch {
	case len(args) == 1 && args[0] == "--version":
		fmt.Fprintf(stdout, "module-kit %s\nUnicode %s\n", currentVersion(), verdict.UnicodeVersion)
		return statusAccepted
	case len(args) == 2 && args[0] == "check":
		return check(ctx, args[1], stdout, stderr)
	case len(args) == 3 && args[0] == "build":
		return build(ctx, args[1], args[2], stdout, stderr)
	}
	fmt.Fprint(stderr, usage)
	return statusFailed
}

// check checks the module whose folder is dir and writes its report to stdout: a line for each
// breach, its place from dir, the ID of its rule and what breaks it, in the order of their places,
// and a last line that says whether the format accepts the module.
func check(ctx context.Context, dir string, stdout, stderr io.Writer) int {
	fsys, err := moduleFolder(dir)
	if err != nil {
		return failed(stderr, "check", err)
	}
	report, err := module.Check(ctx, fsys)
	if err != nil {
		return failed(stderr, "check", err)
	}
	if len(report.Breaches) > 0 {
		writeReport(stdout, dir, report)
		return statusRefused
	}
	fmt.Fprintf(stdout, "%s: the module breaks no rule of the format %s\n", dir, module.FormatVersion)
	return statusAccepted
}

// moduleFolder gives the folder dir as an fs.FS, and an error when it is no folder.
func moduleFolder(dir string) (fs.FS, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is no folder: a module is the folder of its files", dir)
	}
	return os.DirFS(dir), nil
}

// writeReport writes the breaches of a report, each with its path from dir as the command line gave
// it, and a last line that counts them. A breach of the module's folder itself has the path dir.
func writeReport(w io.Writer, dir string, report module.Report) {
	for _, b := range report.Breaches {
		b.Path = filepath.ToSlash(filepath.Join(dir, filepath.FromSlash(b.Path)))
		fmt.Fprintln(w, b)
	}
	noun := "breaches"
	if len(report.Breaches) == 1 {
		noun = "breach"
	}
	fmt.Fprintf(w, "%s: the format %s refuses the module: %d %s\n", dir, module.FormatVersion, len(report.Breaches), noun)
}

// failed writes why a command could not do its work, and gives its status.
func failed(stderr io.Writer, command string, err error) int {
	fmt.Fprintf(stderr, "module-kit %s: %v\n", command, err)
	return statusFailed
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
