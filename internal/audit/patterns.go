package audit

import "regexp"

// Severity levels for audit findings.
const (
	SeverityCritical = "CRITICAL"
	SeverityHigh     = "HIGH"
	SeverityMedium   = "MEDIUM"
)

// rule defines a single scanning pattern.
type rule struct {
	Severity string
	Pattern  string // rule name
	Message  string
	Regex    *regexp.Regexp
}

// rules is the ordered list of all scanning patterns.
// CRITICAL rules can block installation; HIGH warns strongly; MEDIUM is informational.
var rules = []rule{
	// ── CRITICAL: prompt injection ──
	{SeverityCritical, "prompt-injection", "Prompt injection attempt detected",
		regexp.MustCompile(`(?i)(ignore\s+(all\s+)?previous\s+instructions|disregard\s+(all\s+)?rules|you\s+are\s+now\b|forget\s+everything|override\s+safety)`)},
	{SeverityCritical, "prompt-injection", "SYSTEM: prompt override detected",
		regexp.MustCompile(`(?m)^SYSTEM:`)},

	// ── CRITICAL: data exfiltration (curl/wget + secrets) ──
	{SeverityCritical, "data-exfiltration", "Command may exfiltrate sensitive data",
		regexp.MustCompile(`(?i)(curl|wget)\s+.*(\$SECRET|\$TOKEN|\$API_KEY|\$OPENAI_API_KEY|\$ANTHROPIC_API_KEY|\$AWS_SECRET)`)},
	{SeverityCritical, "data-exfiltration", "Command sends environment variables externally",
		regexp.MustCompile(`(?i)(curl|wget)\s+.*\$\{?(HOME|USER|PATH|SSH|GPG|AWS|GITHUB_TOKEN|OPENAI|ANTHROPIC)`)},

	// ── CRITICAL: credential access ──
	{SeverityCritical, "credential-access", "Accessing SSH private keys",
		regexp.MustCompile(`cat\s+~/?\.\s*ssh/(id_|known_hosts|authorized_keys|config)`)},
	{SeverityCritical, "credential-access", "Accessing .env secrets file",
		regexp.MustCompile(`cat\s+\.env\b`)},
	{SeverityCritical, "credential-access", "Accessing AWS credentials",
		regexp.MustCompile(`cat\s+~/?\.\s*aws/(credentials|config)`)},

	// ── HIGH: hidden unicode (zero-width characters) ──
	{SeverityHigh, "hidden-unicode", "Hidden zero-width Unicode characters detected",
		regexp.MustCompile("[\u200B\u200C\u200D\u2060\uFEFF]")},

	// ── HIGH: destructive commands ──
	{SeverityHigh, "destructive-commands", "Potentially destructive command",
		regexp.MustCompile(`(?i)\brm\s+-rf\s+(/(\s|$|\*)|\*|\./)`)},
	{SeverityHigh, "destructive-commands", "Unsafe permission change",
		regexp.MustCompile(`(?i)\bchmod\s+777\b`)},
	{SeverityHigh, "destructive-commands", "Sudo escalation",
		regexp.MustCompile(`(?i)\bsudo\s+`)},
	{SeverityHigh, "destructive-commands", "Disk overwrite command",
		regexp.MustCompile(`(?i)\bdd\s+if=`)},
	{SeverityHigh, "destructive-commands", "Filesystem format command",
		regexp.MustCompile(`(?i)\bmkfs\.`)},

	// ── HIGH: obfuscation ──
	{SeverityHigh, "obfuscation", "Base64 decode pipe may hide malicious content",
		regexp.MustCompile(`(?i)(base64\s+(-d|--decode)|\bbase64\b.*\|\s*(sh|bash|zsh|eval))`)},
	{SeverityHigh, "obfuscation", "Long base64-encoded string detected",
		regexp.MustCompile(`[A-Za-z0-9+/]{100,}={0,2}`)},

	// ── MEDIUM: suspicious URL usage (command + URL, not mere documentation links) ──
	// "fetch" excluded (too common in English). Localhost filtered in ScanContent.
	{SeverityMedium, "suspicious-fetch", "URL used in command context",
		regexp.MustCompile(`(?i)(curl|wget|invoke-webrequest|iwr)\s+['"]?https?://`)},

	// ── MEDIUM: system path writes ──
	{SeverityMedium, "system-writes", "References to system directories",
		regexp.MustCompile(`(?i)(write|copy|cp|mv|install)\s+.*/?(usr|etc|var|opt)/`)},
}
