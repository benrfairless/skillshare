package ui

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/pterm/pterm"
)

// ansiRegex matches ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// displayWidth returns the visible width of a string (excluding ANSI codes, handling wide chars)
func displayWidth(s string) int {
	// Remove ANSI codes first, then calculate Unicode-aware width
	clean := ansiRegex.ReplaceAllString(s, "")
	return runewidth.StringWidth(clean)
}

// IsTTY returns true if stdout is a terminal
func IsTTY() bool {
	fi, _ := os.Stdout.Stat()
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Box prints content in a styled box
func Box(title string, lines ...string) {
	if !IsTTY() {
		if title != "" {
			fmt.Printf("── %s ──\n", title)
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		return
	}

	// Find max display width for consistent box width (excludes ANSI codes)
	maxLen := 0
	for _, line := range lines {
		w := displayWidth(line)
		if w > maxLen {
			maxLen = w
		}
	}

	// Pad lines to same display width
	content := ""
	for i, line := range lines {
		padded := line
		w := displayWidth(line)
		if w < maxLen {
			padded = line + strings.Repeat(" ", maxLen-w)
		}
		content += padded
		if i < len(lines)-1 {
			content += "\n"
		}
	}

	box := pterm.DefaultBox.WithTitle(title)
	box.Println(content)
}

// HeaderBox prints command header box
func HeaderBox(command, subtitle string) {
	if !IsTTY() {
		fmt.Printf("%s\n%s\n", command, subtitle)
		return
	}

	box := pterm.DefaultBox.
		WithTitle(pterm.Cyan(command)).
		WithTitleTopCenter()
	box.Println(subtitle)
}

// Spinner wraps pterm spinner with step tracking
type Spinner struct {
	spinner     *pterm.SpinnerPrinter
	start       time.Time
	currentStep int
	totalSteps  int
	stepPrefix  string
}

// StartSpinner starts a spinner with message
func StartSpinner(message string) *Spinner {
	if !IsTTY() {
		fmt.Printf("... %s\n", message)
		return &Spinner{start: time.Now()}
	}

	s, _ := pterm.DefaultSpinner.Start(message)
	return &Spinner{spinner: s, start: time.Now()}
}

// StartSpinnerWithSteps starts a spinner that shows step progress
func StartSpinnerWithSteps(message string, totalSteps int) *Spinner {
	if !IsTTY() {
		fmt.Printf("... [1/%d] %s\n", totalSteps, message)
		return &Spinner{start: time.Now(), currentStep: 1, totalSteps: totalSteps}
	}

	stepPrefix := fmt.Sprintf("[1/%d] ", totalSteps)
	s, _ := pterm.DefaultSpinner.Start(stepPrefix + message)
	return &Spinner{
		spinner:     s,
		start:       time.Now(),
		currentStep: 1,
		totalSteps:  totalSteps,
		stepPrefix:  stepPrefix,
	}
}

// Update updates spinner text
func (s *Spinner) Update(message string) {
	if s.spinner != nil {
		s.spinner.UpdateText(s.stepPrefix + message)
	} else {
		if s.totalSteps > 0 {
			fmt.Printf("... [%d/%d] %s\n", s.currentStep, s.totalSteps, message)
		} else {
			fmt.Printf("... %s\n", message)
		}
	}
}

// NextStep advances to next step and updates message
func (s *Spinner) NextStep(message string) {
	if s.totalSteps > 0 && s.currentStep < s.totalSteps {
		s.currentStep++
		s.stepPrefix = fmt.Sprintf("[%d/%d] ", s.currentStep, s.totalSteps)
	}
	s.Update(message)
}

// Success stops spinner with success
func (s *Spinner) Success(message string) {
	elapsed := time.Since(s.start)
	msg := fmt.Sprintf("%s (%.1fs)", message, elapsed.Seconds())
	if s.spinner != nil {
		s.spinner.Success(msg)
	} else {
		fmt.Printf("✓ %s\n", msg)
	}
}

// Fail stops spinner with failure (red)
func (s *Spinner) Fail(message string) {
	if s.spinner != nil {
		s.spinner.Fail(message)
	} else {
		fmt.Printf("✗ %s\n", message)
	}
}

// Warn stops spinner with warning (yellow)
func (s *Spinner) Warn(message string) {
	elapsed := time.Since(s.start)
	msg := fmt.Sprintf("%s (%.1fs)", message, elapsed.Seconds())
	if s.spinner != nil {
		s.spinner.Warning(msg)
	} else {
		fmt.Printf("! %s\n", msg)
	}
}

// Stop stops spinner without message
func (s *Spinner) Stop() {
	if s.spinner != nil {
		s.spinner.Stop()
	}
}

// SuccessMsg prints success message
func SuccessMsg(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if IsTTY() {
		pterm.Success.Println(msg)
	} else {
		fmt.Printf("✓ %s\n", msg)
	}
}

// ErrorMsg prints error message
func ErrorMsg(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if IsTTY() {
		pterm.Error.Println(msg)
	} else {
		fmt.Printf("✗ %s\n", msg)
	}
}

// WarningBox prints warning in a box
func WarningBox(title string, lines ...string) {
	if !IsTTY() {
		fmt.Printf("! %s\n", title)
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
		return
	}

	// Find max display width for consistent box width (excludes ANSI codes)
	maxLen := 0
	for _, line := range lines {
		w := displayWidth(line)
		if w > maxLen {
			maxLen = w
		}
	}

	// Pad lines to same display width
	content := ""
	for i, line := range lines {
		padded := line
		w := displayWidth(line)
		if w < maxLen {
			padded = line + strings.Repeat(" ", maxLen-w)
		}
		content += padded
		if i < len(lines)-1 {
			content += "\n"
		}
	}

	box := pterm.DefaultBox.
		WithTitle(pterm.Yellow(title)).
		WithBoxStyle(pterm.NewStyle(pterm.FgYellow))
	box.Println(content)
}

// SummaryBox prints a summary box with key-value pairs
func SummaryBox(title string, items map[string]string) {
	if !IsTTY() {
		fmt.Printf("── %s ──\n", title)
		for k, v := range items {
			fmt.Printf("  %s: %s\n", k, v)
		}
		return
	}

	var lines []string
	for k, v := range items {
		lines = append(lines, fmt.Sprintf("  %-10s %s", k+":", v))
	}

	// Find max display width for consistent box width (excludes ANSI codes)
	maxLen := 0
	for _, line := range lines {
		w := displayWidth(line)
		if w > maxLen {
			maxLen = w
		}
	}

	// Pad lines to same display width
	content := ""
	for i, line := range lines {
		padded := line
		w := displayWidth(line)
		if w < maxLen {
			padded = line + strings.Repeat(" ", maxLen-w)
		}
		content += padded
		if i < len(lines)-1 {
			content += "\n"
		}
	}

	box := pterm.DefaultBox.WithTitle(title)
	box.Println(content)
}

// ProgressBar wraps pterm progress bar
type ProgressBar struct {
	bar   *pterm.ProgressbarPrinter
	total int
}

// StartProgress starts a progress bar
func StartProgress(title string, total int) *ProgressBar {
	if !IsTTY() {
		fmt.Printf("%s (0/%d)\n", title, total)
		return &ProgressBar{total: total}
	}

	bar, _ := pterm.DefaultProgressbar.
		WithTotal(total).
		WithTitle(title).
		Start()
	return &ProgressBar{bar: bar, total: total}
}

// Increment increments progress
func (p *ProgressBar) Increment() {
	if p.bar != nil {
		p.bar.Increment()
	}
}

// UpdateTitle updates progress bar title
func (p *ProgressBar) UpdateTitle(title string) {
	if p.bar != nil {
		p.bar.UpdateTitle(title)
	} else {
		fmt.Printf("  %s\n", title)
	}
}

// Stop stops the progress bar
func (p *ProgressBar) Stop() {
	if p.bar != nil {
		p.bar.Stop()
	}
}

// UpdateNotification prints a colorful update notification
func UpdateNotification(currentVersion, latestVersion string) {
	if !IsTTY() {
		fmt.Printf("\n! Update available: %s -> %s\n", currentVersion, latestVersion)
		fmt.Println("  Run 'skillshare upgrade' to update")
		return
	}

	fmt.Println()

	// Build content lines
	lines := []string{
		"",
		fmt.Sprintf("  Version: %s -> %s", currentVersion, latestVersion),
		"",
		"  Run: skillshare upgrade",
		"",
	}

	// Find max display width for consistent box width (excludes ANSI codes)
	maxLen := 0
	for _, line := range lines {
		w := displayWidth(line)
		if w > maxLen {
			maxLen = w
		}
	}

	// Pad lines to same display width
	content := ""
	for i, line := range lines {
		padded := line
		w := displayWidth(line)
		if w < maxLen {
			padded = line + strings.Repeat(" ", maxLen-w)
		}
		content += padded
		if i < len(lines)-1 {
			content += "\n"
		}
	}

	box := pterm.DefaultBox.
		WithTitle(pterm.Yellow("Update Available")).
		WithBoxStyle(pterm.NewStyle(pterm.FgYellow))
	box.Println(content)
}

// SyncSummary prints a beautiful sync summary box
func SyncSummary(stats SyncStats) {
	if !IsTTY() {
		fmt.Printf("\n─── Sync Complete ───\n")
		fmt.Printf("  Targets: %d  Linked: %d  Local: %d  Updated: %d  Pruned: %d\n",
			stats.Targets, stats.Linked, stats.Local, stats.Updated, stats.Pruned)
		if stats.Duration > 0 {
			fmt.Printf("  Duration: %.1fs\n", stats.Duration.Seconds())
		}
		return
	}

	// Build stats line with colors
	statsLine := fmt.Sprintf(
		"  %s targets  %s linked  %s local  %s updated  %s pruned",
		pterm.Cyan(fmt.Sprint(stats.Targets)),
		pterm.Green(fmt.Sprint(stats.Linked)),
		pterm.Blue(fmt.Sprint(stats.Local)),
		pterm.Yellow(fmt.Sprint(stats.Updated)),
		pterm.Gray(fmt.Sprint(stats.Pruned)),
	)

	var durationLine string
	if stats.Duration > 0 {
		durationLine = fmt.Sprintf("  Completed in %s", pterm.Gray(fmt.Sprintf("%.1fs", stats.Duration.Seconds())))
	}

	// Build content with proper padding
	lines := []string{"", statsLine}
	if durationLine != "" {
		lines = append(lines, durationLine)
	}
	lines = append(lines, "")

	// Find max display width
	maxLen := 0
	for _, line := range lines {
		w := displayWidth(line)
		if w > maxLen {
			maxLen = w
		}
	}

	// Pad lines
	var content strings.Builder
	for i, line := range lines {
		padded := line
		w := displayWidth(line)
		if w < maxLen {
			padded = line + strings.Repeat(" ", maxLen-w)
		}
		content.WriteString(padded)
		if i < len(lines)-1 {
			content.WriteString("\n")
		}
	}

	box := pterm.DefaultBox.
		WithTitle(pterm.Green("✓ Sync Complete")).
		WithBoxStyle(pterm.NewStyle(pterm.FgGreen))
	box.Println(content.String())
}

// SyncStats holds statistics for sync summary
type SyncStats struct {
	Targets  int
	Linked   int
	Local    int
	Updated  int
	Pruned   int
	Duration time.Duration
}

// ListItem prints a list item with status
func ListItem(status, name, detail string) {
	var statusIcon string
	var style pterm.Style

	switch status {
	case "success":
		statusIcon = "✓"
		style = *pterm.NewStyle(pterm.FgGreen)
	case "error":
		statusIcon = "✗"
		style = *pterm.NewStyle(pterm.FgRed)
	case "warning":
		statusIcon = "!"
		style = *pterm.NewStyle(pterm.FgYellow)
	default:
		statusIcon = "→"
		style = *pterm.NewStyle(pterm.FgCyan)
	}

	if IsTTY() {
		fmt.Printf("  %s %-20s %s\n", style.Sprint(statusIcon), name, pterm.Gray(detail))
	} else {
		fmt.Printf("  %s %-20s %s\n", statusIcon, name, detail)
	}
}
