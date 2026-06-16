package cal

import "time"

// Month exposes the unexported month renderer to the external test package.
func Month(year int, m time.Month, mondayFirst bool) string {
	return month(year, m, mondayFirst)
}

// Center exposes the unexported center helper so the external test package can
// cover its no-padding (len(s) >= width) branch directly.
func Center(s string, width int) string {
	return center(s, width)
}
