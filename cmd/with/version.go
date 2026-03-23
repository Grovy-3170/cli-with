package main

import "runtime/debug"

// Version is set via ldflags at build time
var Version string

func init() {
	if Version == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			v := info.Main.Version
			if v != "" && v != "(devel)" {
				Version = v
			} else {
				Version = "dev"
			}
		} else {
			Version = "dev"
		}
	}
}
