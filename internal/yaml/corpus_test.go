package yaml

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/altnovatrail/spec"
)

// Every YAML file of the format but the vectors of the rules yaml.* follows section 4 — the
// examples, invalid ones too, and the other vectors (format/README.md, sections 8 and 9) — so the
// reader reads each with no breach: 2302 files at v0.9.2. It holds each move to a newer go-yaml
// too, since the format's files use most of what YAML 1.2.2 has (R21).
func TestReadsTheFormat(t *testing.T) {
	read := 0
	err := fs.WalkDir(spec.Format, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(name, ".yaml") || strings.HasPrefix(name, "vectors/rules/yaml.") {
			return err
		}
		data, err := fs.ReadFile(spec.Format, name)
		if err != nil {
			return err
		}
		if _, err := Read(data); err != nil {
			t.Errorf("format/%s:\n%v", name, err)
		}
		read++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if read == 0 {
		t.Fatal("spec.Format holds no YAML file")
	}
	t.Logf("%d files of the format read", read)
}
