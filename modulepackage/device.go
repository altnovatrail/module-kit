package modulepackage

import (
	"io/fs"
	"strings"

	"github.com/altnovatrail/module-kit/internal/canonical"
)

// The files of a module that go into its device part, the folder that stays out of its package,
// and the members of the listing of the device part (format/README.md, sections 10 and 25).
const (
	licenseFile   = "LICENSE"
	noticeFile    = "NOTICE"
	mediaFolder   = "media/"
	evalFolder    = "eval"
	licenseMember = "license"
	noticeMember  = "notice"
	mediaMember   = "media"
)

// deviceListing gives the listing of the device part of a module (format/README.md, section 25):
// LICENSE, NOTICE when the module has one, and each file of media/ by its path from media/, when it
// has any, each by its digest and size; and the files of the device part, by their paths in the
// folder of a package as files.
func deviceListing(fsys fs.FS, files []string) (map[string]any, map[string][]byte, error) {
	listing := map[string]any{}
	device := map[string][]byte{}
	media := map[string]any{}
	for _, file := range files {
		inMedia, isMedia := strings.CutPrefix(file, mediaFolder)
		if file != licenseFile && file != noticeFile && !isMedia {
			continue
		}
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, nil, err
		}
		device[DeviceFolder+file] = data
		d := descriptorValue(canonical.Descriptor(data))
		switch {
		case isMedia:
			media[inMedia] = d
		case file == licenseFile:
			listing[licenseMember] = d
		default:
			listing[noticeMember] = d
		}
	}
	if len(media) > 0 {
		listing[mediaMember] = media
	}
	return listing, device, nil
}
