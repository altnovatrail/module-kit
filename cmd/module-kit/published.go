package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/moduleindex"
	"github.com/altnovatrail/module-kit/modulepackage"
	"github.com/altnovatrail/module-kit/moduleversion"
)

// publishedOf reads the index file, by the places every version of the format keeps
// (moduleindex.Read), and gives the published versions of the module id that it names, withdrawn
// ones too: each the package among the folders whose inventory.json has the digest of its entry. A
// version whose package no folder holds cannot be compared, and folders of other packages are
// skipped. The comparison reads each package by the chain of digests from that digest.
func publishedOf(indexFile string, id formatv0.ModuleID, folders []string) ([]moduleversion.Published, error) {
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("the index: %w", err)
	}
	index, err := moduleindex.Read(data)
	if err != nil {
		return nil, fmt.Errorf("the index, %s: %w", indexFile, err)
	}
	byDigest := map[formatv0.Digest]string{}
	for _, dir := range folders {
		inventory, err := os.ReadFile(filepath.Join(dir, modulepackage.InventoryFile))
		if err != nil {
			return nil, fmt.Errorf("%s holds no package: %w", dir, err)
		}
		byDigest[canonical.Digest(inventory)] = dir
	}
	entries := index.Modules[id]
	var versions []moduleversion.Published
	for _, version := range slices.Sorted(maps.Keys(entries)) {
		entry := entries[version]
		dir, ok := byDigest[entry.PackageDigest]
		if !ok {
			return nil, fmt.Errorf("the index names %s %s, with the package %s, and no --published folder holds it: every published version is compared",
				id, version, entry.PackageDigest)
		}
		versions = append(versions, moduleversion.Published{Digest: entry.PackageDigest, Files: os.DirFS(dir)})
	}
	return versions, nil
}
