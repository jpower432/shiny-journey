package main

import (
	"flag"
	"time"

	"github.com/perses/perses/go-sdk"
	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listVar "github.com/perses/perses/go-sdk/variable/list-variable"
	txtVar "github.com/perses/perses/go-sdk/variable/text-variable"
	markdownPanel "github.com/perses/plugins/markdown/sdk/go"
	"github.com/perses/plugins/prometheus/sdk/go/query"
	labelValuesVar "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	timeSeriesPanel "github.com/perses/plugins/timeserieschart/sdk/go"
)

func main() {
	flag.Parse()
	exec := sdk.NewExec()
	prometheusDatasourceName := "prometheuslocal"
	// Create the dashboard definition
	builder, buildErr := dashboard.New("ContinuousMonitoring",
		dashboard.ProjectName("ShinyJourney"),
		dashboard.RefreshInterval(1*time.Minute),

		dashboard.AddVariable("job",
			txtVar.Text("agent", txtVar.Constant(true)),
		),

		dashboard.AddVariable(
			"evidenceResource", // Variable name (e.g., used as $evidenceResource in queries)
			listVar.List(
				labelValuesVar.PrometheusLabelValues(
					prometheusDatasourceName,
					// This is the key part: get values for the 'evidence_resource' label
					// filtered to the 'evidence.processed_total' metric
					labelValuesVar.LabelName("evidence_resource"),
					labelValuesVar.Matchers(`{__name__="evidence_processed_total"}`),
				),
				listVar.DisplayName("Evidence Resource"),
				listVar.AllowAllValue(true),
			),
		),

		// Define another variable to list all unique 'evidence_source' values
		dashboard.AddVariable(
			"evidenceSource",
			listVar.List(
				labelValuesVar.PrometheusLabelValues(
					prometheusDatasourceName,
					labelValuesVar.LabelName("evidence_source"),
					labelValuesVar.Matchers(`{__name__="evidence_processed_total"}`),
				),
				listVar.DisplayName("Evidence Source"),
				listVar.AllowAllValue(true),
				listVar.AllowMultiple(true),
			),
		),

		dashboard.AddPanelGroup("Directions",
			panelgroup.PanelsPerLine(1),
			// PANELS
			panelgroup.AddPanel("Reading the Dashboard",
				markdownPanel.Markdown("**This panel could probably contain some information about how to read the dashboard.**"),
			),
		),

		dashboard.AddPanelGroup("Evidence",
			panelgroup.PanelsPerLine(3),
			// PANELS
			panelgroup.AddPanel("Evidence Processed Rate",
				timeSeriesPanel.Chart(),
				panel.AddQuery(
					query.PromQL(
						`rate(evidence_processed_total{evidence_resource=~"$evidenceResource", evidence_source=~"$evidenceSource", job="agent"}[5m])`,
					),
				),
			),
		),
	)
	exec.BuildDashboard(builder, buildErr)
}
