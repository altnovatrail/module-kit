package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/modulepackage"
)

// errUsage is a command line that a command does not take, with nothing more to say of it than the
// usage.
var errUsage = errors.New("usage")

// A usageError is a command line that a command does not take: its usage goes to stderr, after
// err when there is one, and the status is 2.
type usageError struct{ err error }

func (e *usageError) Error() string { return errors.Join(errUsage, e.err).Error() }

// usageOf says that a command does not take its line, by err, the error of its flags: nil, or one
// that is errUsage, says nothing more than the usage.
func usageOf(err error) error {
	if errors.Is(err, errUsage) {
		err = nil
	}
	return &usageError{err}
}

// newFlags gives the flags of a command, which say nothing themselves and so need no name: a
// command that cannot parse its line returns the error by usageOf, and run writes it after the
// command's name.
func newFlags() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	return f
}

// optional is a flag with a value that a command line may give once, and whose being given is what
// it says, as --previous: an empty value is given too.
type optional struct {
	value string
	set   bool
}

func (o *optional) String() string { return o.value }

func (o *optional) Set(value string) error {
	if o.set {
		return errors.New("given twice")
	}
	o.value, o.set = value, true
	return nil
}

type repeated []string

func (r *repeated) String() string { return strings.Join(*r, ",") }

func (r *repeated) Set(value string) error {
	*r = append(*r, value)
	return nil
}

// oneOf says that exactly one of two choices is made: a flag with a value, or a flag that says there
// is none, as --previous and --first. Neither, or both, is no command line.
func oneOf(value *optional, none bool) error {
	if value.set == none {
		return errUsage
	}
	if value.set && value.value == "" {
		return errors.New("an empty path names no file: say that there is none by its own flag")
	}
	return nil
}

// readPackage reads the package as files in the folder dir by the chain of digests from the digest of
// the bytes of its inventory.json, which is the digest of the package, with the files of device/ or
// without them.
func readPackage(dir string, withDevice bool) (*modulepackage.Package, error) {
	fsys, err := moduleFolder(dir)
	if err != nil {
		return nil, err
	}
	inventory, err := fs.ReadFile(fsys, modulepackage.InventoryFile)
	if err != nil {
		return nil, fmt.Errorf("%s holds no package: %w", dir, err)
	}
	read := modulepackage.ReadParts
	if withDevice {
		read = modulepackage.Read
	}
	pkg, err := read(fsys, canonical.Digest(inventory))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	return pkg, nil
}

// writeNew writes data to the file out, which must not exist yet: into a file of its own within a new
// temporary folder beside out, which is moved to out once it is whole, so that a failure leaves no
// part of the file behind. The temporary folder, which only its owner may read, is removed whatever
// happens; the file is made as build makes the files of a package, under the process's umask.
func writeNew(out string, data []byte) (err error) {
	out = filepath.Clean(out)
	if _, err := os.Lstat(out); err == nil {
		return fmt.Errorf("%s exists already: the file is written where none is", out)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	temp, err := os.MkdirTemp(filepath.Dir(out), "."+filepath.Base(out)+".")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(temp)) }()
	written := filepath.Join(temp, "file")
	if err := os.WriteFile(written, data, worldReadableFile); err != nil {
		return err
	}
	if _, err := os.Lstat(out); err == nil {
		return fmt.Errorf("%s came to exist while it was written: the file is written where none is", out)
	}
	return os.Rename(written, out)
}
