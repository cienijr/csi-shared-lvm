package lvm

type Attr string
type LogicalVolume struct {
	Name string
	VG   string
	Size int64
	Tags []string
	Attr Attr
}

func (a Attr) IsActive() bool {
	return rune(a[4]) == 'a'
}

type VolumeGroup struct {
	Name     string
	FreeSize int64
}
