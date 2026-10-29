package util

import "golang.org/x/sys/unix"

func init() {
	hostArch = func() string {
		// the key is "1" on Apple Silicon even under Rosetta and is absent on Intel Macs
		if v, err := unix.SysctlUint32("hw.optional.arm64"); err == nil && v == 1 {
			return "arm64"
		}

		return "amd64"
	}
}
