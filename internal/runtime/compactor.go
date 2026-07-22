package runtime

import (
	"github.com/alfred/alfred/internal/contract"
)

// CompactionMode controls how aggressively older turns are removed.
type CompactionMode string

const (
	CompactionNormal     CompactionMode = "normal"     // summarize, keep recent
	CompactionAggressive CompactionMode = "aggressive" // drop old turns, keep summary
	CompactionForce      CompactionMode = "force"      // truncate to minimum
)

// Compactor summarizes older turns to keep the context window under budget.
type Compactor struct {
	maxTurns     int
	summaryAfter int
}

// NewCompactor creates a Compactor.
func NewCompactor(maxTurns, summaryAfter int) *Compactor {
	if maxTurns <= 0 {
		maxTurns = 40
	}
	if summaryAfter <= 0 {
		summaryAfter = 20
	}
	return &Compactor{maxTurns: maxTurns, summaryAfter: summaryAfter}
}

// Compact applies the given mode to the turn list.
func (c *Compactor) Compact(turns []contract.Turn, mode CompactionMode) []contract.Turn {
	switch mode {
	case CompactionAggressive:
		return c.compactAggressive(turns)
	case CompactionForce:
		return c.compactForce(turns)
	default:
		return c.compactNormal(turns)
	}
}

// MaybeCompact applies normal compaction if over budget.
func (c *Compactor) MaybeCompact(turns []contract.Turn) []contract.Turn {
	if !c.ShouldCompact(len(turns)) {
		return turns
	}
	return c.compactNormal(turns)
}

// compactNormal keeps a summary turn + the most recent turns.
func (c *Compactor) compactNormal(turns []contract.Turn) []contract.Turn {
	if len(turns) <= c.summaryAfter {
		return turns
	}
	cutoff := len(turns) - c.summaryAfter
	compacted := []contract.Turn{makeSummaryTurn(turns[:cutoff])}
	compacted = append(compacted, turns[cutoff:]...)
	return compacted
}

// compactAggressive drops all but the last N/2 turns + a summary.
func (c *Compactor) compactAggressive(turns []contract.Turn) []contract.Turn {
	keep := c.summaryAfter / 2
	if keep < 2 {
		keep = 2
	}
	if len(turns) <= keep {
		return turns
	}
	cutoff := len(turns) - keep
	compacted := []contract.Turn{makeSummaryTurn(turns[:cutoff])}
	compacted = append(compacted, turns[cutoff:]...)
	return compacted
}

// compactForce keeps only the last 2 turns + a summary.
func (c *Compactor) compactForce(turns []contract.Turn) []contract.Turn {
	keep := 2
	if len(turns) <= keep {
		return turns
	}
	cutoff := len(turns) - keep
	compacted := []contract.Turn{makeSummaryTurn(turns[:cutoff])}
	compacted = append(compacted, turns[cutoff:]...)
	return compacted
}

// ShouldCompact reports whether compaction is needed.
func (c *Compactor) ShouldCompact(turnCount int) bool {
	return turnCount > c.summaryAfter
}

// SummaryCount returns the number of turns that were compacted away.
func SummaryCount(original, compacted int) int {
	d := original - compacted
	if d < 0 {
		return 0
	}
	return d
}

func makeSummaryTurn(compacted []contract.Turn) contract.Turn {
	text := "[compacted: "
	if len(compacted) == 1 {
		text += "1 older turn summarized]"
	} else {
		text += string(rune('0'+len(compacted))) + " older turns summarized]"
	}
	if len(compacted) > 9 {
		text = "[compacted: " + itoa(len(compacted)) + " older turns summarized]"
	}
	return contract.Turn{
		ID:     contract.TurnID("compact-summary"),
		Status: contract.TurnStatusCompleted,
		Items: []contract.TurnItem{{
			Kind: contract.ItemKindCompaction,
			Text: &text,
		}},
	}
}

// itoa is a minimal int-to-string for small positive numbers.
// ponytail: strconv.Itoa would work, this avoids the import for 3 lines.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
