package picture

import (
	"bytes"
	"image/png"
	"strings"
	"testing"

	"golang.org/x/image/webp"
)

// Check reads a picture that anyone may write, so no file may make it panic, and none may make it
// decode more than maxPixels, which the fuzz test lowers so that a short input can reach it. Its
// warning agrees with the decoder: a file that decodes to a picture within the limit is warned of as
// transparent exactly when a pixel is (R33 in spec; R39).
func FuzzCheck(f *testing.F) {
	f.Add(".png", pngOf(f, filled(3, 2, white, 1, 1, clear)))
	f.Add(".png", pngOf(f, paletted(true)))
	f.Add(".webp", losslessWebP(4, 3, half))
	f.Add(".webp", extendedWebP(2, 2, white))
	f.Add(".webp", animatedWebP(2, 2))
	defer func(saved int) { maxPixels = saved }(maxPixels)
	maxPixels = 1 << 16
	f.Fuzz(func(t *testing.T, ext string, data []byte) {
		if ext != ".png" && ext != ".webp" {
			return
		}
		message, warn := Check("media/a"+ext, data)
		if !decoders[ext].signature(data) {
			if warn {
				t.Fatalf("a file with no signature, which module.media refuses, gives %q", message)
			}
			return
		}
		decode := png.Decode
		config := png.DecodeConfig
		if ext == ".webp" {
			decode, config = webp.Decode, webp.DecodeConfig
		}
		c, err := config(bytes.NewReader(data))
		if err != nil || c.Width*c.Height > maxPixels {
			if !warn {
				t.Fatalf("a file that cannot be read, or too large, gives no warning")
			}
			return
		}
		m, err := decode(bytes.NewReader(data))
		if err != nil {
			if !warn || !strings.Contains(message, "cannot decode") {
				t.Fatalf("a file that does not decode gives %q, %t", message, warn)
			}
			return
		}
		if _, _, _, transparent := firstTransparent(m); transparent != warn {
			t.Fatalf("a picture transparent %t gives %q, %t", transparent, message, warn)
		}
	})
}
