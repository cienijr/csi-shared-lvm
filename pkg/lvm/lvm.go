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
	DeleteLV(vg, name string) error
	ResizeLV(vg, name string, size int64) error
	ActivateLV(vg, name string) error
	DeactivateLV(vg, name string) error
	GetVG(name string) (*VolumeGroup, error)
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

func (c *client) DeleteLV(vg, name string) error {
	command, args := buildLvremoveCmd(vg, name)
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete lv: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func (c *client) ResizeLV(vg, name string, size int64) error {
	command, args := buildLvextendCmd(vg, name, size)
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to resize lv: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func (c *client) ActivateLV(vg, name string) error {
	command, args := buildLvchangeActivateCmd(vg, name)
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to activate lv: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func (c *client) DeactivateLV(vg, name string) error {
	command, args := buildLvchangeDeactivateCmd(vg, name)
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deactivate lv: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func (c *client) GetVG(name string) (*VolumeGroup, error) {
	command, args := buildVgsCmg(name)
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return parseVgsOutput(stdout.String(), stderr.String(), err)
}
