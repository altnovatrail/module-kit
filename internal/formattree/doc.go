// Package formattree knows the layout of the folder of the module format, spec.Format: where its
// examples and test vectors lie, of which kind each case is, what a case folder holds, and which
// folders are laid out as a module and which files are files of a part on their own, as the format
// text lays them out. The vector test and the tests of the module check take it from here, so that
// a place where a new tag of spec puts a case comes to all of them at once. Only tests import it.
package formattree
