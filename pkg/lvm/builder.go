package lvm

import "fmt"

func buildLvcreateCmd(vg, name string, size int64, tags []string) (string, []string) {
	args := []string{"--name", name, "--wipesignatures", "y", "--yes", "--size", fmt.Sprintf("%db", size), "--setautoactivation", "n"}
	for _, tag := range tags {
		args = append(args, "--addtag", tag)
	}
	args = append(args, vg)
	return "lvcreate", args
}

func buildLvsCmd(vg, name string) (string, []string) {
	args := []string{"--noheadings", "--nosuffix", "--units", "b", "-o", "lv_name,lv_size,lv_tags", fmt.Sprintf("%s/%s", vg, name)}
	return "lvs", args
}
