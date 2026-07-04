// Package safety classifies shell commands by how destructive they may be.
package safety

import "regexp"

// Level describes how risky a command is.
type Level int

const (
	// Safe commands are read-only or otherwise low-impact.
	Safe Level = iota
	// Caution commands modify state but are usually recoverable.
	Caution
	// Danger commands can cause irreversible data loss or system damage.
	Danger
)

// String returns a short label for the level.
func (l Level) String() string {
	switch l {
	case Danger:
		return "DANGER"
	case Caution:
		return "CAUTION"
	default:
		return "SAFE"
	}
}

type rule struct {
	re    *regexp.Regexp
	level Level
}

// rules are evaluated in order; the highest matched level wins.
var rules = []rule{
	// Irreversible / destructive.
	{regexp.MustCompile(`\brm\b.*(-[a-zA-Z]*f|-[a-zA-Z]*r)`), Danger},
	{regexp.MustCompile(`\brm\s+-[a-zA-Z]*\s*/\b`), Danger},
	{regexp.MustCompile(`\bdd\b`), Danger},
	{regexp.MustCompile(`\bmkfs\b`), Danger},
	{regexp.MustCompile(`\b(shutdown|reboot|halt|poweroff)\b`), Danger},
	{regexp.MustCompile(`>\s*/dev/sd[a-z]`), Danger},
	{regexp.MustCompile(`:\(\)\s*\{.*\};`), Danger}, // fork bomb
	{regexp.MustCompile(`\bchmod\b.*-R.*\b777\b`), Danger},
	{regexp.MustCompile(`\bgit\b.*\b(reset\s+--hard|clean\s+-[a-zA-Z]*f|push\s+.*--force)`), Danger},

	// Modifies state but usually recoverable.
	{regexp.MustCompile(`\brm\b`), Caution},
	{regexp.MustCompile(`\b(mv|cp)\b`), Caution},
	{regexp.MustCompile(`\b(chmod|chown)\b`), Caution},
	{regexp.MustCompile(`\b(kill|pkill|killall)\b`), Caution},
	{regexp.MustCompile(`\bsudo\b`), Caution},
	{regexp.MustCompile(`>>?`), Caution}, // output redirection writes files
	{regexp.MustCompile(`\b(apt|apt-get|brew|npm|pip|pip3|yum|dnf)\b.*\b(install|remove|uninstall)\b`), Caution},
}

// Classify returns the risk Level for a command string.
func Classify(command string) Level {
	highest := Safe
	for _, r := range rules {
		if r.re.MatchString(command) && r.level > highest {
			highest = r.level
		}
	}
	return highest
}
