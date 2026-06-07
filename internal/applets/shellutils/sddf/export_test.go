package sddf

// FindDuplicatesForTest exposes the pure duplicate-detection core to the
// external test package. It returns, for every group of byte-identical files,
// the list of their paths.
func FindDuplicatesForTest(files []string) [][]string {
	groups := findDuplicates(files)
	out := make([][]string, 0, len(groups))
	for _, g := range groups {
		out = append(out, g.Paths)
	}
	return out
}
