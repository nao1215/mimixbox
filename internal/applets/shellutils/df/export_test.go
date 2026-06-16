package df

// This file exposes unexported internals to the external df_test package so the
// pure computations can be asserted directly and the syscall can be faked.

// StatfsResult mirrors the unexported statfsResult for use in tests.
type StatfsResult = statfsResult

// Usage mirrors the unexported usage for use in tests.
type Usage = usage

// InodeUsage mirrors the unexported inodeUsage for use in tests.
type InodeUsage = inodeUsage

// HumanReadable exposes humanReadable.
func HumanReadable(n uint64) string { return humanReadable(n) }

// FormatSize exposes formatSize.
func FormatSize(bytes uint64, human bool) string { return formatSize(bytes, human) }

// ComputeUsage exposes computeUsage.
func ComputeUsage(s StatfsResult) Usage { return computeUsage(s) }

// ComputeInodeUsage exposes computeInodeUsage.
func ComputeInodeUsage(s StatfsResult) InodeUsage { return computeInodeUsage(s) }

// Total/Used/Avail/UsePct expose the unexported usage fields.
func (u Usage) Total() uint64 { return u.total }
func (u Usage) Used() uint64  { return u.used }
func (u Usage) Avail() uint64 { return u.avail }
func (u Usage) UsePct() int   { return u.usePct }

// Files/IUsed/IFree/IUsePct expose the unexported inodeUsage fields.
func (u InodeUsage) Files() uint64 { return u.files }
func (u InodeUsage) IUsed() uint64 { return u.used }
func (u InodeUsage) IFree() uint64 { return u.free }
func (u InodeUsage) IUsePct() int  { return u.usePct }

// SetStatfs replaces the package statfs func and returns a restore func, letting
// a test inject a deterministic fake.
func SetStatfs(fake func(path string) (StatfsResult, error)) (restore func()) {
	orig := statfs
	statfs = fake
	return func() { statfs = orig }
}

// MountEntry mirrors the unexported mountEntry for use in tests.
type MountEntry = mountEntry

// NewMountEntry builds a mountEntry from its parts.
func NewMountEntry(source, target, fstype string) MountEntry {
	return mountEntry{source: source, target: target, fstype: fstype}
}

// SetReadMounts replaces the package readMounts func (the injectable mount
// source seam) and returns a restore func.
func SetReadMounts(fake func() ([]MountEntry, error)) (restore func()) {
	orig := readMounts
	readMounts = fake
	return func() { readMounts = orig }
}

// FsEntry mirrors the unexported fsEntry for use in tests.
type FsEntry = fsEntry

// NewFsEntry builds an fsEntry for the helpers under test.
func NewFsEntry(source, fstype, target string, s StatfsResult) FsEntry {
	return fsEntry{source: source, fstype: fstype, target: target, stat: s}
}

// ParseOutput exposes parseOutput.
func ParseOutput(spec string) ([]string, error) { return parseOutput(spec) }

// ParseSize exposes parseSize.
func ParseSize(spec string) (int64, error) { return parseSize(spec) }

// ScaleSize exposes scaleSize.
func ScaleSize(bytes uint64, human bool, blockSize int64) string {
	return scaleSize(bytes, human, blockSize)
}

// FilterByType exposes filterByType (returns the target fields for assertions).
func FilterByType(entries []FsEntry, types []string) []string {
	out := filterByType(entries, types)
	targets := make([]string, len(out))
	for i, e := range out {
		targets[i] = e.target
	}
	return targets
}

// UnescapeMount exposes unescapeMount.
func UnescapeMount(s string) string { return unescapeMount(s) }

// FsTypeName exposes fsTypeName.
func FsTypeName(magic int64) string { return fsTypeName(magic) }
