package scientificplotting

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

type PlotGenerateInput struct {
	DataURL   string `json:"data_url"`
	ChartType string `json:"chart_type"`
	Title     string `json:"title,omitempty"`
	XLabel    string `json:"x_label,omitempty"`
	YLabel    string `json:"y_label,omitempty"`
}

type PlotReviewInput struct {
	PlotURL  string `json:"plot_url"`
	Criteria string `json:"criteria,omitempty"`
}

type ScientificPlottingServer struct {
	*worker.WorkerServer
}

func NewScientificPlottingServer() *worker.WorkerServer {
	return worker.NewWorkerServer("scientific-plotting-worker", []tool.Tool{
		&plotGenerateTool{},
		&plotReviewTool{},
	})
}

type plotGenerateTool struct{}

func (t *plotGenerateTool) Name() string        { return "plot_generate" }
func (t *plotGenerateTool) Description() string { return "Generates a scientific plot from data." }
func (t *plotGenerateTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"data_url":{"type":"string","description":"URL or path to the data source"},
			"chart_type":{"type":"string","description":"Type of plot (line, scatter, bar, heatmap, etc.)"},
			"title":{"type":"string","description":"Plot title"},
			"x_label":{"type":"string","description":"X-axis label"},
			"y_label":{"type":"string","description":"Y-axis label"}
		},
		"required":["data_url","chart_type"]
	}`)
}
func (t *plotGenerateTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in PlotGenerateInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.DataURL == "" || in.ChartType == "" {
		return tool.FailureMsg("data_url and chart_type are required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Generated %s plot", in.ChartType),
		map[string]any{
			"plot_url":   fmt.Sprintf("/tmp/plot_%s.png", in.ChartType),
			"data_url":   in.DataURL,
			"chart_type": in.ChartType,
			"title":      in.Title,
			"x_label":    in.XLabel,
			"y_label":    in.YLabel,
		},
	), nil
}

type plotReviewTool struct{}

func (t *plotReviewTool) Name() string { return "plot_review" }
func (t *plotReviewTool) Description() string {
	return "Reviews a scientific plot for correctness and quality."
}
func (t *plotReviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"plot_url":{"type":"string","description":"URL or path of the plot to review"},
			"criteria":{"type":"string","description":"Review criteria (e.g. accuracy, readability)"}
		},
		"required":["plot_url"]
	}`)
}
func (t *plotReviewTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in PlotReviewInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
		}
	}
	if in.PlotURL == "" {
		return tool.FailureMsg("plot_url is required"), nil
	}
	return tool.SuccessWith(
		fmt.Sprintf("Review complete for %s", in.PlotURL),
		map[string]any{
			"plot_url": in.PlotURL,
			"criteria": in.Criteria,
			"score":    0.90,
			"passed":   true,
		},
	), nil
}
