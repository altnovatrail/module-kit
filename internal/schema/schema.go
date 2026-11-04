package schema

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// files is schemas/ of spec.Format at the tag go.mod requires, byte for byte (TestSchemasAreSpecs).
//
//go:embed schemas/*.schema.json
var files embed.FS

// An Error is one keyword that fails, at the place where the schema holding it applies, as an
// example of the format names it: the object for required, unevaluatedProperties and the other
// keywords of an object, the array for those of an array, and the value for type, pattern and the
// other keywords of a value (format/README.md, section 8).
type Error struct {
	Keyword string
	// The place in the value, token by token from its root: the name of a member, or the index of
	// an element in decimal digits. Pointer writes it as a JSON Pointer.
	Location []string
	Message  string
}

// Pointer gives the place of the error as a JSON Pointer in URI fragment form, as the first line
// of an invalid example writes it: #/answer_slots/0.
func (e Error) Pointer() string {
	return Pointer(e.Location)
}

func (e Error) String() string {
	return e.Keyword + " at " + e.Pointer() + ": " + e.Message
}

// Keywords are the keywords that a schema of the format uses, and so the only ones an Error may
// name, as section 5 of format/README.md lists them; a test holds the list to the text.
var Keywords = []string{
	"$schema", "$id", "$ref", "$defs", "$comment",
	"allOf", "anyOf", "oneOf", "not", "if", "then", "else", "prefixItems", "items", "contains", "properties",
	"patternProperties", "additionalProperties", "propertyNames", "unevaluatedProperties",
	"type", "enum", "const", "multipleOf", "maximum", "exclusiveMaximum", "minimum", "exclusiveMinimum", "maxLength",
	"minLength", "pattern", "maxItems", "minItems", "uniqueItems", "maxProperties", "minProperties", "required",
	"dependentRequired",
	"title", "description", "default", "deprecated", "readOnly", "writeOnly", "examples", "format",
	"x-open",
}

// Validate checks value against the schema at location: the name of a schema's file, such as
// task.schema.json, or that name with a JSON Pointer to a subschema in it as its fragment, such as
// task.schema.json#/$defs/check. value is a JSON value as the reader of YAML gives it:
// map[string]any, []any, string, int64, float64, bool or nil. It gives every keyword that fails,
// and none when the value is valid. Each schema is compiled once, and Validate may be called from
// many goroutines at once.
//
// A location that names no schema of the format, and a schema that does not compile, are faults in
// module-kit rather than in the value, and Validate panics on them, as regexp.MustCompile does.
func Validate(location string, value any) []Error {
	return format().validate(location, value)
}

// Files names the schemas of the format, as schemas/ of spec.Format names them.
func Files() []string {
	return names(files)
}

func names(fsys fs.FS) []string {
	// The pattern is well formed, and fs.Glob fails on nothing else.
	found, _ := fs.Glob(fsys, "schemas/*.schema.json")
	for i, name := range found {
		found[i] = path.Base(name)
	}
	return found
}

// format is the set of the schemas of the format, built once.
var format = sync.OnceValue(func() *set { return newSet(files) })

// A set is every schema of schemas/ in a file system, added to one compiler by its $id, and what
// is compiled of them so far, by location. A schema once compiled is read without a lock, and the
// compiler, which is not safe from many goroutines, is held by mu. The open view walks the schemas
// as they are written, docs, and reads the keys of patternProperties as patterns, compiled once.
type set struct {
	compiler *jsonschema.Compiler
	ids      map[string]string // the $id of each schema, by the name of its file
	docs     map[string]any    // each schema as JSON, by the name of its file
	mu       sync.Mutex
	schemas  sync.Map // *jsonschema.Schema by location
	patterns sync.Map // *regexp.Regexp by the key of patternProperties
}

// newSet adds every schema of schemas/ in fsys to a compiler. A reference between schemas resolves
// against the $id of the schema it stands in (format/README.md, section 5), so each schema is added
// by its own $id, which ends in the name of its file, and the compiler loads nothing else: its
// loader reads only file: URLs, and no schema of the format has one.
func newSet(fsys fs.FS) *set {
	s := &set{compiler: jsonschema.NewCompiler(), ids: map[string]string{}, docs: map[string]any{}}
	s.compiler.RegisterVocabulary(propertyNames)
	s.compiler.AssertVocabs()
	for _, name := range names(fsys) {
		data, err := fs.ReadFile(fsys, "schemas/"+name)
		if err != nil {
			panic(err)
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			panic(fmt.Sprintf("schema: %s: %v", name, err))
		}
		root, _ := doc.(map[string]any)
		id, _ := root["$id"].(string)
		if path.Base(id) != name {
			panic(fmt.Sprintf("schema: %s has the $id %q, which does not end in its name", name, id))
		}
		if err := s.compiler.AddResource(id, doc); err != nil {
			panic(fmt.Sprintf("schema: %s: %v", name, err))
		}
		s.ids[name] = id
		s.docs[name] = doc
	}
	return s
}

func (s *set) validate(location string, value any) []Error {
	err := s.schema(location).Validate(value)
	if err == nil {
		return nil
	}
	verr, ok := err.(*jsonschema.ValidationError)
	if !ok {
		panic(fmt.Sprintf("schema: %s: %v", location, err))
	}
	return collect(verr, nil)
}

// schema compiles the schema at location, the first time it is asked for.
func (s *set) schema(location string) *jsonschema.Schema {
	name, fragment, hasFragment := strings.Cut(location, "#")
	id, ok := s.ids[name]
	if !ok {
		panic(fmt.Sprintf("schema: %s is no schema of the set", location))
	}
	if hasFragment {
		id += "#" + fragment
	}
	if sch, ok := s.schemas.Load(location); ok {
		return sch.(*jsonschema.Schema)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if sch, ok := s.schemas.Load(location); ok {
		return sch.(*jsonschema.Schema)
	}
	sch, err := s.compiler.Compile(id)
	if err != nil {
		panic(fmt.Sprintf("schema: %s: %v", location, err))
	}
	s.schemas.Store(location, sch)
	return sch
}
