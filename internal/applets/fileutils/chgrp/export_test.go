package chgrp

// LookupGidForTest exposes the unexported lookupGid helper to the external test
// package so it can verify group-name and numeric-gid resolution directly.
func LookupGidForTest(group string) (int, bool) {
	return lookupGid(group)
}
