package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/semver"
	"github.com/altnovatrail/module-kit/module"
	"github.com/altnovatrail/module-kit/moduleindex"
	"github.com/altnovatrail/module-kit/modulepackage"
	"github.com/altnovatrail/module-kit/openview"
	"github.com/altnovatrail/module-kit/template"
)

// The file of the open views that open writes, and what its root says it is.
const (
	openViewsFile    = "open-views.json"
	openViewsFormat  = "module-kit-open-views"
	openViewsVersion = 0
	modulesFolder    = "modules"
)

// openViews writes the open views of a release of modules, for the build of the page of samples of
// canvas (R25 spec; R15; R43):
//
//	module-kit open --index <module-index.json> --latest <module-index.json>
//	    --seed N [--seed N]... --out <folder> <package folder>...
//
// --index is the index of the release the page is built from, and --latest the index of the newest
// release, which alone knows of every withdrawal so far: the same file when the release is the
// newest. Of each module of the release the command takes the version that a new attempt would,
// the newest by semver that the newest index does not withdraw and that lies in the format's
// window (moduleindex.WithLatest, moduleindex.Newest), and names on stdout each newer version it
// skips. Each version's package is the folder whose inventory.json has the digest of its entry,
// read by the chain of digests with its device part.
//
// Into out, which does not exist yet or is empty, goes open-views.json: of each module its ID, its
// version, its format version, the release that published it, the digest of its package, the open
// view of its manifest and its tasks in the order of their IDs, each the open view and its digest
// (R93 spec). A template gives an instance for each seed, in the order of the seeds, with the seed
// beside it: the instance an attempt of that seed gets (format/README.md, section 23). Beside it go
// modules/<ID>/<version>/ with the files of the device part, LICENSE, NOTICE and media/, byte for
// byte as the chain of digests checked them. The command checks no rule of the format, so it exits
// 0 or 2.
func openViews(_ context.Context, args []string, stdout, stderr io.Writer) int {
	f := newFlags("open")
	indexFile := f.String("index", "", "")
	latestFile := f.String("latest", "", "")
	var seeds seedList
	f.Var(&seeds, "seed", "")
	out := f.String("out", "", "")
	if err := f.Parse(args); err != nil {
		return usageOf(stderr, "open", err)
	}
	if *indexFile == "" || *latestFile == "" || len(seeds) == 0 || *out == "" {
		return usageOf(stderr, "open", nil)
	}
	release, err := readIndex("the index", *indexFile)
	if err != nil {
		return failed(stderr, "open", err)
	}
	latest, err := readIndex("the newest index", *latestFile)
	if err != nil {
		return failed(stderr, "open", err)
	}
	index, err := moduleindex.WithLatest(release, latest)
	if err != nil {
		return failed(stderr, "open", fmt.Errorf("%s as the newest index of %s: %w", *latestFile, *indexFile, err))
	}
	if err := emptyOrAbsent(*out); err != nil {
		return failed(stderr, "open", err)
	}
	byDigest, err := packagesByDigest(f.Args())
	if err != nil {
		return failed(stderr, "open", err)
	}
	slices.Sort(seeds)
	files := map[string][]byte{}
	modules := []any{}
	for _, id := range slices.Sorted(maps.Keys(index.Modules)) {
		version, ok := moduleindex.Newest(index, id)
		writeSkipped(stdout, id, index.Modules[id], version, ok)
		if !ok {
			continue
		}
		entry := index.Modules[id][version]
		dir, found := byDigest[entry.PackageDigest]
		if !found {
			return failed(stderr, "open", fmt.Errorf("the index names %s %s, with the package %s, and no package folder holds it",
				id, version, entry.PackageDigest))
		}
		pkg, err := modulepackage.Read(os.DirFS(dir), entry.PackageDigest)
		if err != nil {
			return failed(stderr, "open", fmt.Errorf("%s: %w", dir, err))
		}
		if inv := pkg.Inventory; inv.ID != id || inv.Version != version || inv.ModuleFormatVersion != entry.ModuleFormatVersion {
			return failed(stderr, "open", fmt.Errorf("the index names the package %s as %s %s of the format %s, and it is %s %s of the format %s",
				pkg.Digest, id, version, entry.ModuleFormatVersion, inv.ID, inv.Version, inv.ModuleFormatVersion))
		}
		views, count, err := moduleViews(pkg, entry, seeds)
		if err != nil {
			return failed(stderr, "open", fmt.Errorf("%s %s: %w", id, version, err))
		}
		modules = append(modules, views)
		for name, data := range pkg.Files {
			if rest, ok := strings.CutPrefix(name, modulepackage.DeviceFolder); ok {
				files[strings.Join([]string{modulesFolder, string(id), string(version), rest}, "/")] = data
			}
		}
		fmt.Fprintf(stdout, "%s %s: %d %s\n", id, version, count, plural(count, "open view of a task", "open views of tasks"))
	}
	text, err := canonical.Text(map[string]any{
		"format":                openViewsFormat,
		"version":               int64(openViewsVersion),
		"module_format_version": module.FormatVersion,
		"modules":               modules,
	})
	if err != nil {
		return failed(stderr, "open", err)
	}
	files[openViewsFile] = text
	if err := writePackage(*out, files); err != nil {
		return failed(stderr, "open", err)
	}
	fmt.Fprintf(stdout, "the open views of %d %s are in %s\n", len(modules), plural(len(modules), "module", "modules"), *out)
	return statusAccepted
}

// moduleViews gives the member of open-views.json of one version of a module: what names it, the
// open view of its manifest, and an open view of each task, a template's for each seed; and how
// many open views of tasks it holds.
func moduleViews(pkg *modulepackage.Package, entry formatv0.Entry, seeds seedList) (map[string]any, int, error) {
	manifest, err := openview.ManifestView(pkg.View.Manifest)
	if err != nil {
		return nil, 0, err
	}
	tasks := []any{}
	for _, id := range slices.Sorted(maps.Keys(pkg.View.Tasks)) {
		task := pkg.View.Tasks[id]
		if task.Template == nil {
			view, digest, err := openview.TaskView(task)
			if err != nil {
				return nil, 0, err
			}
			tasks = append(tasks, map[string]any{"open_view_digest": string(digest), "task": view})
			continue
		}
		numbers := pkg.View.Instances[id]
		for _, seed := range seeds {
			picked, err := template.AttemptInstance(seed, id, len(numbers))
			if err != nil {
				return nil, 0, fmt.Errorf("the template %s: %w", id, err)
			}
			instance, err := template.Instance(task, numbers[picked])
			if err != nil {
				return nil, 0, err
			}
			view, digest, err := openview.TaskView(instance)
			if err != nil {
				return nil, 0, err
			}
			tasks = append(tasks, map[string]any{
				"open_view_digest": string(digest),
				"seed":             strconv.FormatUint(seed, 10),
				"task":             view,
			})
		}
	}
	return map[string]any{
		"id":                    string(pkg.Inventory.ID),
		"version":               string(pkg.Inventory.Version),
		"module_format_version": string(pkg.Inventory.ModuleFormatVersion),
		"release":               string(entry.Release),
		"package_digest":        string(pkg.Digest),
		"manifest":              manifest,
		"tasks":                 tasks,
	}, len(tasks), nil
}

// writeSkipped names on w each version of the module id newer than the one taken, or every version
// when none is, with why Newest skips it: withdrawn, by its reason; of a format version outside
// module-kit's window; or not in the form X.Y.Z, which no index that moduleindex.Read gives has.
func writeSkipped(w io.Writer, id formatv0.ModuleID, entries map[formatv0.ModuleVersion]formatv0.Entry, taken formatv0.ModuleVersion, ok bool) {
	versions := slices.Collect(maps.Keys(entries))
	slices.SortFunc(versions, func(a, b formatv0.ModuleVersion) int { return compareVersions(b, a) })
	for _, version := range versions {
		if ok && compareVersions(version, taken) <= 0 {
			break
		}
		entry := entries[version]
		switch _, err := semver.Parse(string(version)); {
		case entry.Withdrawn != nil:
			fmt.Fprintf(w, "%s %s: skipped, withdrawn (%s)\n", id, version, *entry.Withdrawn)
		case !module.ReadsFormat(string(entry.ModuleFormatVersion)):
			fmt.Fprintf(w, "%s %s: skipped, of the format %s: module-kit reads the format %s\n", id, version, entry.ModuleFormatVersion, module.FormatVersion)
		case err != nil:
			fmt.Fprintf(w, "%s %s: skipped, not a version X.Y.Z\n", id, version)
		}
	}
	if !ok {
		fmt.Fprintf(w, "%s: no version to open\n", id)
	}
}

// compareVersions orders two versions of a module by semver, and one not in the form X.Y.Z, which no
// index that moduleindex.Read gives names, by its text.
func compareVersions(a, b formatv0.ModuleVersion) int {
	x, errX := semver.Parse(string(a))
	y, errY := semver.Parse(string(b))
	if errX != nil || errY != nil {
		return strings.Compare(string(a), string(b))
	}
	return x.Compare(y)
}

// readIndex reads an index file by the places every version of the format keeps.
func readIndex(what, file string) (formatv0.ModuleIndex, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return formatv0.ModuleIndex{}, fmt.Errorf("%s: %w", what, err)
	}
	index, err := moduleindex.Read(data)
	if err != nil {
		return formatv0.ModuleIndex{}, fmt.Errorf("%s, %s: %w", what, file, err)
	}
	return index, nil
}

// seedList is --seed, given again and again: each a seed of the format, an unsigned 64-bit integer
// in decimal digits with no leading zero (format/README.md, section 22), and each once.
type seedList []uint64

func (s *seedList) String() string {
	texts := make([]string, len(*s))
	for i, seed := range *s {
		texts[i] = strconv.FormatUint(seed, 10)
	}
	return strings.Join(texts, ",")
}

func (s *seedList) Set(value string) error {
	seed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || strconv.FormatUint(seed, 10) != value {
		return errors.New("a seed is an unsigned 64-bit integer in decimal digits, with no leading zero")
	}
	if slices.Contains(*s, seed) {
		return fmt.Errorf("the seed %s is given twice", value)
	}
	*s = append(*s, seed)
	return nil
}
