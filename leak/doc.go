// Package leak finds the keys and the leak phrases of a task in a text: in the tutor's words before
// they reach the device, for api, and in the texts of a module that a learner sees before solving,
// for the check of a module, which holds task.no-leaks and lesson.no-leaks by it (format/README.md,
// section 20; R54 spec, R57 spec; R30).
//
// For gives the Search of a task — the keys of its checks of a number, a fraction and a choice that
// the texts of their own items do not show, and its leak phrases — and Search.In what of them a text
// holds, each a Leak. A template is searched through its instances, so For refuses one,
// ErrTemplate. The normal form of a text is that of package internal/normal, which follows the
// version of Unicode of verdict.UnicodeVersion.
package leak
