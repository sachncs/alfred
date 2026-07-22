package runtime

import (
	"strings"
)

// TokenBudget tracks token usage and enforces limits.
type TokenBudget struct {
	maxInput   int
	maxOutput  int
	inputUsed  int
	outputUsed int
}

// NewTokenBudget creates a budget with the given limits. 0 = unlimited.
func NewTokenBudget(maxInput, maxOutput int) *TokenBudget {
	return &TokenBudget{maxInput: maxInput, maxOutput: maxOutput}
}

// EstimateTokens gives a rough token count (4 chars ≈ 1 token).
// ponytail: naive char/4, replace with tiktoken if accuracy matters.
func EstimateTokens(text string) int {
	n := len(text)
	if n == 0 {
		return 0
	}
	return (n + 3) / 4
}

// AddInput records input tokens.
func (b *TokenBudget) AddInput(n int) { b.inputUsed += n }

// AddOutput records output tokens.
func (b *TokenBudget) AddOutput(n int) { b.outputUsed += n }

// InputRemaining returns remaining input budget (-1 = unlimited).
func (b *TokenBudget) InputRemaining() int {
	if b.maxInput <= 0 {
		return -1
	}
	r := b.maxInput - b.inputUsed
	if r < 0 {
		return 0
	}
	return r
}

// OutputRemaining returns remaining output budget (-1 = unlimited).
func (b *TokenBudget) OutputRemaining() int {
	if b.maxOutput <= 0 {
		return -1
	}
	r := b.maxOutput - b.outputUsed
	if r < 0 {
		return 0
	}
	return r
}

// OverInput reports whether input budget is exceeded.
func (b *TokenBudget) OverInput() bool {
	if b.maxInput <= 0 {
		return false
	}
	return b.inputUsed > b.maxInput
}

// TruncateToBudget truncates text to fit within the remaining token budget.
func (b *TokenBudget) TruncateToBudget(text string, isOutput bool) string {
	var remaining int
	if isOutput {
		remaining = b.OutputRemaining()
	} else {
		remaining = b.InputRemaining()
	}
	if remaining < 0 {
		return text
	}
	maxChars := remaining * 4
	if maxChars <= 0 {
		return ""
	}
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars] + "...[truncated]"
}

// BuildPrompt constructs the model prompt from conversation history,
// respecting token budget.
func BuildPrompt(history []string, maxTokens int) string {
	if maxTokens <= 0 {
		return strings.Join(history, "\n\n")
	}
	maxChars := maxTokens * 4
	var b strings.Builder
	for i := len(history) - 1; i >= 0; i-- {
		entry := history[i]
		if b.Len()+len(entry) > maxChars {
			break
		}
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(entry)
	}
	return b.String()
}
