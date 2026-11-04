// Package yaml reads a file of a module, YAML 1.2.2 under its Core schema, into a JSON value, and
// refuses a file that breaks a rule yaml.*, naming the rule and its place, or that is not YAML
// 1.2.2 at all (format/README.md, section 4; R45 spec). The check of a module reads each file
// through it, and the test of vectors reads the vectors of the rules yaml.* (R21).
//
// go.yaml.in/yaml/v4 parses the file into its tree of nodes. It keeps some of YAML 1.1 and lets
// through some of what YAML 1.2.2 refuses, so passes of this package go around it, each in a file
// of its own: text.go before the parse, errors.go on its errors, walk.go over its tree for the
// rules, and layout.go over its tree for what YAML 1.2.2 refuses and v4 reads all the same. Each
// pass names what of v4 it closes, and a test holds it (R21).
package yaml
