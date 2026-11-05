package openview

import (
	"context"
	"path"
	"strings"
	"testing"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/template"
)

// The digest of the open view of each task of the vectors of the open view, and of an instance of
// heads-and-legs, as spec's second implementation computes them: openViewOf and canonicalText of
// tools/format-check at v0.9.2 over the same files, run on 2026-11-05. The format has no vector of
// the digest; these pin the canonical text of an open view, byte for byte, between the two.
var openViewDigests = map[string]formatv0.Digest{
	"answer-anywhere":     "sha256:1bc73a7266f205d220a14033b5a210a911b9a4d41861a8f1f7c06eb581cc47ac",
	"choice-for-adults":   "sha256:a6b4bfa50d317a0d43cec3df43da6a64536ae6643021be5739f287643827b222",
	"essay-rubric-closed": "sha256:72db0fe7fc1a978e7af1de52a9067e4914263940de3ac53b7160512127b59531",
	"hints-closed":        "sha256:59a6253b98779dcd2affe098dc06fbe3b57b74722b2f1b9855316b929a569048",
	"template-closed":     "sha256:824216b1d588a31d68cdd4616f0158e99ba05a69272f26261bec5fb172189a61",
	"workbook-lines":      "sha256:d2311b7b53ab0f6b10d384a6afe63b715b39d24b44eed400ddd20ff6e1a46d80",
	"zones-and-picture":   "sha256:d4784fb2a59d1577bd220c9980e4c7bc2f77f4de7f67b9be277370af92384931",
}

// The digest of the open view of a task is the one spec's second implementation gives, of a task
// and of a template as written alike.
func TestTaskDigest(t *testing.T) {
	for c, want := range openViewDigests {
		var task formatv0.Task
		readPart(t, path.Join("vectors/open-view/task", c, "part.yaml"), &task)
		got, err := TaskDigest(task)
		if err != nil || got != want {
			t.Errorf("%s: TaskDigest = %s, %v; want %s", c, got, err, want)
		}
	}
}

// The digest of an instance is of its open view with its numbers in place: of the instance that the
// seed 0 picks of heads-and-legs, chickens = 1 and rabbits = 8, it is the one spec's second
// implementation gives of the instance the vector holds, and not the digest of the template.
func TestInstanceDigest(t *testing.T) {
	var vector formatv0.TemplateVector
	readPart(t, "vectors/templates/heads-and-legs.yaml", &vector)
	numbers, breaches, err := template.Instances(context.Background(), vector.Task)
	if err != nil || len(breaches) > 0 {
		t.Fatal(err, breaches)
	}
	picked, err := template.AttemptInstance(0, vector.Task.ID, len(numbers))
	if err != nil {
		t.Fatal(err)
	}
	instance, err := template.Instance(vector.Task, numbers[picked])
	if err != nil {
		t.Fatal(err)
	}
	const want = "sha256:98b4fd15d4a17e04e35e9a733c837e8f26081c3db8622ad8e6925321a47c5d17"
	if got, err := TaskDigest(instance); err != nil || got != want {
		t.Errorf("TaskDigest of the instance = %s, %v; want %s", got, err, want)
	}
	if got, _ := TaskDigest(vector.Task); got == want {
		t.Error("the template and its instance have one digest")
	}
}

// A task that is not valid against its schema has no open view, and no digest.
func TestTaskDigestOfAnInvalidTask(t *testing.T) {
	if got, err := TaskDigest(formatv0.Task{ID: "no-items"}); err == nil || !strings.Contains(err.Error(), "not valid") || got != "" {
		t.Errorf("TaskDigest = %q, %v; want an error", got, err)
	}
}
