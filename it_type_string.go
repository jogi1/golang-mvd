// Code generated by "stringer -type=IT_TYPE"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[IT_SHOTGUN-1]
	_ = x[IT_SUPER_SHOTGUN-2]
	_ = x[IT_NAILGUN-4]
	_ = x[IT_SUPER_NAILGUN-8]
	_ = x[IT_GRENADE_LAUNCHER-16]
	_ = x[IT_ROCKET_LAUNCHER-32]
	_ = x[IT_LIGHTNING-64]
	_ = x[IT_SUPER_LIGHTNING-128]
	_ = x[IT_SHELLS-256]
	_ = x[IT_NAILS-512]
	_ = x[IT_ROCKETS-1024]
	_ = x[IT_CELLS-2048]
	_ = x[IT_AXE-4096]
	_ = x[IT_ARMOR1-8192]
	_ = x[IT_ARMOR2-16384]
	_ = x[IT_ARMOR3-32768]
	_ = x[IT_SUPERHEALTH-65536]
	_ = x[IT_KEY1-131072]
	_ = x[IT_KEY2-262144]
	_ = x[IT_INVISIBILITY-524288]
	_ = x[IT_INVULNERABILITY-1048576]
	_ = x[IT_SUIT-2097152]
	_ = x[IT_QUAD-4194304]
	_ = x[IT_UNKNOWN1-8388608]
	_ = x[IT_UNKNOWN2-16777216]
	_ = x[IT_UNKNOWN3-33554432]
	_ = x[IT_UNKNOWN4-67108864]
	_ = x[IT_UNKNOWN5-134217728]
	_ = x[IT_SIGIL1-268435456]
	_ = x[IT_SIGIL2-536870912]
	_ = x[IT_SIGIL3-1073741824]
	_ = x[IT_SIGIL4-2147483648]
}

const _IT_TYPE_name = "IT_SHOTGUNIT_SUPER_SHOTGUNIT_NAILGUNIT_SUPER_NAILGUNIT_GRENADE_LAUNCHERIT_ROCKET_LAUNCHERIT_LIGHTNINGIT_SUPER_LIGHTNINGIT_SHELLSIT_NAILSIT_ROCKETSIT_CELLSIT_AXEIT_ARMOR1IT_ARMOR2IT_ARMOR3IT_SUPERHEALTHIT_KEY1IT_KEY2IT_INVISIBILITYIT_INVULNERABILITYIT_SUITIT_QUADIT_UNKNOWN1IT_UNKNOWN2IT_UNKNOWN3IT_UNKNOWN4IT_UNKNOWN5IT_SIGIL1IT_SIGIL2IT_SIGIL3IT_SIGIL4"

var _IT_TYPE_map = map[IT_TYPE]string{
	1:          _IT_TYPE_name[0:10],
	2:          _IT_TYPE_name[10:26],
	4:          _IT_TYPE_name[26:36],
	8:          _IT_TYPE_name[36:52],
	16:         _IT_TYPE_name[52:71],
	32:         _IT_TYPE_name[71:89],
	64:         _IT_TYPE_name[89:101],
	128:        _IT_TYPE_name[101:119],
	256:        _IT_TYPE_name[119:128],
	512:        _IT_TYPE_name[128:136],
	1024:       _IT_TYPE_name[136:146],
	2048:       _IT_TYPE_name[146:154],
	4096:       _IT_TYPE_name[154:160],
	8192:       _IT_TYPE_name[160:169],
	16384:      _IT_TYPE_name[169:178],
	32768:      _IT_TYPE_name[178:187],
	65536:      _IT_TYPE_name[187:201],
	131072:     _IT_TYPE_name[201:208],
	262144:     _IT_TYPE_name[208:215],
	524288:     _IT_TYPE_name[215:230],
	1048576:    _IT_TYPE_name[230:248],
	2097152:    _IT_TYPE_name[248:255],
	4194304:    _IT_TYPE_name[255:262],
	8388608:    _IT_TYPE_name[262:273],
	16777216:   _IT_TYPE_name[273:284],
	33554432:   _IT_TYPE_name[284:295],
	67108864:   _IT_TYPE_name[295:306],
	134217728:  _IT_TYPE_name[306:317],
	268435456:  _IT_TYPE_name[317:326],
	536870912:  _IT_TYPE_name[326:335],
	1073741824: _IT_TYPE_name[335:344],
	2147483648: _IT_TYPE_name[344:353],
}

func (i IT_TYPE) String() string {
	if str, ok := _IT_TYPE_map[i]; ok {
		return str
	}
	return "IT_TYPE(" + strconv.FormatInt(int64(i), 10) + ")"
}
