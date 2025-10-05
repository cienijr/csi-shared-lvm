package lvm

import (
	"bytes"
	"fmt"
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

type Executor interface {
	Run(command string, args ...string) (stdout []byte, stderr []byte, err error)
}

type exitError interface {
	ExitCode() int
}

var _ exitError = &exec.ExitError{}

type client struct {
	exec Executor
}

func NewLVM(exec Executor) LVM {
	return &client{exec: exec}
}

type realExecutor struct{}

func NewRealExecutor() Executor {
	return &realExecutor{}
}

func (e *realExecutor) Run(command string, args ...string) ([]byte, []byte, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (c *client) CreateLV(vg, name string, size int64, tags []string) error {
	args := []string{"--name", name, "--wipesignatures", "y", "--yes", "--size", fmt.Sprintf("%db", size), "--setautoactivation", "n"}
	for _, tag := range tags {
		args = append(args, "--addtag", tag)
	}
	args = append(args, vg)
	_, stderr, err := c.exec.Run("lvcreate", args...)
	if err != nil {
		return fmt.Errorf("failed to create lv: %v, stderr: %s", err, string(stderr))
	}
	return nil
}

func (c *client) GetLV(vg, name string) (*LogicalVolume, error) {
	args := []string{"--noheadings", "--nosuffix", "--units", "b", "-o", "lv_name,lv_size,lv_tags", fmt.Sprintf("%s/%s", vg, name)}
	stdout, stderr, err := c.exec.Run("lvs", args...)

	if err != nil {
		errOutput := string(stderr)
		if isNotFound(err, errOutput) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lv: %v, stderr: %s", err, errOutput)
	}

	output := strings.TrimSpace(string(stdout))
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

func isNotFound(err error, stderr string) bool {
	exitErr, ok := err.(exitError)
	if !ok {
		return false
	}

	if exitErr.ExitCode() != 5 {
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
