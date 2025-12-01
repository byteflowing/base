package version

import (
	"fmt"
	"runtime/debug"
)

// Version information.
// 这些变量会在编译阶段将实际值注入
// 详细方法参考Makefile
var (
	BuildHash = "None"
	BuildTS   = "None"
	GitBranch = "None"
	Version   = "None"
	Service   = "None"
)

func GetVersion() string {
	return fmt.Sprintf("%s-%s-%s", Version, GitBranch, BuildHash)
}

func PrintVersion() {
	goVersion := "unknown"
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		goVersion = buildInfo.GoVersion
	}
	fmt.Printf("%-16s %s\n", "Name", Service)
	fmt.Printf("%-16s %s\n", "Version", Version)
	fmt.Printf("%-16s %s\n", "Git Branch", GitBranch)
	fmt.Printf("%-16s %s\n", "Build Hash", BuildHash)
	fmt.Printf("%-16s %s\n", "Build Time(UTC)", BuildTS)
	fmt.Printf("%-16s %s\n", "Go Version", goVersion)
}
