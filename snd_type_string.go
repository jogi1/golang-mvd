// Code generated by "stringer -type=SND_TYPE"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SND_VOLUME-32768]
	_ = x[SND_ATTENUATION-16384]
}

const (
	_SND_TYPE_name_0 = "SND_ATTENUATION"
	_SND_TYPE_name_1 = "SND_VOLUME"
)

func (i SND_TYPE) String() string {
	switch {
	case i == 16384:
		return _SND_TYPE_name_0
	case i == 32768:
		return _SND_TYPE_name_1
	default:
		return "SND_TYPE(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
