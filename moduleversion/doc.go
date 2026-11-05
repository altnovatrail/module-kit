// Package moduleversion compares a new version of a module with its published ones: it finds the
// part of the version that the change asks for, by the table of format/README.md, section 27, and
// reports every change at its place (R89 spec, R101 spec; R13, R36). The CLI's diff calls it, and
// module-registry from stage 4.
//
// Compare reads each published package by the chain of digests, without the files of its device
// part, and compares the new version with the newest published one: Changes gives the part and the
// changes of two packages of one module and one version of the format, each row of the table at its
// place and anything else that differs as a patch. Steps, StepPart and SameMajor give the rule of
// 0.y.z: the next versions after a version, the part a step raises, and whether two versions are of
// the same major, as diff and api read it. The rules module-version.* and the reading of a package
// of an earlier version of the format by the places every version keeps are not here yet: the epic
// of code brings them (R16).
package moduleversion
