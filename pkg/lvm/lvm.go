package lvm

import (
	"bytes"
	"fmt"
	"os/exec"
)

const (
	OwnershipTag = "csi-shared-lvm.cienijr.github.com"
)

type LVM interface {
	GetLV(vg, name string) (*LogicalVolume, error)
	CreateLV(vg, name string, size int64, tags []string) error
}
type client struct {
}

func NewLVM() LVM {
	return &client{}
}

func (c *client) CreateLV(vg, name string, size int64, tags []string) error {
	command, args := buildLvcreateCmd(vg, name, size, tags)
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create lv: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func (c *client) GetLV(vg, name string) (*LogicalVolume, error) {
	command, args := buildLvsCmd(vg, name)
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return parseLvsOutput(vg, stdout.String(), stderr.String(), err)
}
