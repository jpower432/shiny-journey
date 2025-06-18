package metrics

// For metrics catalogs?

// Metric represents a single metric entry in a catalog.
type Metric struct {
	ID                string
	PrimaryControlID  string
	RelatedControlIDs []string
	MetricDescription string
	Expression        Expression
	Rules             string
	// SLO/for alerting
	Threshold string
}

// Expression defines the formula and its parameters for a metric.
type Expression struct {
	// Could this be type?
	Formula    string      `yaml:"formula"`
	Parameters []Parameter `yaml:"parameters"`
}

// Parameter represents a single parameter used in a metrics expression.
type Parameter struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}
