// Package picture reads the pictures of a module for what the format asks of them but does not
// refuse: a picture SHOULD have an opaque background, and module-kit warns of one with a pixel that
// is not fully opaque (format/README.md, section 14; the second amendment of 2026-11-01 to R51 in
// spec). A warning is no rule: it has no ID and no test vectors, and refuses nothing (R39).
//
// Warnings reads every PNG and WebP file under media/ of a module's folder, decodes it — PNG by
// image/png, WebP by golang.org/x/image/webp — and gives a Warning for each picture with a pixel
// whose alpha is below full, naming the first such pixel; and for each picture it could not check: one
// larger than MaxPixels, which it does not decode, and one it cannot decode, such as an animated
// WebP. A JPEG has no transparency, and a sound none either.
//
// The CLI's check calls it, and module-registry will from stage 4. It stands apart from package
// module, so that api, which takes package module through modulepackage, carries no decoder of
// pictures (R38).
package picture
