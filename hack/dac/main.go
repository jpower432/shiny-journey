package main

import (
	"flag"
	"time"

	"github.com/perses/perses/go-sdk"
	"github.com/perses/perses/go-sdk/common"
	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listVar "github.com/perses/perses/go-sdk/variable/list-variable"
	txtVar "github.com/perses/perses/go-sdk/variable/text-variable"
	gaugePanel "github.com/perses/plugins/gaugechart/sdk/go"
	markdownPanel "github.com/perses/plugins/markdown/sdk/go"
	"github.com/perses/plugins/prometheus/sdk/go/query"
	labelValuesVar "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	statPanel "github.com/perses/plugins/statchart/sdk/go"
	tablePanel "github.com/perses/plugins/table/sdk/go"
	timeSeriesPanel "github.com/perses/plugins/timeserieschart/sdk/go"
)

func main() {
	flag.Parse()
	exec := sdk.NewExec()
	prometheusDatasourceName := "prometheuslocal"

	var columnSettings = []tablePanel.ColumnSettings{
		{
			Name:   "resource",
			Header: "Resource",
			Align:  tablePanel.LeftAlign,
			Format: common.Format{
				Unit: common.BytesUnit,
			},
		},
		{
			Name:   "attestation_id",
			Header: "Attestation ID",
			Align:  tablePanel.LeftAlign,
			Format: common.Format{
				Unit: common.BytesUnit,
			},
		},
		{
			Name:   "value",
			Header: "Compliance Status",
			Align:  tablePanel.LeftAlign,
			Format: common.Format{
				Unit: common.DecimalUnit,
			},
		},
		{
			Name:   "baseline_id",
			Header: "Baseline ID",
			Align:  tablePanel.LeftAlign,
			Format: common.Format{
				Unit: common.BytesUnit,
			},
		},
		{
			Name:   "requirement_id",
			Header: "Requirement ID",
			Align:  tablePanel.LeftAlign,
			Format: common.Format{
				Unit: common.BytesUnit,
			},
		},
		{
			Name: "timestamp",
			Format: common.Format{
				Unit: common.BytesUnit,
			},
			Hide: true,
		},
		{
			Name: "metric_type",
			Format: common.Format{
				Unit: common.BytesUnit,
			},
			Hide: true,
		},
		{
			Name: "__name__",
			Format: common.Format{
				Unit: common.BytesUnit,
			},
			Hide: true,
		},
	}

	cellSettings := []tablePanel.CellSettings{
		{
			Condition: tablePanel.Condition{
				Kind: tablePanel.ValueConditionKind,
				Spec: tablePanel.ValueConditionSpec{
					Value: "1",
				},
			},
			Text:            "COMPLIANT",
			BackgroundColor: "#00FF00",
		},
		{
			Condition: tablePanel.Condition{
				Kind: tablePanel.ValueConditionKind,
				Spec: tablePanel.ValueConditionSpec{
					Value: "0",
				},
			},
			Text:            "NON_COMPLIANT",
			BackgroundColor: "#FF0000",
		},
	}

	// Create the dashboard definition
	builder, buildErr := dashboard.New("ContinuousMonitoring",
		dashboard.ProjectName("ShinyJourney"),
		dashboard.RefreshInterval(1*time.Minute),

		dashboard.AddVariable("job",
			txtVar.Text("agent", txtVar.Constant(true)),
		),

		dashboard.AddVariable(
			"evidenceResource",
			listVar.List(
				labelValuesVar.PrometheusLabelValues(
					prometheusDatasourceName,
					labelValuesVar.LabelName("evidence_resource"),
					labelValuesVar.Matchers(`{__name__="evidence_processed_total"}`),
				),
				listVar.DisplayName("Assessment Subject"),
				listVar.AllowMultiple(true),
				listVar.AllowAllValue(true),
			),
		),

		dashboard.AddVariable(
			"evidenceSource",
			listVar.List(
				labelValuesVar.PrometheusLabelValues(
					prometheusDatasourceName,
					labelValuesVar.LabelName("evidence_source"),
					labelValuesVar.Matchers(`{__name__="evidence_processed_total"}`),
				),
				listVar.DisplayName("Policy Source"),
				listVar.AllowAllValue(true),
				listVar.AllowMultiple(true),
			),
		),

		dashboard.AddVariable(
			"catalogId",
			listVar.List(
				labelValuesVar.PrometheusLabelValues(
					prometheusDatasourceName,
					labelValuesVar.LabelName("baseline_id"),
					labelValuesVar.Matchers(`{__name__="compliance_assessment_status"}`),
				),
				listVar.DisplayName("Baseline ID"),
				listVar.AllowAllValue(true),
				listVar.AllowMultiple(true),
			),
		),

		dashboard.AddVariable(
			"requirementId",
			listVar.List(
				labelValuesVar.PrometheusLabelValues(
					prometheusDatasourceName,
					labelValuesVar.LabelName("requirement_id"),
					labelValuesVar.Matchers(`{__name__="compliance_assessment_status"}`),
				),
				listVar.DisplayName("Requirement ID"),
				listVar.AllowAllValue(true),
				listVar.AllowMultiple(true),
			),
		),

		dashboard.AddVariable(
			"attestationID", // New variable for attestation_id
			listVar.List(
				labelValuesVar.PrometheusLabelValues(
					prometheusDatasourceName,
					labelValuesVar.LabelName("attestation_id"),
					labelValuesVar.Matchers(`{__name__="compliance_assessment_status"}`),
				),
				listVar.DisplayName("Attestation ID"),
				listVar.AllowAllValue(true),
				listVar.AllowMultiple(true),
			),
		),

		dashboard.AddPanelGroup("Overview",
			panelgroup.PanelsPerLine(2),
			panelgroup.AddPanel("Baseline Compliance Percentage",
				gaugePanel.Chart(
					gaugePanel.Format(common.Format{
						Unit: string(common.PercentUnit),
					}),
				),
				panel.AddQuery(
					query.PromQL(
						`avg by (baseline_id) (baseline_compliance_percentage{resource=~"$evidenceResource"})`,
					),
				),
				panel.Description("Overall percentage of compliant assessments within selected baseline(s) and resource(s)."),
			),
			panelgroup.AddPanel("Team Documentation",
				markdownPanel.Markdown("**This panel could probably contain some information about how to read the dashboard.**"),
			),
		),

		dashboard.AddPanelGroup("Detailed Compliance Status",
			panelgroup.PanelsPerLine(1),
			// Table of Requirements and their Compliance Status
			panelgroup.AddPanel("Requirements Status Table",
				tablePanel.Table(
					tablePanel.WithColumnSettings(columnSettings),
					tablePanel.WithCellSettings(cellSettings),
				),
				panel.AddLink("http://localhost:8082/"),
				panel.AddQuery(
					query.PromQL(
						`requirement_binary_compliance_status{resource=~"$evidenceResource", requirement_id=~"$requirementId", baseline_id=~"$catalogId"}`,
					),
				),
				panel.Description("Detailed status of each requirement, filtered by selected dimensions."),
			),
		),

		dashboard.AddPanelGroup("Evidence Stats",
			panelgroup.PanelsPerLine(2),
			panelgroup.AddPanel("Total Non-Compliant Assessments (Raw)",
				statPanel.Chart(),
				panel.AddQuery(
					query.PromQL(
						`sum(non_compliant_requirements_count{resource=~"$evidenceResource"})`,
					),
				),
			),
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
