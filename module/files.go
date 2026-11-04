package module

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
	"unicode/utf8"
)

// The rules of the files of a module that no schema checks (format/README.md, section 10): its
// media files, the text of its license and its notices.

// A signature of a media format: its name, and whether the first bytes of a file are of it.
type signature struct {
	format  string
	matches func(head []byte) bool
}

// signatures holds the signature of each media format by its extension, as WHATWG MIME Sniffing,
// 6.1 and 6.2, gives them (format/README.md, section 10). headLength bytes of a file are enough for
// each.
var signatures = map[string]signature{
	"png": {"PNG", func(head []byte) bool { return bytes.HasPrefix(head, []byte("\x89PNG\r\n\x1a\n")) }},
	"jpg": {"JPEG", func(head []byte) bool { return bytes.HasPrefix(head, []byte{0xff, 0xd8, 0xff}) }},
	"webp": {"WebP", func(head []byte) bool {
		return len(head) >= 14 && string(head[:4]) == "RIFF" && string(head[8:14]) == "WEBPVP"
	}},
	"mp3": {"MP3", func(head []byte) bool { return bytes.HasPrefix(head, []byte("ID3")) || mp3Frame(head) }},
}

const headLength = 14

// mp3Frame says whether a file begins with the header of an MPEG audio frame of layer III: FF, a
// byte whose highest three bits are set and whose bits 2 and 1 are 01, and a byte whose highest four
// bits are not all set and whose bits 3 and 2 are not both set. The header is four bytes, and
// WHATWG's "match an mp3 header" refuses fewer (6.2.3, step 1; read 2026-11-04), as spec's
// format-check does.
func mp3Frame(head []byte) bool {
	return len(head) >= 4 && head[0] == 0xff && head[1]&0xe0 == 0xe0 && head[1]&0x06 == 0x02 &&
		head[2]&0xf0 != 0xf0 && head[2]&0x0c != 0x0c
}

// media checks module.media on each file of media/: that it stands at a path of the form media_path
// from media/, and begins with the signature of the format its extension names. What lies in media/
// is no concern of module.layout, which only places it.
func (c *checker) media() error {
	for _, file := range c.files {
		media, ok := strings.CutPrefix(file, mediaFolder)
		if !ok {
			continue
		}
		if !mediaPath.matches(media) {
			c.add(RuleMedia, file, 0, 0, "", fmt.Sprintf("%s is no media_path of common.schema.json: "+
				"names in kebab-case that begin with a letter, joined by /, and the extension .png, .jpg, .webp or .mp3", media))
			continue
		}
		head, err := readHead(c.fsys, file)
		if err != nil {
			return err
		}
		extension := strings.TrimPrefix(path.Ext(media), ".")
		if s := signatures[extension]; !s.matches(head) {
			c.add(RuleMedia, file, 0, 0, "", fmt.Sprintf("does not begin with the signature of %s, as its extension .%s says it does", s.format, extension))
		}
	}
	return nil
}

// readHead reads the first headLength bytes of a file, or all of a shorter one.
func readHead(fsys fs.FS, name string) ([]byte, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	head := make([]byte, headLength)
	n, err := io.ReadFull(f, head)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil
	}
	return head[:n], err
}

// texts checks module.license — LICENSE is there and holds something — and module.plain-text on
// LICENSE and NOTICE, which the package carries byte for byte.
func (c *checker) texts() error {
	if !slices.Contains(c.files, licensePath) {
		c.add(RuleLicense, moduleFolder, 0, 0, "", "the folder holds no file "+licensePath+": the text of the license that the manifest names")
	}
	for _, file := range []string{licensePath, noticePath} {
		if !slices.Contains(c.files, file) {
			continue
		}
		data, err := fs.ReadFile(c.fsys, file)
		if err != nil {
			return err
		}
		if file == licensePath && len(data) == 0 {
			c.add(RuleLicense, file, 0, 0, "", "is empty: it holds the text of the license that the manifest names")
		}
		c.plainText(file, data)
	}
	return nil
}

// plainText checks module.plain-text on a file: UTF-8, with no byte order mark, and no carriage
// return, since every line ends with LF alone. Each of the three is a breach once, at its first
// place: the line from 1, counted by LF alone, as plain text of the format ends a line, and the
// column, in characters, a byte that is no part of a UTF-8 character counting as one.
func (c *checker) plainText(file string, data []byte) {
	if bytes.HasPrefix(data, []byte("\ufeff")) {
		c.add(RulePlainText, file, 1, 1, "", "begins with a byte order mark")
	}
	invalid := false
	returns, returnLine, returnColumn := 0, 0, 0
	line, column := 1, 1
	for at := 0; at < len(data); {
		r, size := utf8.DecodeRune(data[at:])
		switch {
		case r == utf8.RuneError && size == 1 && !invalid:
			invalid = true
			c.add(RulePlainText, file, line, column, "", fmt.Sprintf("byte %d is not UTF-8", at+1))
		case r == '\r':
			if returns == 0 {
				returnLine, returnColumn = line, column
			}
			returns++
		}
		at += size
		if r == '\n' {
			line, column = line+1, 1
		} else {
			column++
		}
	}
	if returns > 0 {
		held := "a carriage return"
		if returns > 1 {
			held = fmt.Sprintf("%d carriage returns, the first here", returns)
		}
		c.add(RulePlainText, file, returnLine, returnColumn, "", "holds "+held+": every line ends with LF alone")
	}
}
