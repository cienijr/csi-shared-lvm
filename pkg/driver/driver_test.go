package driver

import (
	"github.com/cienijr/csi-shared-lvm/pkg/lvm"
)

type mockLVM struct {
	getLV        func(vg, name string) (*lvm.LogicalVolume, error)
	createLV     func(vg, name string, size int64, tags []string) error
	deleteLV     func(vg, name string) error
	resizeLV     func(vg, name string, size int64) error
	activateLV   func(vg, name string) error
	deactivateLV func(vg, name string) error
}

func (m *mockLVM) GetLV(vg, name string) (*lvm.LogicalVolume, error) {
	return m.getLV(vg, name)
}

func (m *mockLVM) CreateLV(vg, name string, size int64, tags []string) error {
	return m.createLV(vg, name, size, tags)
}

func (m *mockLVM) DeleteLV(vg, name string) error {
	return m.deleteLV(vg, name)
}

func (m *mockLVM) ResizeLV(vg, name string, size int64) error {
	return m.resizeLV(vg, name, size)
}

func (m *mockLVM) ActivateLV(vg, name string) error {
	return m.activateLV(vg, name)
}

func (m *mockLVM) DeactivateLV(vg, name string) error {
	return m.deactivateLV(vg, name)
}
