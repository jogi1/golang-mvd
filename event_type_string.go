// Code generated by "stringer -type=Event_Type"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EPT_Spawn-0]
	_ = x[EPT_Death-1]
	_ = x[EPT_Suicide-2]
	_ = x[EPT_Kill-3]
	_ = x[EPT_Teamkill-4]
	_ = x[EPT_Pickup-5]
	_ = x[EPT_Drop-6]
}

const _Event_Type_name = "EPT_SpawnEPT_DeathEPT_SuicideEPT_KillEPT_TeamkillEPT_PickupEPT_Drop"

var _Event_Type_index = [...]uint8{0, 9, 18, 29, 37, 49, 59, 67}

func (i Event_Type) String() string {
	if i >= Event_Type(len(_Event_Type_index)-1) {
		return "Event_Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Event_Type_name[_Event_Type_index[i]:_Event_Type_index[i+1]]
}
