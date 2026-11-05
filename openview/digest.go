package openview

import (
	"fmt"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/canonical"
	"github.com/altnovatrail/module-kit/internal/schema"
)

// TaskDigest gives the digest of the open view of a task: sha256: and the SHA-256 of its canonical
// text (format/README.md, section 25; R93 spec). Of an instance of a template, the task is what
// template.Instance gives, with its numbers in place, so that the digest is of the open view the
// learner's app received. tutor's eval runner finds the task of a sample by it, and the CLI's open
// writes it beside each open view (contracts/README.md, section 11).
//
// The open view is the one Task gives, taken as the JSON value that the walk of the schemas gives,
// before any type holds it, so that a field is in the text exactly when the open view has it. It
// fails, as Task does, when the task is not valid against its schema.
func TaskDigest(part formatv0.Task) (formatv0.Digest, error) {
	const file = "task.schema.json"
	value := schema.ValueOf(part)
	if errs := schema.Validate(file, value); len(errs) > 0 {
		return "", fmt.Errorf("openview: the task is not valid against %s: %s", file, errs[0])
	}
	text, err := canonical.Text(schema.Open(file, value))
	if err != nil {
		panic(fmt.Sprintf("openview: the open view of a valid task: %v", err))
	}
	return canonical.Digest(text), nil
}
