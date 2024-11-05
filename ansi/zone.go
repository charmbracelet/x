package ansi

import "strconv"

// SetZone returns a sequence for setting the zone id of a cell.
//
// See: https://github.com/lrstanley/bubblezone
func SetZone(zone int) string {
	return "\x1b[" + strconv.Itoa(zone) + "z"
}
