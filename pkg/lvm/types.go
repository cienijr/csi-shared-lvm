package lvm

type LogicalVolume struct {
	Name string
	VG   string
	Size int64
	Tags []string
}
