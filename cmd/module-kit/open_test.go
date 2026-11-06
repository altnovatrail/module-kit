package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
	"github.com/altnovatrail/module-kit/modulepackage"
	"github.com/altnovatrail/module-kit/openview"
)

// vectorPackages builds the module of each test vector of packages named, by the command build, and
// gives the package folders by the vector's name.
func vectorPackages(t *testing.T, names ...string) map[string]string {
	t.Helper()
	folders := map[string]string{}
	for _, name := range names {
		module := fromSpec(t, path.Join("vectors/packages", name, "module"))
		out := filepath.Join(t.TempDir(), name)
		var stdout, stderr bytes.Buffer
		if status := run(context.Background(), []string{"build", module, out}, &stdout, &stderr); status != statusAccepted {
			t.Fatalf("build %s: status %d, %s%s", name, status, stdout.String(), stderr.String())
		}
		folders[name] = out
	}
	return folders
}

// allPackages is every test vector of packages.
func allPackages(t *testing.T) map[string]string {
	t.Helper()
	entries, err := fs.ReadDir(spec.Format, "vectors/packages")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return vectorPackages(t, names...)
}

// entry is a version of a module that a test's index names: by a package folder, or by a digest of a
// package that no folder holds.
type entry struct {
	folder, digest, format, withdrawn string
}

// indexText gives the text of a module index of the release v0.1.0 that names each entry under the ID
// and version of its package, or under the ones given.
func indexText(t *testing.T, entries map[string]map[string]entry) []byte {
	t.Helper()
	modules := map[string]any{}
	for id, versions := range entries {
		m := map[string]any{}
		for version, e := range versions {
			digest, format := e.digest, e.format
			if e.folder != "" {
				inventory, err := os.ReadFile(filepath.Join(e.folder, modulepackage.InventoryFile))
				if err != nil {
					t.Fatal(err)
				}
				digest = string(canonical.Digest(inventory))
			}
			if format == "" {
				format = "0.3"
			}
			value := map[string]any{"module_format_version": format, "package_digest": digest, "release": "v0.1.0"}
			if e.withdrawn != "" {
				value["withdrawn"] = e.withdrawn
			}
			m[version] = value
		}
		modules[id] = m
	}
	text, err := canonical.Text(map[string]any{"module_format_version": "0.3", "modules": modules})
	if err != nil {
		t.Fatal(err)
	}
	return text
}

// inventoryOf reads the ID and the version of the package in a folder.
func inventoryOf(t *testing.T, folder string) (id, version string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(folder, modulepackage.InventoryFile))
	if err != nil {
		t.Fatal(err)
	}
	var inventory formatv0.Inventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	return string(inventory.ID), string(inventory.Version)
}

// releaseOf gives the text of an index that names each package folder as its own ID and version.
func releaseOf(t *testing.T, folders map[string]string) []byte {
	t.Helper()
	entries := map[string]map[string]entry{}
	for _, folder := range folders {
		id, version := inventoryOf(t, folder)
		if entries[id] == nil {
			entries[id] = map[string]entry{}
		}
		entries[id][version] = entry{folder: folder}
	}
	return indexText(t, entries)
}

// runOpen runs open on an index as the release's and the newest, the seeds and the package folders,
// into a new folder, and gives the status, stdout, stderr and the folder.
func runOpen(t *testing.T, release, latest []byte, seeds []string, folders []string) (int, string, string, string) {
	t.Helper()
	releaseFile := writeTemp(t, "module-index.json", release)
	latestFile := releaseFile
	if latest != nil {
		latestFile = writeTemp(t, "latest.json", latest)
	}
	out := filepath.Join(t.TempDir(), "open")
	args := []string{"open", "--index", releaseFile, "--latest", latestFile, "--out", out}
	for _, seed := range seeds {
		args = append(args, "--seed", seed)
	}
	args = append(args, folders...)
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), args, &stdout, &stderr)
	return status, stdout.String(), stderr.String(), out
}

// openViewsOf reads open-views.json in a folder that open wrote, and holds it to its canonical text.
func openViewsOf(t *testing.T, out string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(out, openViewsFile))
	if err != nil {
		t.Fatal(err)
	}
	value, err := canonical.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if text, err := canonical.Text(value); err != nil || !bytes.Equal(text, data) {
		t.Errorf("open-views.json is not its canonical text: %v", err)
	}
	return value.(map[string]any)
}

// The open views of every test vector of packages: of each module, what names it and its manifest's
// open view, and of each task its open view and the digest of that view's canonical text; a template
// gives an instance for each seed, in the order of the seeds, with the seed beside it. Each open view
// is the one openview gives of the package's part.
func TestOpenPackages(t *testing.T) {
	folders := allPackages(t)
	seeds := []string{"20261027", "0"}
	status, stdout, stderr, out := runOpen(t, releaseOf(t, folders), nil, seeds, slices.Collect(maps.Values(folders)))
	if status != statusAccepted {
		t.Fatalf("status %d, %s%s", status, stdout, stderr)
	}
	views := openViewsOf(t, out)
	if views["format"] != openViewsFormat || views["version"] != int64(openViewsVersion) || views["module_format_version"] != "0.3" {
		t.Errorf("the root is %v, %v, %v", views["format"], views["version"], views["module_format_version"])
	}
	modules := views["modules"].([]any)
	if len(modules) != len(folders) {
		t.Fatalf("%d modules, not %d", len(modules), len(folders))
	}
	byID := map[string]string{}
	for _, folder := range folders {
		id, _ := inventoryOf(t, folder)
		byID[id] = folder
	}
	var ids []string
	for _, m := range modules {
		module := m.(map[string]any)
		id := module["id"].(string)
		ids = append(ids, id)
		pkg, err := readPackage(byID[id], true)
		if err != nil {
			t.Fatal(err)
		}
		if module["version"] != string(pkg.Inventory.Version) || module["module_format_version"] != "0.3" ||
			module["release"] != "v0.1.0" || module["package_digest"] != string(pkg.Digest) {
			t.Errorf("%s is named %v", id, module)
		}
		manifest, err := openview.ManifestView(pkg.View.Manifest)
		if err != nil || !reflect.DeepEqual(module["manifest"], manifest) {
			t.Errorf("%s: the manifest is %v, not its open view: %v", id, module["manifest"], err)
		}
		var want []string
		for _, taskID := range slices.Sorted(maps.Keys(pkg.View.Tasks)) {
			if pkg.View.Tasks[taskID].Template == nil {
				want = append(want, string(taskID))
			} else {
				want = append(want, string(taskID)+"#0", string(taskID)+"#20261027")
			}
		}
		var got []string
		for _, v := range module["tasks"].([]any) {
			task := v.(map[string]any)
			view := task["task"].(map[string]any)
			text, err := canonical.Text(view)
			if err != nil || task["open_view_digest"] != string(canonical.Digest(text)) {
				t.Errorf("%s %s: the digest %v is not of the open view written: %v", id, view["id"], task["open_view_digest"], err)
			}
			name := view["id"].(string)
			members := []string{"open_view_digest", "task"}
			if seed, ok := task["seed"]; ok {
				name += "#" + seed.(string)
				members = []string{"open_view_digest", "seed", "task"}
			} else if v, _, err := openview.TaskView(pkg.View.Tasks[formatv0.TaskID(name)]); err != nil || !reflect.DeepEqual(v, view) {
				t.Errorf("%s %s: the open view is %v, not the task's: %v", id, name, view, err)
			}
			if got := slices.Sorted(maps.Keys(task)); !slices.Equal(got, members) {
				t.Errorf("%s %s has the members %v, not %v", id, name, got, members)
			}
			got = append(got, name)
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s: the tasks are %v, not %v", id, got, want)
		}
	}
	if !slices.IsSorted(ids) {
		t.Errorf("the modules are not in the order of their IDs: %v", ids)
	}
	for _, id := range ids {
		if !strings.Contains(stdout, id+" 0.1.0: ") {
			t.Errorf("stdout does not name %s:\n%s", id, stdout)
		}
	}
}

// The instance of a template for a seed is the instance that an attempt of that seed gets: of
// missing-factor, for each seed of the test vector of templates, the open view of the vector's
// instance, which the package of with-templates holds the template of.
func TestOpenInstances(t *testing.T) {
	vector := vectorOf(t, "vectors/templates/missing-factor.yaml")
	want := map[string]any{}
	var seeds []string
	for _, a := range vector["attempts"].([]any) {
		attempt := a.(map[string]any)
		instance, err := schema.Decode[formatv0.Task](attempt["instance"])
		if err != nil {
			t.Fatal(err)
		}
		view, _, err := openview.TaskView(instance)
		if err != nil {
			t.Fatal(err)
		}
		seed := attempt["seed"].(string)
		want[seed] = view
		seeds = append(seeds, seed)
	}
	folders := vectorPackages(t, "with-templates")
	status, stdout, stderr, out := runOpen(t, releaseOf(t, folders), nil, seeds, []string{folders["with-templates"]})
	if status != statusAccepted {
		t.Fatalf("status %d, %s%s", status, stdout, stderr)
	}
	var order []string
	for _, v := range openViewsOf(t, out)["modules"].([]any)[0].(map[string]any)["tasks"].([]any) {
		task := v.(map[string]any)
		if task["task"].(map[string]any)["id"] != "missing-factor" {
			continue
		}
		seed := task["seed"].(string)
		order = append(order, seed)
		if !reflect.DeepEqual(task["task"], want[seed]) {
			t.Errorf("the seed %s gives %v, not the vector's instance %v", seed, task["task"], want[seed])
		}
	}
	if !slices.Equal(order, []string{"0", "20261027", "18446744073709551615"}) {
		t.Errorf("the seeds come as %v, not in the order of their numbers", order)
	}
}

// The files of the device part of each module go under modules/<ID>/<version>/, byte for byte as the
// module has them, and no other file does; the folder and its files are made as build makes a
// package's, and no temporary folder stays behind.
func TestOpenMedia(t *testing.T) {
	folders := vectorPackages(t, "with-media-and-notice")
	module := fromSpec(t, "vectors/packages/with-media-and-notice/module")
	status, stdout, stderr, out := runOpen(t, releaseOf(t, folders), nil, []string{"1"}, []string{folders["with-media-and-notice"]})
	if status != statusAccepted {
		t.Fatalf("status %d, %s%s", status, stdout, stderr)
	}
	id, version := inventoryOf(t, folders["with-media-and-notice"])
	root := filepath.Join(out, modulesFolder, filepath.FromSlash(id), version)
	var written []string
	err := filepath.WalkDir(out, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(out, file)
		written = append(written, filepath.ToSlash(rel))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	prefix := modulesFolder + "/" + id + "/" + version + "/"
	want := []string{openViewsFile, prefix + "LICENSE", prefix + "NOTICE", prefix + "media/pictures/dot.png"}
	slices.Sort(written)
	slices.Sort(want)
	if !slices.Equal(written, want) {
		t.Errorf("open wrote %v, not %v", written, want)
	}
	for _, name := range []string{"LICENSE", "NOTICE", "media/pictures/dot.png"} {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if source, err := os.ReadFile(filepath.Join(module, filepath.FromSlash(name))); err != nil || !bytes.Equal(got, source) {
			t.Errorf("%s is not the module's: %v", name, err)
		}
	}
	siblings, err := os.ReadDir(filepath.Dir(out))
	if err != nil || len(siblings) != 1 {
		t.Errorf("beside the folder: %v, %v", siblings, err)
	}
}

// Of each module, open takes the newest version that the newest index does not withdraw and whose
// format version module-kit reads, and names each newer one it skips; a module with no such version
// has no open views, and stdout says so.
func TestOpenWithdrawn(t *testing.T) {
	folders := vectorPackages(t, "smallest", "strings")
	smallest, smallestVersion := inventoryOf(t, folders["smallest"])
	stringsID, stringsVersion := inventoryOf(t, folders["strings"])
	other := "sha256:" + strings.Repeat("0", 64)
	release := map[string]map[string]entry{
		smallest: {
			smallestVersion: {folder: folders["smallest"]},
			"9.0.0":         {digest: other},
			"8.0.0":         {digest: other, format: "0.4"},
		},
		stringsID: {stringsVersion: {folder: folders["strings"]}},
	}
	latest := map[string]map[string]entry{
		smallest: {
			smallestVersion: {folder: folders["smallest"]},
			"9.0.0":         {digest: other, withdrawn: "wrong_key"},
			"8.0.0":         {digest: other, format: "0.4"},
		},
		stringsID:       {stringsVersion: {folder: folders["strings"], withdrawn: "answer_leak"}},
		"example/later": {"0.1.0": {digest: other}},
	}
	status, stdout, stderr, out := runOpen(t, indexText(t, release), indexText(t, latest), []string{"7"}, []string{folders["smallest"], folders["strings"]})
	if status != statusAccepted {
		t.Fatalf("status %d, %s%s", status, stdout, stderr)
	}
	for _, line := range []string{
		smallest + " 9.0.0: skipped, withdrawn (wrong_key)\n",
		smallest + " 8.0.0: skipped, of the format 0.4: module-kit reads the format 0.3\n",
		smallest + " " + smallestVersion + ": ",
		stringsID + " " + stringsVersion + ": skipped, withdrawn (answer_leak)\n",
		stringsID + ": no version to open\n",
		"the open views of 1 module are in ",
	} {
		if !strings.Contains(stdout, line) {
			t.Errorf("stdout does not say %q:\n%s", line, stdout)
		}
	}
	if strings.Contains(stdout, "example/later") {
		t.Errorf("stdout names a module of a later release:\n%s", stdout)
	}
	modules := openViewsOf(t, out)["modules"].([]any)
	if len(modules) != 1 || modules[0].(map[string]any)["id"] != smallest || modules[0].(map[string]any)["version"] != smallestVersion {
		t.Errorf("the modules are %v", modules)
	}
	if _, err := os.Stat(filepath.Join(out, modulesFolder, filepath.FromSlash(stringsID))); err == nil {
		t.Errorf("the files of %s, which is withdrawn, are written", stringsID)
	}
}

// What open cannot do exits 2, says why on stderr, and writes nothing.
func TestOpenFails(t *testing.T) {
	folders := vectorPackages(t, "smallest", "strings")
	release := releaseOf(t, folders)
	smallest, smallestVersion := inventoryOf(t, folders["smallest"])
	stringsID, _ := inventoryOf(t, folders["strings"])
	earlier := indexText(t, map[string]map[string]entry{smallest: {smallestVersion: {folder: folders["smallest"]}}})
	misnamed := indexText(t, map[string]map[string]entry{stringsID: {smallestVersion: {folder: folders["smallest"]}}})
	unheld := indexText(t, map[string]map[string]entry{smallest: {smallestVersion: {digest: "sha256:" + strings.Repeat("0", 64)}}})
	all := []string{folders["smallest"], folders["strings"]}
	noPackage := t.TempDir()
	full := t.TempDir()
	if err := os.WriteFile(filepath.Join(full, "file"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	indexFile := writeTemp(t, "module-index.json", release)
	openArgs := func(args ...string) []string { return append([]string{"open"}, args...) }
	for _, c := range []struct {
		name   string
		args   []string
		stderr string
	}{
		{"nothing", openArgs(), usage},
		{"no seed", openArgs("--index", indexFile, "--latest", indexFile, "--out", "o", all[0]), usage},
		{"no newest index", openArgs("--index", indexFile, "--seed", "1", "--out", "o", all[0]), usage},
		{"no package", openArgs("--index", indexFile, "--latest", indexFile, "--seed", "1", "--out", filepath.Join(t.TempDir(), "o")), "no package folder holds it"},
		{"no out", openArgs("--index", indexFile, "--latest", indexFile, "--seed", "1", all[0]), usage},
		{"a leading zero", openArgs("--seed", "01"), "no leading zero"},
		{"a sign", openArgs("--seed", "+1"), "no leading zero"},
		{"a negative seed", openArgs("--seed", "-1"), "no leading zero"},
		{"above 64 bits", openArgs("--seed", "18446744073709551616"), "no leading zero"},
		{"a seed twice", openArgs("--seed", "1", "--seed", "1"), "given twice"},
		{"no index file", openArgs("--index", filepath.Join(noPackage, "none.json"), "--latest", indexFile, "--seed", "1", "--out", filepath.Join(t.TempDir(), "o"), all[0]), "the index:"},
		{"no index", openArgs("--index", writeTemp(t, "x.json", []byte("[]")), "--latest", indexFile, "--seed", "1", "--out", filepath.Join(t.TempDir(), "o"), all[0]), "no module index"},
		{"no newest index file", openArgs("--index", indexFile, "--latest", filepath.Join(noPackage, "none.json"), "--seed", "1", "--out", filepath.Join(t.TempDir(), "o"), all[0]), "the newest index:"},
		{"an earlier index as the newest", openArgs("--index", indexFile, "--latest", writeTemp(t, "e.json", earlier), "--seed", "1", "--out", filepath.Join(t.TempDir(), "o"), all[0], all[1]), "not an index of a release at or after"},
		{"a folder that is no package", openArgs("--index", indexFile, "--latest", indexFile, "--seed", "1", "--out", filepath.Join(t.TempDir(), "o"), noPackage), "holds no package"},
		{"a package no folder holds", openArgs("--index", writeTemp(t, "u.json", unheld), "--latest", writeTemp(t, "u2.json", unheld), "--seed", "1", "--out", filepath.Join(t.TempDir(), "o"), all[0], all[1]), "no package folder holds it"},
		{"a package of another module", openArgs("--index", writeTemp(t, "m.json", misnamed), "--latest", writeTemp(t, "m2.json", misnamed), "--seed", "1", "--out", filepath.Join(t.TempDir(), "o"), all[0], all[1]), "and it is " + smallest},
		{"an out folder that is not empty", openArgs("--index", indexFile, "--latest", indexFile, "--seed", "1", "--out", full, all[0], all[1]), "is not empty"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if status := run(context.Background(), c.args, &stdout, &stderr); status != statusFailed {
				t.Errorf("status %d, %s%s", status, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), c.stderr) {
				t.Errorf("stderr is %q, without %q", stderr.String(), c.stderr)
			}
		})
	}
	if entries, err := os.ReadDir(full); err != nil || len(entries) != 1 {
		t.Errorf("the folder that was not empty holds %v, %v", entries, err)
	}
}

// A package whose chain of digests breaks past its inventory, and an out folder that cannot be made,
// exit 2 too.
func TestOpenFailsLater(t *testing.T) {
	folders := vectorPackages(t, "smallest")
	release := releaseOf(t, folders)
	broken := filepath.Join(t.TempDir(), "broken")
	if err := os.CopyFS(broken, os.DirFS(folders["smallest"])); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, modulepackage.ServerFile), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name, folder, out, stderr string
	}{
		{"a broken chain", broken, filepath.Join(t.TempDir(), "o"), "server.json"},
		{"an out folder in no folder", folders["smallest"], filepath.Join(t.TempDir(), "none", "o"), "module-kit open: "},
	} {
		t.Run(c.name, func(t *testing.T) {
			indexFile := writeTemp(t, "module-index.json", release)
			var stdout, stderr bytes.Buffer
			args := []string{"open", "--index", indexFile, "--latest", indexFile, "--seed", "1", "--out", c.out, c.folder}
			if status := run(context.Background(), args, &stdout, &stderr); status != statusFailed || !strings.Contains(stderr.String(), c.stderr) {
				t.Errorf("status %d, %s%s; want 2 and %q", status, stdout.String(), stderr.String(), c.stderr)
			}
		})
	}
}

// moduleViews fails on what no package read by its chain of digests holds: a manifest or a task
// that is not valid against its schema, a template without instances, and an instance without the
// numbers of its placeholders.
func TestModuleViewsFails(t *testing.T) {
	folders := vectorPackages(t, "with-templates")
	pkg, err := readPackage(folders["with-templates"], true)
	if err != nil {
		t.Fatal(err)
	}
	seeds := seedList{1}
	if _, _, err := moduleViews(pkg, formatv0.Entry{}, seeds); err != nil {
		t.Fatalf("the package's own views: %v", err)
	}
	const id = formatv0.TaskID("missing-factor")
	for name, change := range map[string]func(p *modulepackage.Package){
		"an invalid manifest": func(p *modulepackage.Package) { p.View.Manifest = formatv0.Manifest{} },
		"an invalid task": func(p *modulepackage.Package) {
			p.View.Tasks = map[formatv0.TaskID]formatv0.Task{"no-items": {ID: "no-items"}}
		},
		"a template of no instance": func(p *modulepackage.Package) { p.View.Instances = map[formatv0.TaskID][]formatv0.Numbers{} },
		"an instance of no numbers": func(p *modulepackage.Package) {
			p.View.Instances = map[formatv0.TaskID][]formatv0.Numbers{id: {{}}}
		},
	} {
		changed := *pkg
		change(&changed)
		if _, _, err := moduleViews(&changed, formatv0.Entry{}, seeds); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

// Versions not in the form X.Y.Z order by their text, and the seeds of a command line are written as
// it gives them.
func TestOpenHelpers(t *testing.T) {
	if compareVersions("b", "a") <= 0 || compareVersions("0.10.0", "0.9.0") <= 0 || compareVersions("1.0.0", "x") >= 0 {
		t.Error("compareVersions orders versions otherwise")
	}
	var seeds seedList
	for _, s := range []string{"7", "18446744073709551615"} {
		if err := seeds.Set(s); err != nil {
			t.Fatal(err)
		}
	}
	if got := seeds.String(); got != "7,18446744073709551615" {
		t.Errorf("String = %q", got)
	}
	var skipped strings.Builder
	writeSkipped(&skipped, "example/x", map[formatv0.ModuleVersion]formatv0.Entry{"x": {ModuleFormatVersion: "0.3"}}, "", false)
	if got := skipped.String(); got != "example/x x: skipped, not a version X.Y.Z\nexample/x: no version to open\n" {
		t.Errorf("writeSkipped wrote %q", got)
	}
}

// A package whose inventory names another format version than the entry that names it is no
// package of that entry: exit 2, whatever its ID and version.
func TestOpenFormatOfThePackage(t *testing.T) {
	folder := fromSpec(t, "vectors/versions/earlier-format-older-version/published/0.9.0")
	id, version := inventoryOf(t, folder)
	release := indexText(t, map[string]map[string]entry{id: {version: {folder: folder}}})
	status, stdout, stderr, _ := runOpen(t, release, nil, []string{"1"}, []string{folder})
	if status != statusFailed || !strings.Contains(stderr, "0.2") {
		t.Errorf("status %d, %s%s; want 2 and the format 0.2 named", status, stdout, stderr)
	}
}

// A release whose every version the newest index withdraws has no module to open, and needs no
// package folder: open writes open-views.json with no module.
func TestOpenNoModule(t *testing.T) {
	withdrawn := indexText(t, map[string]map[string]entry{"example/gone": {"0.1.0": {digest: "sha256:" + strings.Repeat("0", 64), withdrawn: "unsuitable_content"}}})
	status, stdout, stderr, out := runOpen(t, withdrawn, nil, []string{"1"}, nil)
	if status != statusAccepted {
		t.Fatalf("status %d, %s%s", status, stdout, stderr)
	}
	if modules := openViewsOf(t, out)["modules"].([]any); len(modules) != 0 {
		t.Errorf("the modules are %v", modules)
	}
	if !strings.Contains(stdout, "the open views of 0 modules are in ") {
		t.Errorf("stdout: %s", stdout)
	}
}
