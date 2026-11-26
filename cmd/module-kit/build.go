package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/altnovatrail/module-kit/modulepackage"
)

// The modes of the files and folders the CLI writes, before the process's umask: their owner may
// change them and everyone may read them, since a package or an index is written to be published.
const (
	worldReadableFile   fs.FileMode = 0o644
	worldReadableFolder fs.FileMode = 0o755
)

// build builds the package of the module whose folder is the first argument into the folder that
// is the second, out, as files: out must not exist yet, or be an empty folder. A module that the
// format refuses has no package, and its report goes to stdout as check writes it. The files are
// written in a folder of their own beside out and moved to out once all are written, so that a
// failure leaves no part of a package behind.
func build(ctx context.Context, args []string, stdout io.Writer) (int, error) {
	if len(args) != 2 {
		return statusFailed, usageOf(nil)
	}
	dir, out := args[0], args[1]
	fsys, err := moduleFolder(dir)
	if err != nil {
		return statusFailed, err
	}
	if err := emptyOrAbsent(out); err != nil {
		return statusFailed, err
	}
	pkg, err := modulepackage.Build(ctx, fsys)
	var refusal *modulepackage.Refusal
	if errors.As(err, &refusal) {
		writeReport(stdout, dir, refusal.Report, nil)
		return statusRefused, nil
	}
	if err != nil {
		return statusFailed, err
	}
	if err := writePackage(out, pkg.Files); err != nil {
		return statusFailed, err
	}
	fmt.Fprintf(stdout, "%s: the package of %s %s, %s, is in %s\n", dir, pkg.Inventory.ID, pkg.Inventory.Version, pkg.Digest, out)
	return statusAccepted, nil
}

// emptyOrAbsent says, by no error, that out does not exist or is an empty folder.
func emptyOrAbsent(out string) error {
	entries, err := os.ReadDir(out)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	case len(entries) > 0:
		return fmt.Errorf("%s is not empty: the command writes into a folder that does not exist yet, or an empty one", out)
	}
	return nil
}

// writePackage writes the files of a package — or of the open views, which open writes alike — by
// their paths in its folder, into a folder of its own within a new temporary folder beside out, and
// moves it to out, in place of an empty folder there. The temporary folder, which only its owner may
// read, is removed whatever happens; the package's folder is made as every folder of it is, under
// the process's umask.
func writePackage(out string, files map[string][]byte) (err error) {
	out = filepath.Clean(out)
	temp, err := os.MkdirTemp(filepath.Dir(out), "."+filepath.Base(out)+".")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(temp)) }()
	built := filepath.Join(temp, "package")
	if err := os.MkdirAll(built, worldReadableFolder); err != nil {
		return err
	}
	if err := writeFiles(built, files); err != nil {
		return err
	}
	if err := os.Remove(out); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.Rename(built, out)
}

func writeFiles(root string, files map[string][]byte) error {
	for name, data := range files {
		if !fs.ValidPath(name) || strings.Contains(name, `\`) {
			return fmt.Errorf("a file at %q is no path in the folder", name)
		}
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), worldReadableFolder); err != nil {
			return err
		}
		if err := os.WriteFile(file, data, worldReadableFile); err != nil {
			return err
		}
	}
	return nil
}
