// DO NOT RUN DIRECTLY

package main

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	binDir     = "bin"
	filePrefix = "yt-proxy"
	goBin      = "go"
	myos       = "linux"
	myarch     = "amd64"
)

func main() {
	total := len(knownArch) * len(knownOS)
	for os_ := range knownOS {
		for arch := range knownArch {
			file := fmt.Sprintf("%s-%s-%s", filePrefix, os_, arch)
			filePath := fmt.Sprintf("../%s/%s", binDir, file)
			fmt.Printf("%d %-40s ", total, file)
			cmd := exec.Command(goBin, "build", "-o", filePath, "-ldflags", "-s -w")
			cmd.Env = os.Environ()
			if myos == os_ && myarch == arch {
				cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
			}
			cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", os_))
			cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", arch))
			if err := cmd.Run(); err == nil {
				fmt.Print("OK")
			}
			fmt.Println()
			total--
		}
	}
}

// https://github.com/golang/go/blob/master/src/internal/syslist/syslist.go

var knownOS = map[string]bool{
	"aix":       true,
	"android":   true,
	"darwin":    true,
	"dragonfly": true,
	"freebsd":   true,
	"hurd":      true,
	"illumos":   true,
	"ios":       true,
	"js":        true,
	"linux":     true,
	"nacl":      true,
	"netbsd":    true,
	"openbsd":   true,
	"plan9":     true,
	"solaris":   true,
	"wasip1":    true,
	"windows":   true,
	"zos":       true,
}

var knownArch = map[string]bool{
	"386":         true,
	"amd64":       true,
	"amd64p32":    true,
	"arm":         true,
	"armbe":       true,
	"arm64":       true,
	"arm64be":     true,
	"loong64":     true,
	"mips":        true,
	"mipsle":      true,
	"mips64":      true,
	"mips64le":    true,
	"mips64p32":   true,
	"mips64p32le": true,
	"ppc":         true,
	"ppc64":       true,
	"ppc64le":     true,
	"riscv":       true,
	"riscv64":     true,
	"s390":        true,
	"s390x":       true,
	"sparc":       true,
	"sparc64":     true,
	"wasm":        true,
}
