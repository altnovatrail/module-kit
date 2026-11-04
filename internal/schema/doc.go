// Package schema checks a value against the JSON Schema 2020-12 schemas of the format, as
// spec.Format holds them, reporting every keyword that fails with its place, and collects the marks
// x-open where their subschema applies and passes (format/README.md, sections 5, 6 and 8).
//
// It holds no code yet: the epic of code brings it (R16).
package schema
