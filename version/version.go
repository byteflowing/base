package version

import (
	"fmt"
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

func GetProductVersion() string {
	return Version
}

func PrintFullVersion() {
	fmt.Println("Service:			", Service)
	fmt.Println("Version:        	", Version)
	fmt.Println("Git Branch: 		", GitBranch)
	fmt.Println("Git Commit:		", BuildHash)
	fmt.Println("Build Time (UTC): 	", BuildTS)
}
