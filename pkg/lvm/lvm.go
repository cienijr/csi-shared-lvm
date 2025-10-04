package lvm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	OwnershipTag = "csi-shared-lvm.cienijr.github.com"
)

var (
	lvNotFoundRegex = regexp.MustCompile(`Failed to find logical volume "(.*?)"`)
	vgNotFoundRegex = regexp.MustCompile(`Volume group "(.*?)" not found`)
)

type LogicalVolume struct {
	Name string
	VG   string
	Size int64
	Tags []string
}

type LVM interface {
	GetLV(vg, name string) (*LogicalVolume, error)
	CreateLV(vg, name string, size int64, tags []string) error
}

type client struct{}

func NewLVM() LVM {
	return &client{}
}

func (c *client) CreateLV(vg, name string, size int64, tags []string) error {
	args := []string{"--name", name, "--wipesignatures", "y", "--yes", "--size", fmt.Sprintf("%db", size), "--setautoactivation", "n"}
	for _, tag := range tags {
		args = append(args, "--addtag", tag)
	}
	args = append(args, vg)
	cmd := exec.Command("lvcreate", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create lv: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func (c *client) GetLV(vg, name string) (*LogicalVolume, error) {
	args := []string{"--noheadings", "--nosuffix", "--units", "b", "-o", "lv_name,lv_size,lv_tags", fmt.Sprintf("%s/%s", vg, name)}
	cmd := exec.Command("lvs", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errOutput := stderr.String()
		if isNotFound(cmd.ProcessState, errOutput) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lv: %v, stderr: %s", err, errOutput)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil, nil
	}

	fields := strings.Fields(output)
	if len(fields) < 2 {
		return nil, fmt.Errorf("failed to parse lvs output: %s", output)
	}

	size, err := parseLVSize(fields[1])
	if err != nil {
		return nil, err
	}

	var tags []string
	if len(fields) > 2 {
		tags = strings.Split(fields[2], ",")
	}

	return &LogicalVolume{
		Name: fields[0],
		VG:   vg,
		Size: size,
		Tags: tags,
	}, nil
}

func isNotFound(state *os.ProcessState, stderr string) bool {
	if state.ExitCode() != 5 {
		// exit code for this error should be 5
		return false
	}

	if lvNotFoundRegex.MatchString(stderr) {
		return true
	}

	if vgNotFoundRegex.MatchString(stderr) {
		return true
	}

	return false
}

func parseLVSize(sizeStr string) (int64, error) {
	return strconv.ParseInt(strings.TrimSuffix(sizeStr, "B"), 10, 64)
}
