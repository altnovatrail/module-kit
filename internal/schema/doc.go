// Package schema checks a value against the JSON Schema 2020-12 schemas of the format version the
// library reads, reporting every keyword that fails with its place, and builds the open view of a
// value from the marks x-open in the same schemas. The schemas are a copy of schemas/ of
// spec.Format at the tag go.mod requires, built into the package, so that the code holds the
// schemas of the version it runs; a test holds the copy to spec.Format byte for byte.
//
// The validator collects no annotations, so Open walks the schemas itself and asks the validator
// only whether a subschema passes. Decode reads a JSON value into a type of formatv0, and
// ValueOf gives the JSON value of one back.
package schema
