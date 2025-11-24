package pony

import (
	"fmt"
	"strconv"
	"strings"
)

// SizeConstraint represents a size constraint with units.
type SizeConstraint struct {
	value int
	unit  string // "", "%", "auto", "min", "max"
}

// parseSizeConstraint parses a size string like "50%", "20", "auto".
func parseSizeConstraint(s string) SizeConstraint {
	s = strings.TrimSpace(s)
	if s == "" {
		return SizeConstraint{unit: UnitAuto}
	}

	// Check for special keywords
	switch s {
	case UnitAuto:
		return SizeConstraint{unit: UnitAuto}
	case UnitMin:
		return SizeConstraint{unit: UnitMin}
	case UnitMax:
		return SizeConstraint{unit: UnitMax}
	}

	// Check for percentage
	if strings.HasSuffix(s, UnitPercent) {
		valStr := strings.TrimSuffix(s, UnitPercent)
		if val, err := strconv.Atoi(valStr); err == nil {
			return SizeConstraint{value: val, unit: UnitPercent}
		}
	}

	// Fixed size
	if val, err := strconv.Atoi(s); err == nil {
		return SizeConstraint{value: val, unit: ""}
	}

	// Invalid, default to auto
	return SizeConstraint{unit: UnitAuto}
}

// Apply applies the size constraint to get an actual size.
func (sc SizeConstraint) Apply(available, content int) int {
	// Zero value (unit="" and value=0) is treated as auto
	if sc.unit == "" && sc.value == 0 {
		if content > available {
			return available
		}
		return content
	}

	switch sc.unit {
	case "%":
		// Percentage of available space
		result := available * sc.value / 100
		if result < 0 {
			return 0
		}
		if result > available {
			return available
		}
		return result

	case UnitAuto:
		// Content-based sizing
		if content > available {
			return available
		}
		return content

	case UnitMin:
		// Minimum size (content or 0)
		if content < 0 {
			return 0
		}
		return content

	case UnitMax:
		// Maximum available size
		return available

	default:
		// Fixed size
		if sc.value < 0 {
			return 0
		}
		if sc.value > available {
			return available
		}
		return sc.value
	}
}

// String returns a string representation.
func (sc SizeConstraint) String() string {
	switch sc.unit {
	case UnitPercent:
		return fmt.Sprintf("%d%%", sc.value)
	case UnitAuto, UnitMin, UnitMax:
		return sc.unit
	default:
		return fmt.Sprintf("%d", sc.value)
	}
}

// IsAuto returns true if this is an auto constraint.
func (sc SizeConstraint) IsAuto() bool {
	// Zero value (unit="" and value=0) is auto
	// Otherwise unit must explicitly be "auto"
	return (sc.unit == "" && sc.value == 0) || sc.unit == UnitAuto
}

// IsFixed returns true if this is a fixed size constraint.
func (sc SizeConstraint) IsFixed() bool {
	return sc.unit == ""
}

// IsPercent returns true if this is a percentage constraint.
func (sc SizeConstraint) IsPercent() bool {
	return sc.unit == UnitPercent
}

// NewFixedConstraint creates a fixed size constraint.
func NewFixedConstraint(size int) SizeConstraint {
	return SizeConstraint{value: size, unit: ""}
}

// NewPercentConstraint creates a percentage constraint.
func NewPercentConstraint(percent int) SizeConstraint {
	return SizeConstraint{value: percent, unit: UnitPercent}
}
