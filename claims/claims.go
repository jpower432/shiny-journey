package claims

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/revanite-io/sci/layer4"

	"github.com/jpower432/shiny-journey/evidence"
)

// ConformanceClaim represents a higher-level, mapped conformance assertion.
type ConformanceClaim struct {
	ClaimID        string            `json:"claimId"`
	Timestamp      time.Time         `json:"timestamp"`
	ResourceRef    string            `json:"resourceRef"`
	RawEvidenceRef string            `json:"rawEvidenceRef"`
	Summary        string            `json:"summary"`
	Assessment     layer4.Assessment `json:"assessment"`
}

func (c *ConformanceClaim) MarshalJSON() ([]byte, error) {
	outputMap := make(map[string]interface{})
	outputMap["clamId"] = c.ClaimID
	outputMap["timestamp"] = c.Timestamp
	outputMap["resourceRef"] = c.ResourceRef
	outputMap["summary"] = c.Summary
	assessment := make(map[string]interface{})
	assessment["requirement_id"] = c.Assessment.RequirementID

	methods := []map[string]interface{}{}
	for _, method := range c.Assessment.Methods {
		methodMap := make(map[string]interface{})
		methodMap["name"] = method.Name
		methodMap["description"] = method.Description
		methodMap["run"] = method.Run
		if method.Result != nil {
			methodMap["result"] = map[string]interface{}{
				"status": method.Result.Status,
			}
		}
		methods = append(methods, methodMap)
	}
	assessment["methods"] = methods
	outputMap["assessment"] = assessment
	return json.Marshal(outputMap)
}

// PopulateAssessment simulates evaluations of evidence against policies.
func (c *ConformanceClaim) PopulateAssessment(rawEv evidence.RawEvidence) {
	summary := fmt.Sprintf("Resource '%s' from %s is %s against policy '%s'.",
		rawEv.Resource, rawEv.Source, rawEv.Decision, rawEv.PolicyID)
	c.Summary = summary
	c.Assessment = simulatedCheck(rawEv)
}

// TODO: Update this to accept an plan and provider for assessment generation.
// simulateMapping calls a mapping type depending on evidence type.
// In a real scenario, this would be delegated to a provider.
func simulatedCheck(rawEv evidence.RawEvidence) layer4.Assessment {
	assessment := layer4.Assessment{
		RequirementID: "placeholder",
	}
	methodMapper, ok := sourceToMethod[rawEv.Source]
	if !ok {
		panic(fmt.Sprintf("Wrong source %s", rawEv.Source))
	}
	method := methodMapper(rawEv)
	assessment.Methods = append(assessment.Methods, method)
	return assessment
}

type methodMapperFunc func(rawEv evidence.RawEvidence) layer4.AssessmentMethod

var sourceToMethod = map[string]methodMapperFunc{
	"OPA": func(rawEv evidence.RawEvidence) layer4.AssessmentMethod {
		method := layer4.AssessmentMethod{
			Name:   "OPA",
			Run:    true,
			Result: &layer4.AssessmentResult{},
		}
		if rawEv.Decision == "deny" {
			method.Result.Status = "NOT_COMPLIANT"
			method.Description = fmt.Sprintf("OPA denied access to resource '%s' due to policy '%s' violation. %s", rawEv.Resource, rawEv.PolicyID, string(rawEv.Details))
		} else if rawEv.Decision == "allow" {
			method.Result.Status = "COMPLIANT"
			method.Description = fmt.Sprintf("OPA allowed access to resource '%s' adhering to policy '%s'.", rawEv.Resource, rawEv.PolicyID)
		}
		return method
	},
	"Kyverno": func(rawEv evidence.RawEvidence) layer4.AssessmentMethod {
		method := layer4.AssessmentMethod{
			Name:   "Kyverno",
			Run:    true,
			Result: &layer4.AssessmentResult{},
		}
		// Assume Kyverno 'mutate' implies compliance enforcement
		if rawEv.Decision == "mutate" {
			method.Result.Status = "COMPLIANT"
			method.Description = fmt.Sprintf("Kyverno mutated resource '%s' to enforce policy '%s'.", rawEv.Resource, rawEv.PolicyID)
		} else if rawEv.Decision == "deny" {
			method.Result.Status = "NON_COMPLIANT"
			method.Description = fmt.Sprintf("Kyverno denied resource '%s' due to policy '%s' violation.", rawEv.Resource, rawEv.PolicyID)
		}
		return method
	},
	"OpenSCAP": func(rawEv evidence.RawEvidence) layer4.AssessmentMethod {
		method := layer4.AssessmentMethod{
			Name:   "OpenSCAP",
			Run:    true,
			Result: &layer4.AssessmentResult{},
		}
		if rawEv.Decision == "compliant" {
			method.Result.Status = "COMPLIANT"
			method.Description = fmt.Sprintf("OpenSCAP scan for '%s' reported compliant against profile '%s'.", rawEv.Resource, rawEv.PolicyID)
		} else if rawEv.Decision == "non_compliant" {
			method.Result.Status = "NON_COMPLIANT"
			method.Description = fmt.Sprintf("OpenSCAP scan for '%s' reported non-compliant against profile '%s'. Details: %s", rawEv.Resource, rawEv.PolicyID, string(rawEv.Details))
		}
		return method
	},
}
