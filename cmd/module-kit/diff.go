package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/modulepackage"
	"github.com/altnovatrail/module-kit/moduleversion"
)

// diff compares a new version of a module with every published version of it (format/README.md,
// section 27; architecture 3.5; R40):
//
//	module-kit diff [--json] (--index <module-index.json> | --first) [--published <package folder>]... <module>
//
// <module> is the folder of a module, which diff checks and builds in memory, or a package as files
// that build wrote; manifest.yaml or inventory.json at its root tells which. --index is the index of
// the newest release of modules, which CI has checked by release-verify; the module's versions it
// names are the published ones, and a module it does not name has none. --first says that there is
// no index yet. Each --published is a package as files, found by the digest of its inventory.json:
// the package of every version the index names is needed, and others are skipped. Exit 0 when the
// new version is accepted or the module has not changed, 1 when a rule refuses it — of the check of
// the module or of versions —, 2 when it cannot be compared.
func diff(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	f := newFlags("diff")
	asJSON := f.Bool("json", false, "")
	var indexFile optional
	f.Var(&indexFile, "index", "")
	first := f.Bool("first", false, "")
	var published repeated
	f.Var(&published, "published", "")
	if err := f.Parse(args); err != nil {
		return usageOf(stderr, "diff", err)
	}
	if err := oneOf(&indexFile, *first); err != nil || f.NArg() != 1 {
		return usageOf(stderr, "diff", err)
	}
	dir := f.Arg(0)
	next, report, err := newVersion(ctx, dir)
	if err != nil {
		return failed(stderr, "diff", err)
	}
	if next == nil {
		if *asJSON {
			return writeJSON(stdout, stderr, map[string]any{"check": breachesOf(dir, report)}, statusRefused)
		}
		writeReport(stdout, dir, report, nil)
		return statusRefused
	}
	var versions []moduleversion.Published
	if indexFile.set {
		if versions, err = publishedOf(indexFile.value, next.Inventory.ID, published); err != nil {
			return failed(stderr, "diff", err)
		}
	}
	c, err := moduleversion.Compare(next, versions)
	if err != nil {
		return failed(stderr, "diff", err)
	}
	status := statusAccepted
	if c.Refused() {
		status = statusRefused
	}
	if *asJSON {
		return writeJSON(stdout, stderr, c, status)
	}
	fmt.Fprintln(stdout, c)
	return status
}

// newVersion reads the new version of a module from dir: a package as files when inventory.json
// stands at its root, read by the chain of digests without device/, which the comparison reads by
// its listing; otherwise the folder of a module, checked and built in memory. A module that its check
// refuses gives no package but its report.
func newVersion(ctx context.Context, dir string) (*modulepackage.Package, module.Report, error) {
	fsys, err := moduleFolder(dir)
	if err != nil {
		return nil, module.Report{}, err
	}
	if _, err := fs.Stat(fsys, modulepackage.InventoryFile); err == nil {
		pkg, err := readPackage(dir, false)
		return pkg, module.Report{}, err
	}
	if _, err := fs.Stat(fsys, "manifest.yaml"); err != nil {
		return nil, module.Report{}, fmt.Errorf("%s holds neither manifest.yaml nor %s: it is neither a module nor a package", dir, modulepackage.InventoryFile)
	}
	pkg, err := modulepackage.Build(ctx, fsys)
	var refusal *modulepackage.Refusal
	if errors.As(err, &refusal) {
		return nil, refusal.Report, nil
	}
	return pkg, module.Report{}, err
}

// A checkBreach is a breach of the check of a module as diff --json writes it.
type checkBreach struct {
	Rule    string `json:"rule,omitzero"`
	Keyword string `json:"keyword,omitzero"`
	Path    string `json:"path"`
	Line    int    `json:"line,omitzero"`
	Column  int    `json:"column,omitzero"`
	Pointer string `json:"pointer,omitzero"`
	Message string `json:"message"`
}

// breachesOf gives the breaches of a report, each with its path from dir as the command line gave it.
func breachesOf(dir string, report module.Report) []checkBreach {
	out := make([]checkBreach, len(report.Breaches))
	for i, b := range report.Breaches {
		out[i] = checkBreach{b.Rule, b.Keyword, slashed(dir, b.Path), b.Line, b.Column, b.Pointer, b.Message}
	}
	return out
}

// writeJSON writes v as indented JSON and a newline, and gives status.
func writeJSON(stdout, stderr io.Writer, v any, status int) int {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return failed(stderr, "diff", err)
	}
	fmt.Fprintln(stdout, strings.TrimSpace(string(data)))
	return status
}
