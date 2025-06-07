package claims

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	gowitness "github.com/in-toto/go-witness"
	"github.com/in-toto/go-witness/archivista"
	"github.com/in-toto/go-witness/attestation"
	"github.com/in-toto/go-witness/cryptoutil"
	"github.com/invopop/jsonschema"

	"github.com/jpower432/shiny-journey/evidence"
)

const Name = "conformance-claim"
const Type = "https://example.com/conformance-claim/v1"
const RunType = attestation.VerifyRunType

type AssessmentAttestor struct {
	Claim       ConformanceClaim
	rawEvidence evidence.RawEvidence
	evidenceRef string
}

func NewAttestor(evidence evidence.RawEvidence) *AssessmentAttestor {
	return &AssessmentAttestor{
		Claim:       ConformanceClaim{},
		rawEvidence: evidence,
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
	claimID := uuid.New().String()
	claim := ConformanceClaim{
		ClaimID:        claimID,
		Timestamp:      time.Now(),
		ResourceRef:    a.rawEvidence.Resource.Name,
		RawEvidenceRef: a.evidenceRef,
	}
	claim.PopulateAssessment(a.rawEvidence)
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
	if err := json.Unmarshal(data, &a.Claim); err != nil {
		return err
	}

	return nil
}

// Export exports attestations to remote storage
func Export(ctx context.Context, attestor attestation.Attestor, signer cryptoutil.Signer, archivistaURL string) error {
	opts := []gowitness.RunOption{
		gowitness.RunWithAttestors([]attestation.Attestor{attestor}),
	}

	if signer == nil {
		opts = append(opts, gowitness.RunWithInsecure(true))
	} else {
		opts = append(opts, gowitness.RunWithSigners(signer))
	}

	runResults, err := gowitness.RunWithExports("comply", opts...)
	if err != nil {
		return err
	}

	client := archivista.New(archivistaURL)
	for _, result := range runResults {
		atts := result.Collection.Attestations
		if len(atts) == 0 {
			continue
		}
		gitoid, err := client.Store(ctx, result.SignedEnvelope)
		if err != nil {
			return err
		}
		log.Printf("gitoid: %s", gitoid)
	}
	return nil
}
