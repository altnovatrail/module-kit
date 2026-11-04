package schema

import (
	"fmt"
	"maps"
	"slices"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/message"
)

// propertyNames is the vocabulary that checks propertyNames once more, to give its error the right
// place. The validator, v6.0.3, gives the error of propertyNames the live slice of its place in the
// value rather than a copy, and the validation of a later sibling writes over it, so that the error
// names another place, which also depends on the order in which Go goes through a map
// (santhosh-tekuri/jsonschema#276, open on 2026-11-04). The vocabulary runs at the same place with
// the same subschema, and an error it adds gets a copy of the place. collect leaves out the
// validator's own error of propertyNames and keeps this one: they fail on the same names, so the
// value passes or fails as before.
var propertyNames = &jsonschema.Vocabulary{
	URL: "https://altnovatrail.github.io/module-kit/vocabulary/property-names",
	Compile: func(ctx *jsonschema.CompilerContext, obj map[string]any) (jsonschema.SchemaExt, error) {
		if _, ok := obj["propertyNames"]; !ok {
			return nil, nil
		}
		return &namesCheck{ctx.Enqueue([]string{"propertyNames"})}, nil
	},
}

// namesCheck is propertyNames at one place: the subschema each name of an object is checked against.
type namesCheck struct {
	names *jsonschema.Schema
}

func (n *namesCheck) Validate(ctx *jsonschema.ValidatorContext, v any) {
	obj, ok := v.(map[string]any)
	if !ok {
		return
	}
	for _, name := range slices.Sorted(maps.Keys(obj)) {
		if n.names.Validate(name) != nil {
			ctx.AddError(&badName{name})
		}
	}
}

// badName is the error of propertyNames that the vocabulary adds: a name of the object at its place
// that the subschema refuses.
type badName struct {
	name string
}

func (*badName) KeywordPath() []string {
	return []string{"propertyNames"}
}

func (b *badName) LocalizedString(*message.Printer) string {
	return fmt.Sprintf("the name %q is none that the schema allows here", b.name)
}
