package picture

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"golang.org/x/image/webp"
)

// pngOf encodes m as PNG.
func pngOf(t testing.TB, m image.Image) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, m); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// filled gives a w × h image of c, with the pixel at (px, py) of other when px is not negative.
func filled(w, h int, c color.NRGBA, px, py int, other color.NRGBA) *image.NRGBA {
	m := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			m.SetNRGBA(x, y, c)
		}
	}
	if px >= 0 {
		m.SetNRGBA(px, py, other)
	}
	return m
}

var (
	white = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	clear = color.NRGBA{}
	half  = color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0x80}
)

// paletted gives a 3 × 2 paletted image whose palette has a transparent entry, which png.Encode
// writes into a tRNS chunk; used says whether a pixel takes that entry.
func paletted(used bool) *image.Paletted {
	m := image.NewPaletted(image.Rect(0, 0, 3, 2), color.Palette{white, color.NRGBA{}})
	if used {
		m.SetColorIndex(2, 1, 1)
	}
	return m
}

// The WebP files of the tests decode as the tests take them to: by golang.org/x/image/webp, to the
// size and the opacity they are built with, and the animated one not at all.
func TestTestFiles(t *testing.T) {
	for _, c := range []struct {
		name   string
		data   []byte
		w, h   int
		opaque bool
	}{
		{"lossless opaque", losslessWebP(4, 3, white), 4, 3, true},
		{"lossless transparent", losslessWebP(4, 3, half), 4, 3, false},
		{"extended opaque", extendedWebP(2, 5, white), 2, 5, true},
		{"extended transparent", extendedWebP(2, 5, clear), 2, 5, false},
	} {
		m, err := webp.Decode(bytes.NewReader(c.data))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		_, _, _, transparent := firstTransparent(m)
		if b := m.Bounds(); b.Dx() != c.w || b.Dy() != c.h || transparent == c.opaque {
			t.Errorf("%s: %v, opaque %t; want %d × %d, opaque %t", c.name, b, !transparent, c.w, c.h, c.opaque)
		}
	}
	if _, err := webp.Decode(bytes.NewReader(animatedWebP(2, 2))); err == nil {
		t.Error("the animated file decodes; the tests take it for one golang.org/x/image does not read")
	}
}

func TestCheck(t *testing.T) {
	for _, c := range []struct {
		name, file string
		data       []byte
		want       string // a part of the warning, or "" for none
	}{
		{"PNG opaque", "a.png", pngOf(t, filled(3, 3, white, -1, 0, white)), ""},
		{"PNG with an alpha channel, all opaque", "a.png", pngOf(t, filled(3, 3, color.NRGBA{R: 9, A: 0xff}, 2, 2, white)), ""},
		{"PNG with one transparent pixel", "a.png", pngOf(t, filled(3, 3, white, 2, 1, clear)), "the pixel at x = 2, y = 1 has alpha 0 of 65535"},
		{"PNG half transparent", "a.png", pngOf(t, filled(3, 3, half, -1, 0, half)), "the pixel at x = 0, y = 0 has alpha 32896 of 65535"},
		{"PNG with tRNS, unused", "a.png", pngOf(t, paletted(false)), ""},
		{"PNG with tRNS, used", "a.png", pngOf(t, paletted(true)), "the pixel at x = 2, y = 1"},
		{"PNG gray", "a.png", pngOf(t, image.NewGray(image.Rect(0, 0, 2, 2))), ""},
		{"WebP lossless opaque", "a.webp", losslessWebP(4, 3, white), ""},
		{"WebP lossless transparent", "a.webp", losslessWebP(4, 3, half), "the pixel at x = 0, y = 0 has alpha 32896"},
		{"WebP extended opaque", "a.webp", extendedWebP(2, 5, white), ""},
		{"WebP extended transparent", "a.webp", extendedWebP(2, 5, clear), "has alpha 0"},
		{"WebP animated", "a.webp", animatedWebP(2, 2), "module-kit cannot decode it as WebP"},
		{"PNG that is not", "a.png", []byte("\x89PNG\r\n\x1a\nnot a picture"), "module-kit cannot read it as PNG"},
		{"WebP that is not", "a.webp", []byte("RIFF\x04\x00\x00\x00WEBPVP8L"), "module-kit cannot read it as WebP"},
		{"JPEG", "a.jpg", []byte{0xff, 0xd8, 0xff}, ""},
		{"sound", "a.mp3", []byte("ID3"), ""},
		{"PNG with no signature, which module.media refuses", "a.png", []byte{0xff, 0xd8, 0xff}, ""},
		{"WebP with no signature, which module.media refuses", "a.webp", []byte("RIFF\x04\x00\x00\x00WAVEfmt "), ""},
		{"PNG truncated after its signature", "a.png", []byte("\x89PNG\r\n\x1a\n"), "module-kit cannot read it as PNG"},
	} {
		t.Run(c.name, func(t *testing.T) {
			message, warn := Check("media/"+c.file, c.data)
			if warn != (c.want != "") || !strings.Contains(message, c.want) {
				t.Errorf("Check = %q, %t; want a warning that holds %q", message, warn, c.want)
			}
		})
	}
}

// A picture larger than maxPixels is not decoded, but warned of, by its size alone.
func TestCheckTooLarge(t *testing.T) {
	defer func(saved int) { maxPixels = saved }(maxPixels)
	maxPixels = 11
	for file, data := range map[string][]byte{"a.png": pngOf(t, filled(4, 3, white, -1, 0, white)), "a.webp": losslessWebP(4, 3, white)} {
		message, warn := Check(file, data)
		if want := "it is 4 × 3 pixels, more than the 11 that module-kit decodes"; !warn || !strings.HasPrefix(message, want) {
			t.Errorf("%s: Check = %q, %t; want %q", file, message, warn, want)
		}
	}
	if MaxPixels != 1<<25 {
		t.Errorf("MaxPixels is %d, and R39 says 1<<25", MaxPixels)
	}
}
