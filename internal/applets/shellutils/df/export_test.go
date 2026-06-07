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
