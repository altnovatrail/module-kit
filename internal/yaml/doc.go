// Package yaml reads a file of a module, YAML 1.2.2 under its Core schema, into a JSON value, and
// refuses a file that breaks a rule yaml.*, naming the rule and its place, or that is not YAML
// 1.2.2 at all.
//
// go.yaml.in/yaml/v4 parses the file into its tree of nodes. It keeps some of YAML 1.1 and lets
// through some of what YAML 1.2.2 refuses, so passes of this package go around it: text.go before
// the parse, errors.go on its errors, walk.go over its tree for the rules, and layout.go over its
// tree for what YAML 1.2.2 refuses and v4 reads all the same. A test holds each of them.
package yaml
