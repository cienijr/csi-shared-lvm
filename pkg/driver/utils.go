package driver

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getVGAndLVNames(volumeID string) (string, string, error) {
	parts := strings.Split(volumeID, "/")
	if len(parts) != 2 {
		return "", "", status.Errorf(codes.InvalidArgument, "invalid volume id: %s", volumeID)
	}
	return parts[0], parts[1], nil
}

func getDevicePath(volumeID string) (string, error) {
	vgName, lvName, err := getVGAndLVNames(volumeID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/dev/%s/%s", vgName, lvName), nil
}
