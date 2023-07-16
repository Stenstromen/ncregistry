package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func ClearTerminal() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default: // For Unix-like systems
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ConvertToDaysAgo(timestamp string) (string, error) {
	layout := "2006-01-02T15:04:05.999999999Z" // Go layout string to parse the timestamp
	t, err := time.Parse(layout, timestamp)
	if err != nil {
		return "", err
	}

	duration := time.Since(t)
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%d days ago", days), nil
}

func FormatBytes(bytes int) string {
	const (
		B  = 1
		KB = B * 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	var (
		value float64
		unit  string
	)

	switch {
	case bytes >= GB:
		value = float64(bytes) / float64(GB)
		unit = "GB"
	case bytes >= MB:
		value = float64(bytes) / float64(MB)
		unit = "MB"
	case bytes >= KB:
		value = float64(bytes) / float64(KB)
		unit = "KB"
	default:
		value = float64(bytes)
		unit = "B"
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}
