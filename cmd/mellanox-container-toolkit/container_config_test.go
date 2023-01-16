package main

import (
	"testing"
)

func TestGetDevices(t *testing.T) {
	d, _ := getDevices("/dev", "")
	t.Errorf("%v\n", d)
}
