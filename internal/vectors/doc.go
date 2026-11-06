// Package vectors runs every example and test vector of the module format — spec.Format, at the
// version go.mod requires — and the vectors of the observation, observation/v1/ of
// contracts.Vectors, against module-kit: an implementation that does not pass them does not
// conform (format/README.md, sections 1, 8 and 9; contracts/README.md, section 10). It has no code
// of its own. The test is vectors_test.go, and it skips no case: every release passes them all
// (R16; R20; R42).
package vectors
