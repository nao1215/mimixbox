package cal

import "time"

// Month exposes the unexported month renderer to the external test package.
func Month(year int, m time.Month, mondayFirst bool) string {
	return month(year, m, mondayFirst)
}
