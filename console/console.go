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

var DebuggingEnabled bool

func init() {
	DebuggingEnabled = false
}

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
	fmt.Printf("\n  %s %v%v%v files:\n",
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

func VersionUpdateLine(oldVersion, newVersion, filepath string, line string) {
	fmt.Printf("    %v%v%v -> %v%v%v %v\n    Line: %s\n",
		colorYellow, oldVersion, colorReset,
		colorGreen, newVersion, colorReset,
		filepath,
		line,
	)
}

func VersionUpdateField(oldVersion, newVersion, filepath string, field string) {
	fmt.Printf("    %v%v%v -> %v%v%v %v\n    Field: %s\n",
		colorYellow, oldVersion, colorReset,
		colorGreen, newVersion, colorReset,
		filepath,
		field,
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

func Debug(from string, msg interface{}) {
	if !DebuggingEnabled {
		return
	}
	fmt.Printf("%vdebug:%v %v%s%v -> %v%v%v\n",
		colorYellow, colorReset,
		colorGreen, from, colorReset,
		colorYellow, msg, colorReset,
	)
}

func Error(msg interface{}) {
	fmt.Printf("%verror: %v%v\n",
		colorRed, msg, colorReset,
	)
}

func Fatal(msg interface{}) {
	Error(msg)
	os.Exit(1)
}
