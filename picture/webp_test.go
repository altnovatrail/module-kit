package picture

import (
	"encoding/binary"
	"image/color"
)

// The WebP files of the tests are built here, after RFC 9649, "WebP Image Format" (November 2024),
// rather than kept as files whose making no one could repeat: a lossless image of one colour, alone
// (section 2.5, the simple file format) or in an extended file with VP8X (section 2.7), and an
// extended file that says it is animated, whose frames the decoder of golang.org/x/image does not
// read.

// bits writes a VP8L bitstream, least significant bit first (RFC 9649, section 3.3).
type bits struct {
	out   []byte
	acc   uint64
	count uint
}

func (b *bits) write(value uint32, n uint) {
	b.acc |= uint64(value) << b.count
	b.count += n
	for b.count >= 8 {
		b.out = append(b.out, byte(b.acc))
		b.acc >>= 8
		b.count -= 8
	}
}

func (b *bits) bytes() []byte {
	if b.count > 0 {
		b.out = append(b.out, byte(b.acc))
		b.acc, b.count = 0, 0
	}
	return b.out
}

// vp8l gives the VP8L bitstream of a width × height image whose every pixel is c: the header
// (section 3.2), no transform (section 4), no colour cache and no meta prefix codes (section 5.2),
// and five simple prefix codes of one symbol each, so that every pixel takes no bits (section 6.2.1).
func vp8l(width, height int, c color.NRGBA) []byte {
	var b bits
	b.write(0x2f, 8)
	b.write(uint32(width-1), 14)
	b.write(uint32(height-1), 14)
	alphaUsed := uint32(0)
	if c.A != 0xff {
		alphaUsed = 1
	}
	b.write(alphaUsed, 1)
	b.write(0, 3) // version
	b.write(0, 1) // no transform
	b.write(0, 1) // no colour cache
	b.write(0, 1) // no meta prefix codes
	for _, symbol := range []uint8{c.G, c.R, c.B, c.A, 0} {
		b.write(1, 1) // a simple code
		b.write(0, 1) // of one symbol
		b.write(1, 1) // of 8 bits
		b.write(uint32(symbol), 8)
	}
	return b.bytes()
}

// chunk gives a RIFF chunk: its FourCC, its size and its data, padded to an even size (section 2.3).
func chunk(fourCC string, data []byte) []byte {
	out := append([]byte(fourCC), binary.LittleEndian.AppendUint32(nil, uint32(len(data)))...)
	out = append(out, data...)
	if len(data)%2 == 1 {
		out = append(out, 0)
	}
	return out
}

// riff gives a WebP file of its chunks (section 2.4).
func riff(chunks ...[]byte) []byte {
	var body []byte
	for _, c := range chunks {
		body = append(body, c...)
	}
	out := append([]byte("RIFF"), binary.LittleEndian.AppendUint32(nil, uint32(4+len(body)))...)
	out = append(out, "WEBP"...)
	return append(out, body...)
}

// vp8x gives the VP8X chunk of an extended file, with the flags given, of a width × height canvas
// (section 2.7).
func vp8x(flags byte, width, height int) []byte {
	data := []byte{flags, 0, 0, 0}
	w, h := uint32(width-1), uint32(height-1)
	data = append(data, byte(w), byte(w>>8), byte(w>>16), byte(h), byte(h>>8), byte(h>>16))
	return chunk("VP8X", data)
}

// The flags of VP8X (section 2.7).
const (
	animationFlag = 1 << 1
	alphaFlag     = 1 << 4
)

// losslessWebP gives a simple lossless WebP file of a width × height image of the colour c.
func losslessWebP(width, height int, c color.NRGBA) []byte {
	return riff(chunk("VP8L", vp8l(width, height, c)))
}

// extendedWebP gives an extended WebP file, with alpha, holding the same lossless image.
func extendedWebP(width, height int, c color.NRGBA) []byte {
	return riff(vp8x(alphaFlag, width, height), chunk("VP8L", vp8l(width, height, c)))
}

// animatedWebP gives an extended WebP file that says it is animated: an ANIM chunk and one ANMF
// frame of a lossless image (section 2.7.1.1).
func animatedWebP(width, height int) []byte {
	anim := chunk("ANIM", []byte{0xff, 0xff, 0xff, 0xff, 0, 0})
	frame := []byte{0, 0, 0, 0, 0, 0}
	w, h := uint32(width-1), uint32(height-1)
	frame = append(frame, byte(w), byte(w>>8), byte(w>>16), byte(h), byte(h>>8), byte(h>>16), 0, 0, 0, 0)
	frame = append(frame, chunk("VP8L", vp8l(width, height, color.NRGBA{R: 0xff, A: 0xff}))...)
	return riff(vp8x(animationFlag|alphaFlag, width, height), anim, chunk("ANMF", frame))
}
