package lvm

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	lvNotFoundRegex = regexp.MustCompile(`Failed to find logical volume "(.*?)"`)
	vgNotFoundRegex = regexp.MustCompile(`Volume group "(.*?)" not found`)
)

type exitError interface {
	ExitCode() int
}

var _ exitError = &exec.ExitError{}

func parseLvsOutput(vg, stdout, stderr string, err error) (*LogicalVolume, error) {
	if err != nil {
		if isNotFound(err, stderr) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lv: %v, stderr: %s", err, stderr)
	}

	output := strings.TrimSpace(stdout)
	if output == "" {
		return nil, nil // LV doesn't exist
	}

	fields := strings.Fields(output)
	if len(fields) < 3 {
		return nil, fmt.Errorf("failed to parse lvs output: %s", output)
	}

	size, err := parseLVSize(fields[1])
	if err != nil {
		return nil, err
	}

	var tags []string
	if len(fields) > 3 {
		tags = strings.Split(fields[3], ",")
	}

	return &LogicalVolume{
		Name: fields[0],
		VG:   vg,
		Size: size,
		Tags: tags,
		Attr: fields[2],
	}, nil
}

func parseLVSize(sizeStr string) (int64, error) {
	return strconv.ParseInt(strings.TrimSuffix(sizeStr, "B"), 10, 64)
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
