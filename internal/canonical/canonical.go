package canonical

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/altnovatrail/spec/types/formatv0"

	"github.com/altnovatrail/module-kit/internal/schema"
)

// MaxInteger is the largest integer that the canonical text writes, 2^53 − 1, as section 4 of
// format/README.md bounds an integer; the smallest is −MaxInteger.
const MaxInteger = 1<<53 - 1

// An Error is a value that the canonical text cannot hold, at its place in the value as a JSON
// Pointer in URI fragment form, such as #/tasks/sum/items/0. Of a module that passed its check, it
// is a fault of module-kit, not of the author (R09).
type Error struct {
	Pointer string
	Message string
}

func (e *Error) Error() string {
	return "canonical: " + e.Pointer + ": " + e.Message
}

// Text gives the canonical text of a JSON value, as format/README.md, section 25, writes it — RFC
// 8785, which for the data of the format comes to this: the members of an object in the order of
// the UTF-16 code units of their names, no whitespace between tokens, an integer in decimal digits,
// in a string only ", \, the five short escapes and \u00XX for the rest of U+0000 to U+001F, and
// every other character as it is; UTF-8, with no byte order mark and no newline at the end (R09).
//
// value is a tree of map[string]any, []any, string, int64, bool and nil, as the reader of YAML
// gives it once an integer that it gave as a double is an int64. Text stops with an *Error, rather
// than write it otherwise, at the first value it cannot hold: a float64, an integer outside
// ±(2^53 − 1), a string or a name that is not UTF-8, and a value of any other Go type.
func Text(value any) ([]byte, error) {
	var w writer
	if err := w.value(value, nil); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// Digest gives the digest of bytes: sha256: and their SHA-256 in 64 lowercase hexadecimal digits
// (format/README.md, section 25).
func Digest(data []byte) formatv0.Digest {
	sum := sha256.Sum256(data)
	return formatv0.Digest("sha256:" + hex.EncodeToString(sum[:]))
}

// Descriptor names bytes as a package does: their digest and their size.
func Descriptor(data []byte) formatv0.Descriptor {
	return formatv0.Descriptor{Digest: Digest(data), Size: int64(len(data))}
}

type writer struct {
	bytes.Buffer
}

func (w *writer) value(value any, at []string) error {
	switch v := value.(type) {
	case nil:
		w.WriteString("null")
	case bool:
		w.WriteString(strconv.FormatBool(v))
	case int64:
		if v < -MaxInteger || v > MaxInteger {
			return refuse(at, "%d is past ±(2^53 − 1), the integers of the format", v)
		}
		w.WriteString(strconv.FormatInt(v, 10))
	case float64:
		if v != math.Trunc(v) || math.IsInf(v, 0) {
			return refuse(at, "%v is a number with a fraction, and a number of the format is an integer", v)
		}
		return refuse(at, "%v is an integer given as a double: it comes as an int64", v)
	case string:
		return w.string(v, at)
	case []any:
		w.WriteByte('[')
		for i, element := range v {
			if i > 0 {
				w.WriteByte(',')
			}
			if err := w.value(element, append(at, strconv.Itoa(i))); err != nil {
				return err
			}
		}
		w.WriteByte(']')
	case map[string]any:
		names := slices.SortedFunc(maps.Keys(v), byCodeUnits)
		w.WriteByte('{')
		for i, name := range names {
			if i > 0 {
				w.WriteByte(',')
			}
			if err := w.string(name, append(at, name)); err != nil {
				return err
			}
			w.WriteByte(':')
			if err := w.value(v[name], append(at, name)); err != nil {
				return err
			}
		}
		w.WriteByte('}')
	default:
		return refuse(at, "a value of the type %T is no JSON value", value)
	}
	return nil
}

// The short escapes of section 25; every other character from U+0000 to U+001F is written \u00XX.
var short = map[rune]string{'"': `\"`, '\\': `\\`, '\b': `\b`, '\t': `\t`, '\n': `\n`, '\f': `\f`, '\r': `\r`}

const hexDigits = "0123456789abcdef"

func (w *writer) string(s string, at []string) error {
	if !utf8.ValidString(s) {
		return refuse(at, "%q is not UTF-8", s)
	}
	w.WriteByte('"')
	for _, r := range s {
		switch escape, ok := short[r]; {
		case ok:
			w.WriteString(escape)
		case r < 0x20:
			w.WriteString(`\u00`)
			w.WriteByte(hexDigits[r>>4])
			w.WriteByte(hexDigits[r&0xf])
		default:
			w.WriteRune(r)
		}
	}
	w.WriteByte('"')
	return nil
}

// byCodeUnits orders two names by their UTF-16 code units, as RFC 8785, section 3.2.3, orders the
// members of an object: a character outside the Basic Multilingual Plane, a surrogate pair, comes
// before U+E000 to U+FFFF, though its code point is above them.
func byCodeUnits(a, b string) int {
	return slices.Compare(utf16.Encode([]rune(a)), utf16.Encode([]rune(b)))
}

func refuse(at []string, format string, args ...any) error {
	return &Error{Pointer: schema.Pointer(at), Message: fmt.Sprintf(format, args...)}
}
