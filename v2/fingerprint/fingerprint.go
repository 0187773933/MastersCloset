// Package fingerprint reproduces v1's per-machine identifier so v2 resolves the
// same ~/.config/mct/mct_<fingerprint>.db file v1 created — letting an existing
// install's data be picked up automatically. The string layout and hashing must
// match v1's utils.FingerPrintPassive exactly or the names won't line up.
package fingerprint

import (
	sha256 "crypto/sha256"
	hex "encoding/hex"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"

	cpu "github.com/shirou/gopsutil/cpu"
)

func cpuInfo() string {
	info, err := cpu.Info()
	if err != nil || len(info) == 0 {
		return ""
	}
	i := info[0]
	var parts []string
	if i.VendorID != "" {
		parts = append(parts, i.VendorID)
	}
	if i.Family != "" {
		parts = append(parts, i.Family)
	}
	if i.Model != "" {
		parts = append(parts, i.Model)
	}
	if i.Cores > 0 {
		parts = append(parts, fmt.Sprintf("%d", i.Cores))
	}
	if i.ModelName != "" {
		parts = append(parts, i.ModelName)
	}
	return strings.Join(parts, " === ")
}

// Full returns the human-readable fingerprint string (username === os === arch
// === hostname === cpu).
func Full() string {
	username := ""
	if u, err := user.Current(); err == nil && u != nil {
		username = u.Username
	}
	host, _ := os.Hostname()
	return fmt.Sprintf("%s === %s === %s === %s === %s",
		username, runtime.GOOS, runtime.GOARCH, host, cpuInfo())
}

// Short returns the first 10 hex chars of sha256(Full()) — the token v1 used to
// name per-machine data files.
func Short() string {
	sum := sha256.Sum256([]byte(Full()))
	return hex.EncodeToString(sum[:])[0:10]
}
