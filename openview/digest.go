package openview

import (
	"fmt"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
)

// TaskDigest gives the digest of the open view of a task: sha256: and the SHA-256 of its canonical
// text (format/README.md, section 25; R93 spec). Of an instance of a template, the task is what
// template.Instance gives, with its numbers in place, so that the digest is of the open view the
// learner's app received. tutor's eval runner finds the task of a sample by it, and the CLI's open
// writes it beside each open view (contracts/README.md, section 11).
//
// The open view is the one TaskView gives, taken as the JSON value that the walk of the schemas
// gives, before any type holds it, so that a field is in the text exactly when the open view has it.
// It fails, as Task does, when the task is not valid against its schema.
func TaskDigest(part formatv0.Task) (formatv0.Digest, error) {
	_, digest, err := TaskView(part)
	return digest, err
}

// TaskView gives the open view of a task as a JSON value — map[string]any, []any, string, int64 or
// bool, as the reader of YAML gives one — and the digest of its canonical text, TaskDigest's. The
// value is the open view before any type holds it, so that its canonical text is the one the digest
// is of: the CLI's open writes it, and a reader that digests the text it reads finds the digest
// beside it (R43). It fails, as Task does, when the task is not valid against its schema.
func TaskView(part formatv0.Task) (view any, digest formatv0.Digest, err error) {
	view, err = viewOf("task", part)
	if err != nil {
		return nil, "", err
	}
	text, err := canonical.Text(view)
	if err != nil {
		panic(fmt.Sprintf("openview: the open view of a valid task: %v", err))
	}
	return view, canonical.Digest(text), nil
}

// ManifestView gives the open view of a manifest as a JSON value, as TaskView does of a task.
func ManifestView(part formatv0.Manifest) (any, error) {
	return viewOf("manifest", part)
}
