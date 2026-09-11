# scientific_plotting

`mcp.scientific_plotting` — plot generation and review.

## Purpose

Generate publication-quality plots from tabular data and review them.

## Tools

- `plot_generate` — produce a plot from input data
- `plot_review` — review a generated plot for issues

## Example

```bash
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.scientific_plotting \
  -H 'Content-Type: application/json' \
  -d '{"tool":"plot_generate","input":{"csv":"data.csv","type":"histogram"}}'
```
