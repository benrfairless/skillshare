package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/pterm/pterm"

	"skillshare/internal/config"
	"skillshare/internal/oplog"
	"skillshare/internal/ui"
)

const (
	logDetailTruncateLen = 96
	logTimeWidth         = 16
	logCmdWidth          = 9
	logStatusWidth       = 7
	logDurationWidth     = 7
	logMinWrapWidth      = 24

	logDetailPrefix = "  detail: "
)

var logANSIRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func cmdLog(args []string) error {
	mode, rest, err := parseModeArgs(args)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	if mode == modeAuto {
		if projectConfigExists(cwd) {
			mode = modeProject
		} else {
			mode = modeGlobal
		}
	}

	applyModeLabel(mode)

	if mode == modeProject {
		return cmdLogProject(rest, cwd)
	}

	return runLog(rest, config.ConfigPath())
}

func runLog(args []string, configPath string) error {
	auditOnly := false
	clear := false
	limit := 20

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--audit", "-a":
			auditOnly = true
		case "--clear", "-c":
			clear = true
		case "--tail", "-t":
			if i+1 < len(args) {
				i++
				n := 0
				if _, err := fmt.Sscanf(args[i], "%d", &n); err == nil && n > 0 {
					limit = n
				}
			}
		case "--help", "-h":
			printLogHelp()
			return nil
		}
	}

	if clear {
		filename := oplog.OpsFile
		label := "Operations"
		if auditOnly {
			filename = oplog.AuditFile
			label = "Audit"
		}
		if err := oplog.Clear(configPath, filename); err != nil {
			return fmt.Errorf("failed to clear log: %w", err)
		}
		ui.Success("%s log cleared", label)
		return nil
	}

	if auditOnly {
		return printLogSection(configPath, oplog.AuditFile, "Audit", limit)
	}

	if err := printLogSection(configPath, oplog.OpsFile, "Operations", limit); err != nil {
		return err
	}
	fmt.Println()
	return printLogSection(configPath, oplog.AuditFile, "Audit", limit)
}

func printLogSection(configPath, filename, label string, limit int) error {
	entries, err := oplog.Read(configPath, filename, limit)
	if err != nil {
		return fmt.Errorf("failed to read log: %w", err)
	}

	logPath := filepath.Join(oplog.LogDir(configPath), filename)
	if abs, absErr := filepath.Abs(logPath); absErr == nil {
		logPath = abs
	}
	mode := "global"
	if isProjectLogConfig(configPath) {
		mode = "project"
	}

	subtitle := fmt.Sprintf("%s (last %d)\nmode: %s\nfile: %s", label, len(entries), mode, logPath)
	ui.HeaderBox("skillshare log", subtitle)
	if len(entries) == 0 {
		ui.Info("No %s log entries", strings.ToLower(label))
		return nil
	}

	printLogEntries(entries)
	return nil
}

func printLogEntries(entries []oplog.Entry) {
	if ui.IsTTY() {
		printLogEntriesTTYTwoLine(os.Stdout, entries, logTerminalWidth())
		return
	}

	printLogEntriesNonTTY(os.Stdout, entries)
}

func printLogEntriesTTYTwoLine(w io.Writer, entries []oplog.Entry, termWidth int) {
	printLogTableHeaderTTY(w)

	for _, e := range entries {
		ts := formatLogTimestamp(e.Timestamp)
		cmd := padLogCell(strings.ToUpper(e.Command), logCmdWidth)
		status := colorizeLogStatusCell(padLogCell(e.Status, logStatusWidth), e.Status)
		dur := formatLogDuration(e.Duration)
		durCell := padLogCell(dur, logDurationWidth)

		fmt.Fprintf(w, "  %s%s%s | %s | %s | %s\n",
			ui.Gray,
			padLogCell(ts, logTimeWidth),
			ui.Reset,
			cmd,
			status,
			durCell,
		)

		detail := formatLogDetail(e, false)
		if detail != "" {
			printWrappedLogText(
				w,
				logDetailPrefix,
				detail,
				logWrapWidthForPrefix(termWidth, logDetailPrefix),
			)
		}

		printLogAuditSkillLinesTTY(w, e, termWidth)
	}
}

func printLogEntriesNonTTY(w io.Writer, entries []oplog.Entry) {
	for _, e := range entries {
		ts := formatLogTimestamp(e.Timestamp)
		detail := formatLogDetail(e, true)
		dur := formatLogDuration(e.Duration)

		fmt.Fprintf(w, "  %s  %-9s  %-96s  %-7s  %s\n",
			ts, e.Command, detail, e.Status, dur)

		printLogAuditSkillLinesNonTTY(w, e)
	}
}

func printLogAuditSkillLinesTTY(w io.Writer, e oplog.Entry, termWidth int) {
	if e.Command != "audit" || e.Args == nil {
		return
	}

	if failedSkills, ok := logArgStringSlice(e.Args, "failed_skills"); ok && len(failedSkills) > 0 {
		prefix := "  -> failed skills: "
		printWrappedLogText(w, prefix, strings.Join(failedSkills, ", "), logWrapWidthForPrefix(termWidth, prefix))
	}
	if warningSkills, ok := logArgStringSlice(e.Args, "warning_skills"); ok && len(warningSkills) > 0 {
		prefix := "  -> warning skills: "
		printWrappedLogText(w, prefix, strings.Join(warningSkills, ", "), logWrapWidthForPrefix(termWidth, prefix))
	}
}

func printLogAuditSkillLinesNonTTY(w io.Writer, e oplog.Entry) {
	if e.Command != "audit" || e.Args == nil {
		return
	}

	if failedSkills, ok := logArgStringSlice(e.Args, "failed_skills"); ok && len(failedSkills) > 0 {
		printLogNamedSkillsNonTTY(w, "failed skills", failedSkills)
	}
	if warningSkills, ok := logArgStringSlice(e.Args, "warning_skills"); ok && len(warningSkills) > 0 {
		printLogNamedSkillsNonTTY(w, "warning skills", warningSkills)
	}
}

func printLogNamedSkillsNonTTY(w io.Writer, label string, skills []string) {
	const namesPerLine = 4
	for i := 0; i < len(skills); i += namesPerLine {
		end := i + namesPerLine
		if end > len(skills) {
			end = len(skills)
		}

		currentLabel := label
		if i > 0 {
			currentLabel = label + " (cont)"
		}
		fmt.Fprintf(w, "                     -> %s: %s\n", currentLabel, strings.Join(skills[i:end], ", "))
	}
}

func printLogTableHeaderTTY(w io.Writer) {
	header := fmt.Sprintf("  %-16s | %-9s | %-7s | %-7s", "TIME", "CMD", "STATUS", "DUR")
	separator := fmt.Sprintf(
		"  %s-+-%s-+-%s-+-%s",
		strings.Repeat("-", logTimeWidth),
		strings.Repeat("-", logCmdWidth),
		strings.Repeat("-", logStatusWidth),
		strings.Repeat("-", logDurationWidth),
	)

	fmt.Fprintf(w, "%s%s%s\n", ui.Cyan, header, ui.Reset)
	fmt.Fprintf(w, "%s%s%s\n", ui.Gray, separator, ui.Reset)
}

func printWrappedLogText(w io.Writer, prefix, text string, width int) {
	lines := wrapLogText(text, width)
	if len(lines) == 0 {
		return
	}

	continuationPrefix := strings.Repeat(" ", logDisplayWidth(prefix))
	fmt.Fprintf(w, "%s%s%s%s\n", ui.Gray, prefix, ui.Reset, lines[0])
	for _, line := range lines[1:] {
		fmt.Fprintf(w, "%s%s%s%s\n", ui.Gray, continuationPrefix, ui.Reset, line)
	}
}

func formatLogTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		if len(ts) >= 16 {
			return ts[:16]
		}
		return ts
	}
	return t.Format("2006-01-02 15:04")
}

func formatLogDetail(e oplog.Entry, truncate bool) string {
	detail := ""
	if e.Args != nil {
		switch e.Command {
		case "sync":
			detail = formatSyncLogDetail(e.Args)
		case "audit":
			detail = formatAuditLogDetail(e.Args)
		default:
			detail = formatGenericLogDetail(e.Args)
		}
	}

	if e.Message != "" && detail != "" {
		return formatLogDetailValue(detail+" ("+e.Message+")", truncate)
	}
	if e.Message != "" {
		return formatLogDetailValue(e.Message, truncate)
	}
	if detail != "" {
		return formatLogDetailValue(detail, truncate)
	}
	return ""
}

func formatLogDetailValue(value string, truncate bool) string {
	if !truncate {
		return value
	}
	return truncateLogString(value, logDetailTruncateLen)
}

func formatSyncLogDetail(args map[string]any) string {
	parts := make([]string, 0, 5)

	if total, ok := logArgInt(args, "targets_total", "targets"); ok {
		parts = append(parts, fmt.Sprintf("targets=%d", total))
	}
	if failed, ok := logArgInt(args, "targets_failed"); ok && failed > 0 {
		parts = append(parts, fmt.Sprintf("failed=%d", failed))
	}
	if dryRun, ok := logArgBool(args, "dry_run"); ok && dryRun {
		parts = append(parts, "dry-run")
	}
	if force, ok := logArgBool(args, "force"); ok && force {
		parts = append(parts, "force")
	}
	if scope, ok := logArgString(args, "scope"); ok && scope != "" {
		parts = append(parts, "scope="+scope)
	}

	if len(parts) == 0 {
		return formatGenericLogDetail(args)
	}
	return strings.Join(parts, ", ")
}

func formatAuditLogDetail(args map[string]any) string {
	parts := make([]string, 0, 8)

	scope, hasScope := logArgString(args, "scope")
	name, hasName := logArgString(args, "name")
	if hasScope && scope == "single" && hasName && name != "" {
		parts = append(parts, "skill="+name)
	} else if hasScope && scope == "all" {
		parts = append(parts, "all-skills")
	} else if hasName && name != "" {
		parts = append(parts, name)
	}

	if mode, ok := logArgString(args, "mode"); ok && mode != "" {
		parts = append(parts, "mode="+mode)
	}
	if scanned, ok := logArgInt(args, "scanned"); ok {
		parts = append(parts, fmt.Sprintf("scanned=%d", scanned))
	}
	if passed, ok := logArgInt(args, "passed"); ok {
		parts = append(parts, fmt.Sprintf("passed=%d", passed))
	}
	if warning, ok := logArgInt(args, "warning"); ok && warning > 0 {
		parts = append(parts, fmt.Sprintf("warning=%d", warning))
	}
	if failed, ok := logArgInt(args, "failed"); ok && failed > 0 {
		parts = append(parts, fmt.Sprintf("failed=%d", failed))
	}

	critical, hasCritical := logArgInt(args, "critical")
	high, hasHigh := logArgInt(args, "high")
	medium, hasMedium := logArgInt(args, "medium")
	if (hasCritical && critical > 0) || (hasHigh && high > 0) || (hasMedium && medium > 0) {
		parts = append(parts, fmt.Sprintf("sev(c/h/m)=%d/%d/%d", critical, high, medium))
	}

	if scanErrors, ok := logArgInt(args, "scan_errors"); ok && scanErrors > 0 {
		parts = append(parts, fmt.Sprintf("scan-errors=%d", scanErrors))
	}

	if len(parts) == 0 {
		return formatGenericLogDetail(args)
	}
	return strings.Join(parts, ", ")
}

func formatGenericLogDetail(args map[string]any) string {
	parts := make([]string, 0, 4)

	if source, ok := logArgString(args, "source"); ok {
		parts = append(parts, source)
	}
	if name, ok := logArgString(args, "name"); ok {
		parts = append(parts, name)
	}
	if target, ok := logArgString(args, "target"); ok {
		parts = append(parts, target)
	}
	if targets, ok := logArgInt(args, "targets"); ok {
		parts = append(parts, fmt.Sprintf("targets=%d", targets))
	}
	if summary, ok := logArgString(args, "summary"); ok {
		parts = append(parts, summary)
	}

	return strings.Join(parts, ", ")
}

func logArgString(args map[string]any, key string) (string, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return "", false
	}

	switch s := v.(type) {
	case string:
		return s, true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

func logArgInt(args map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		v, ok := args[key]
		if !ok || v == nil {
			continue
		}

		switch n := v.(type) {
		case int:
			return n, true
		case int64:
			return int(n), true
		case float64:
			return int(n), true
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(n))
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func logArgBool(args map[string]any, key string) (bool, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return false, false
	}

	switch b := v.(type) {
	case bool:
		return b, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(b))
		if err == nil {
			return parsed, true
		}
	}
	return false, false
}

func logArgStringSlice(args map[string]any, key string) ([]string, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return nil, false
	}

	switch raw := v.(type) {
	case []string:
		if len(raw) == 0 {
			return nil, false
		}
		return raw, true
	case []any:
		items := make([]string, 0, len(raw))
		for _, it := range raw {
			s := strings.TrimSpace(fmt.Sprintf("%v", it))
			if s != "" {
				items = append(items, s)
			}
		}
		if len(items) == 0 {
			return nil, false
		}
		return items, true
	case string:
		s := strings.TrimSpace(raw)
		if s == "" {
			return nil, false
		}
		return []string{s}, true
	default:
		return nil, false
	}
}

func colorizeLogStatusCell(cell, status string) string {
	switch status {
	case "ok":
		return ui.Green + cell + ui.Reset
	case "error":
		return ui.Red + cell + ui.Reset
	case "partial":
		return ui.Yellow + cell + ui.Reset
	case "blocked":
		return ui.Red + cell + ui.Reset
	default:
		return cell
	}
}

func formatLogDuration(ms int64) string {
	if ms <= 0 {
		return ""
	}
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

func truncateLogString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func logTerminalWidth() int {
	width := pterm.GetTerminalWidth()
	if width > 0 {
		return width
	}
	return 120
}

func logWrapWidthForPrefix(termWidth int, prefix string) int {
	width := termWidth - logDisplayWidth(prefix)
	if width < logMinWrapWidth {
		return logMinWrapWidth
	}
	return width
}

func wrapLogText(text string, maxWidth int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if maxWidth <= 0 {
		return []string{text}
	}

	lines := make([]string, 0, 4)
	for len(text) > 0 {
		if logDisplayWidth(text) <= maxWidth {
			lines = append(lines, text)
			break
		}

		splitAt := 0
		currentWidth := 0
		lastSpace := -1
		for i, r := range text {
			rw := runewidth.RuneWidth(r)
			if currentWidth+rw > maxWidth {
				break
			}
			currentWidth += rw
			splitAt = i + utf8.RuneLen(r)
			if r == ' ' || r == '\t' {
				lastSpace = splitAt
			}
		}

		if splitAt == 0 {
			_, size := utf8.DecodeRuneInString(text)
			splitAt = size
		} else if lastSpace > 0 {
			splitAt = lastSpace
		}

		line := strings.TrimSpace(text[:splitAt])
		if line != "" {
			lines = append(lines, line)
		}

		text = strings.TrimLeft(text[splitAt:], " \t")
	}

	return lines
}

func padLogCell(value string, width int) string {
	current := logDisplayWidth(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func logDisplayWidth(s string) int {
	return runewidth.StringWidth(logANSIRegex.ReplaceAllString(s, ""))
}

// statusFromErr returns "ok" for nil errors and "error" otherwise.
// Used by all command instrumentation to derive oplog status.
func statusFromErr(err error) string {
	if err == nil {
		return "ok"
	}
	return "error"
}

func printLogHelp() {
	fmt.Println(`Usage: skillshare log [options]

View operations and audit logs for debugging and compliance.

Options:
  --audit, -a       Show only audit log
  --tail, -t <N>    Show last N entries (default: 20)
  --clear, -c       Clear the selected log file
  --project, -p     Use project-level log
  --global, -g      Use global log
  --help, -h        Show this help

Examples:
  skillshare log                  Show operations and audit logs
  skillshare log --audit          Show only audit log
  skillshare log --tail 50        Show last 50 entries per section
  skillshare log --clear          Clear operations log
  skillshare log --clear --audit  Clear audit log
  skillshare log -p               Show project operations and audit logs`)
}

func isProjectLogConfig(configPath string) bool {
	return filepath.Base(filepath.Dir(configPath)) == ".skillshare"
}
