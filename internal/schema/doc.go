// Package schema checks a value against the JSON Schema 2020-12 schemas of the format version the
// library reads, reporting every keyword that fails with its place (format/README.md, sections 5
// and 8). The schemas are a copy of schemas/ of spec.Format at the tag go.mod requires, built into
// the package, so that the code holds the schemas of the version it runs, as module.FormatVersion
// holds the version; a test holds the copy to spec.Format byte for byte (R23).
//
// The validator is github.com/santhosh-tekuri/jsonschema/v6. It collects no annotations: the open
// view walks the schemas by section 6 itself and asks this package only whether a subschema passes
// (R23). The check of a module and the test of vectors call it.
package schema
