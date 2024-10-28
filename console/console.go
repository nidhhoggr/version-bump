package console

import (
	"fmt"
	"os"
)

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
	colorCyan   string = "\033[36m"
)

func IncrementProjectVersion(isDryRun bool) {
	if isDryRun {
		fmt.Println("Dry Run: Incrementing project version...")
	} else {
		fmt.Println("Incrementing project version...")
	}

}

func CommittingChanges() {
	fmt.Println("Committing changes...")
}

func Language(name string, isDryRun bool) {
	action := "Updating"
	if isDryRun {
		action = "Will update"
	}
	fmt.Printf("  %s %v%v%v files:\n",
		action,
		colorCyan,
		name,
		colorReset,
	)
}

func VersionUpdate(oldVersion, newVersion, filepath string) string {
	return fmt.Sprintf("    %v%v%v -> %v%v%v %v\n",
		colorYellow, oldVersion, colorReset,
		colorGreen, newVersion, colorReset,
		filepath,
	)
}

func UpdateAvailable(version string, repoName string) {
	fmt.Printf("%vThe new version is available! Download from https://github.com/%s/releases/tag/%v%v\n",
		colorGreen, repoName, version, colorReset,
	)
}

func ErrorCheckingForUpdate(msg interface{}) {
	fmt.Printf("%vError checking for update: %v%v\n",
		colorYellow, msg, colorReset,
	)
}

func Error(msg interface{}) {
	fmt.Printf("%v%v%v\n",
		colorRed, msg, colorReset,
	)
}

func Fatal(msg interface{}) {
	Error(msg)
	os.Exit(1)
}
