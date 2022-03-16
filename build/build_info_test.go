package build

import (
	"testing"
)

func TestBuildInfo(t *testing.T) {
	_, err := GetBuildInfo()
	if err != nil {
		t.Fatalf("error calling buildinfo: %v", err)
	}
}
