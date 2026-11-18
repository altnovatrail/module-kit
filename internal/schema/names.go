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
// (santhosh-tekuri/jsonschema#276). An error the vocabulary adds gets a copy of the place. collect
// leaves out the validator's own error of propertyNames and keeps this one: they fail on the same
// names, so the vocabulary does not change whether a value passes. A release with the fix is a
// reason to remove it (R23).
var propertyNames = &jsonschema.Vocabulary{
	URL: "https://altnovatrail.github.io/module-kit/vocabulary/property-names",
	Compile: func(ctx *jsonschema.CompilerContext, obj map[string]any) (jsonschema.SchemaExt, error) {
		if _, ok := obj["propertyNames"]; !ok {
			return nil, nil
		}
		return &namesCheck{ctx.Enqueue([]string{"propertyNames"})}, nil
	},
}

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

type badName struct {
	name string
}

func (*badName) KeywordPath() []string {
	return []string{"propertyNames"}
}

func (b *badName) LocalizedString(*message.Printer) string {
	return fmt.Sprintf("the name %q is none that the schema allows here", b.name)
}
