package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/internal/yaml"
	"github.com/altnovatrail/module-kit/moduleindex"
)

// index builds the module index of a release of modules within the shared release (format/README.md,
// section 26; R63 spec; R40):
//
//	module-kit index (--previous <module-index.json> | --first) --tag vX.Y.Z
//	    (--withdrawals <withdrawals.yaml> | --no-withdrawals) --out <file> <package folder>...
//
// What there is not — a previous index, withdrawals — the command hears by a flag of its own, not by
// a file it could not find: --previous with an empty path is no command line, so a release whose
// previous index was not handed over fails rather than start a new index. Each package folder is a
// package as files, as build writes it, read by the chain of digests from its inventory.json. The
// index goes to out, which must not exist yet. A release that breaks a rule of the index exits 1
// with every breach on stdout; one whose previous index, tag, withdrawals or packages cannot be read
// exits 2.
func index(_ context.Context, args []string, stdout, stderr io.Writer) int {
	f := newFlags("index")
	var previous, withdrawals optional
	f.Var(&previous, "previous", "")
	first := f.Bool("first", false, "")
	tag := f.String("tag", "", "")
	f.Var(&withdrawals, "withdrawals", "")
	noWithdrawals := f.Bool("no-withdrawals", false, "")
	out := f.String("out", "", "")
	if err := f.Parse(args); err != nil {
		return usageOf(stderr, "index", err)
	}
	if err := errors.Join(oneOf(&previous, *first), oneOf(&withdrawals, *noWithdrawals)); err != nil || *tag == "" || *out == "" || f.NArg() == 0 {
		return usageOf(stderr, "index", err)
	}
	release := moduleindex.Release{First: *first, Tag: formatv0.Release(*tag)}
	if previous.set {
		data, err := os.ReadFile(previous.value)
		if err != nil {
			return failed(stderr, "index", fmt.Errorf("the previous index: %w", err))
		}
		release.Previous = data
	}
	if withdrawals.set {
		w, err := readWithdrawals(withdrawals.value)
		if err != nil {
			return failed(stderr, "index", err)
		}
		release.Withdrawals = w
	}
	for _, dir := range f.Args() {
		pkg, err := readPackage(dir, true)
		if err != nil {
			return failed(stderr, "index", err)
		}
		release.Packages = append(release.Packages, pkg)
	}
	text, err := moduleindex.Build(release)
	var refusal *moduleindex.Refusal
	if errors.As(err, &refusal) {
		for _, b := range refusal.Breaches {
			fmt.Fprintln(stdout, b)
		}
		fmt.Fprintf(stdout, "the release %s breaks the rules of the module index, so it has no index: %d %s\n",
			*tag, len(refusal.Breaches), plural(len(refusal.Breaches), "breach", "breaches"))
		return statusRefused
	}
	if err != nil {
		return failed(stderr, "index", err)
	}
	if err := writeNew(*out, text); err != nil {
		return failed(stderr, "index", err)
	}
	fmt.Fprintf(stdout, "the module index of %s, of %d %s, is in %s\n", *tag, len(release.Packages),
		plural(len(release.Packages), "package", "packages"), *out)
	return statusAccepted
}

// readWithdrawals reads the withdrawals of a release, withdrawals.yaml of modules: YAML by the rules
// of section 4 of the format, in the form of $defs/withdrawals of the schema of the module index
// (format/README.md, sections 4 and 26). An empty file withdraws nothing, and is no file of
// withdrawals: a release that withdraws none says so by --no-withdrawals.
func readWithdrawals(file string) (formatv0.Withdrawals, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("the withdrawals: %w", err)
	}
	value, err := yaml.Read(data)
	if err != nil {
		return nil, fmt.Errorf("the withdrawals, %s: %w", file, err)
	}
	if errs := schema.Validate("module-index.schema.json#/$defs/withdrawals", value); len(errs) > 0 {
		return nil, fmt.Errorf("the withdrawals, %s, are not in the form of $defs/withdrawals of module-index.schema.json: %s %s: %s",
			file, errs[0].Pointer(), errs[0].Keyword, errs[0].Message)
	}
	w, err := schema.Decode[formatv0.Withdrawals](value)
	if err != nil {
		return nil, fmt.Errorf("the withdrawals, %s: %w", file, err)
	}
	return w, nil
}

// plural gives the noun that counts n.
func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// usageOf writes why a command line is not one the command takes, if it says, and the usage, and
// gives status 2.
func usageOf(stderr io.Writer, command string, err error) int {
	if err != nil && !errors.Is(err, errUsage) {
		fmt.Fprintf(stderr, "module-kit %s: %v\n", command, err)
	}
	fmt.Fprint(stderr, usage)
	return statusFailed
}
