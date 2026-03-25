package schema

// Pipeline represents a Spinnaker pipeline definition.
type Pipeline struct {
	Name   string  `json:"name"`
	Stages []Stage `json:"stages"`
}

// Stage represents a single stage within a Spinnaker pipeline.
type Stage struct {
	Type                 string   `json:"type"`
	Name                 string   `json:"name"`
	RefID                string   `json:"refId"`
	RequisiteStageRefIds []string `json:"requisiteStageRefIds"`
}
