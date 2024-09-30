package build

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"time"
)

type InfoStruct struct {
	NumGoRutines        int
	MemMbUsedAlloc      uint64
	MemMbUsedTotalAlloc uint64
	BuildDate           time.Time
	BuildDateStr        string
	GoVersion           string
	GitRevision         string

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
	d.BuildDateStr = d.BuildDate.Format("2006-01-02 15:04:05")

	idx := slices.IndexFunc(info.Settings, func(c debug.BuildSetting) bool {
		return c.Key == "vcs.revision"
	})
	if idx != -1 {
		d.GitRevision = info.Settings[idx].Value
	}

	return d, nil
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
