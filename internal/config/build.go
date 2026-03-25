package config

import "fmt"

func na(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}

func PrintBuildInfo(buildVersion, buildDate, buildCommit string) {
	fmt.Printf("Build version: %s\n", na(buildVersion))
	fmt.Printf("Build date: %s\n", na(buildDate))
	fmt.Printf("Build commit: %s\n", na(buildCommit))
}
