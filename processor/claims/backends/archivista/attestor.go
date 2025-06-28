package archivista

import (
	"encoding/json"
	"fmt"

	"github.com/in-toto/go-witness/attestation"
	"github.com/in-toto/go-witness/cryptoutil"
	"github.com/invopop/jsonschema"
	"github.com/revanite-io/sci/layer4"

	"github.com/jpower432/shiny-journey/processor/claims"
	"github.com/jpower432/shiny-journey/processor/claims/evidence"
)

const Name = "conformance-claim"
const Type = "https://example.com/conformance-claim/v1"
const RunType = attestation.VerifyRunType

type AssessmentAttestor struct {
	Claim       *claims.ConformanceClaim
	rawEvidence evidence.RawEvidence
	plan        layer4.Layer4
	evidenceRef string
}

func NewAttestor(evidence evidence.RawEvidence, rawEnvRef string, plan layer4.Layer4) *AssessmentAttestor {
	return &AssessmentAttestor{
		plan:        plan,
		rawEvidence: evidence,
		evidenceRef: rawEnvRef,
	}
}

func (a *AssessmentAttestor) Export() bool {
	return true
}

func (a *AssessmentAttestor) Name() string {
	return Name
}

func (a *AssessmentAttestor) Type() string {
	return Type
}

func (a *AssessmentAttestor) RunType() attestation.RunType {
	return RunType
}

func (a *AssessmentAttestor) Attest(ctx *attestation.AttestationContext) error {
	claim := claims.NewFromEvidence(a.rawEvidence, a.evidenceRef, a.plan)
	a.Claim = claim
	return nil
}

func (a *AssessmentAttestor) Subjects() map[string]cryptoutil.DigestSet {
	subj := make(map[string]cryptoutil.DigestSet)
	subj[fmt.Sprintf("resource:%s", a.rawEvidence.Resource.Name)] = a.rawEvidence.Resource.Digest
	return subj
}

func (a *AssessmentAttestor) Schema() *jsonschema.Schema {
	return jsonschema.Reflect(a.Claim)
}

func (a *AssessmentAttestor) MarshalJSON() ([]byte, error) {
	return a.Claim.MarshalJSON()
}

func (a *AssessmentAttestor) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, a.Claim); err != nil {
		return err
	}

	return nil
}
