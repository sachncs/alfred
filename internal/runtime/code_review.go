package runtime

import (
	"context"
	"fmt"

	"github.com/alfred/alfred/internal/model"
)

// CodeReviewer uses the model to review code changes.
type CodeReviewer struct {
	client model.Client
}

// NewCodeReviewer creates a reviewer.
func NewCodeReviewer(client model.Client) *CodeReviewer {
	return &CodeReviewer{client: client}
}

// ReviewRequest is the input for a code review.
type ReviewRequest struct {
	Files []ReviewFile
}

// ReviewFile is a file to review.
type ReviewFile struct {
	Path string
	Diff string
}

// ReviewResult is the output of a code review.
type ReviewResult struct {
	Summary string
	Issues  []ReviewIssue
}

// ReviewIssue is a single issue found during review.
type ReviewIssue struct {
	File    string
	Line    int
	Message string
	Severity string
}

// Review performs a code review on the given files.
func (r *CodeReviewer) Review(ctx context.Context, req ReviewRequest) (*ReviewResult, error) {
	if len(req.Files) == 0 {
		return &ReviewResult{Summary: "No files to review."}, nil
	}

	prompt := "Review these code changes for bugs, security issues, and style problems:\n\n"
	for _, f := range req.Files {
		prompt += fmt.Sprintf("File: %s\n%s\n\n", f.Path, f.Diff)
	}

	resp, err := r.client.Stream(ctx, model.Request{
		Messages: []model.Message{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return nil, err
	}

	var summary string
	for chunk := range resp {
		if chunk.DeltaText != "" {
			summary += chunk.DeltaText
		}
		if chunk.Error != "" {
			return nil, fmt.Errorf("model error: %s", chunk.Error)
		}
	}

	return &ReviewResult{Summary: summary}, nil
}
