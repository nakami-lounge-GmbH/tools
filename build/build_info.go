package build

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

type InfoStruct struct {
	NumGoRutines        int
	MemMbUsedAlloc      uint64
	MemMbUsedTotalAlloc uint64
	BuildDate           time.Time
	GoVersion           string

	Settings []debug.BuildSetting
}

func GetBuildInfo() (*InfoStruct, error) {
	d := new(InfoStruct)
	d.NumGoRutines = runtime.NumGoroutine()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	d.MemMbUsedAlloc = bToMb(m.Alloc)
	d.MemMbUsedTotalAlloc = bToMb(m.TotalAlloc)

	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("error reading Executable: %w", err)
	} else {
		fi, err := os.Stat(exe)
		if err != nil {
			return nil, fmt.Errorf("error reading stat: %w", err)
		}
		d.BuildDate = fi.ModTime()
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("no build info")
	}
	d.GoVersion = info.GoVersion
	d.Settings = info.Settings

	return d, nil
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
