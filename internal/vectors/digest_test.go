package vectors

import (
	"io/fs"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/altnovatrail/spec"
)

// A digest that is not the open view's fails, of a task's case of the open view and of an attempt of
// templates; a digest.yaml with no digest fails too, and so does one with a field besides it.
func TestDigestsRefuse(t *testing.T) {
	const task = "vectors/open-view/task/answer-anywhere"
	const wrong = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	files := fstest.MapFS{}
	for _, name := range []string{"digest.yaml", "open-view.yaml", "part.yaml"} {
		data, err := fs.ReadFile(spec.Format, path.Join(task, name))
		if err != nil {
			t.Fatal(err)
		}
		files[path.Join(task, name)] = &fstest.MapFile{Data: data}
	}
	if err := openView(vector{fsys: files, path: task}); err != nil {
		t.Fatalf("the case of spec: %v", err)
	}
	for _, c := range []struct {
		name, digest, says string
	}{
		{"another digest", "open_view_digest: " + wrong + "\n", "digest.yaml holds " + wrong},
		{"no digest", "{}\n", "digest.yaml holds \n"},
		{"a field besides", string(files[path.Join(task, "digest.yaml")].Data) + "note: x\n", "unknown field"},
	} {
		changed := fstest.MapFS{}
		for name, f := range files {
			changed[name] = f
		}
		changed[path.Join(task, "digest.yaml")] = &fstest.MapFile{Data: []byte(c.digest)}
		if err := openView(vector{fsys: changed, path: task}); err == nil || !strings.Contains(err.Error()+"\n", c.says) {
			t.Errorf("%s: %v, want %q", c.name, err, c.says)
		}
	}

	const templateCase = "vectors/templates/one-instance.yaml"
	data, err := fs.ReadFile(spec.Format, templateCase)
	if err != nil {
		t.Fatal(err)
	}
	if err := templates(vector{fsys: fstest.MapFS{"v.yaml": {Data: data}}, path: "v.yaml"}); err != nil {
		t.Fatalf("the case of spec: %v", err)
	}
	digest := "sha256:c928cb2cec7ebd00c71f5f0f1be7a25b0d9393fdb35ca039454d1a10053cb50f"
	if !strings.Contains(string(data), digest) {
		t.Fatalf("%s gives no digest %s", templateCase, digest)
	}
	changed := strings.Replace(string(data), digest, wrong, 1)
	err = templates(vector{fsys: fstest.MapFS{"v.yaml": {Data: []byte(changed)}}, path: "v.yaml"})
	if want := "attempt 1: the digest of the instance's open view is " + digest; err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("another digest of an attempt: %v, want %q", err, want)
	}
}
