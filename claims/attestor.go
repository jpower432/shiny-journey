package claims

import (
	"context"
	"time"

	"github.com/google/uuid"
	gowitness "github.com/in-toto/go-witness"
	"github.com/in-toto/go-witness/archivista"
	"github.com/in-toto/go-witness/attestation"
	"github.com/invopop/jsonschema"
)

const Name = "conformance-claim"
const Type = "https://example.com/conformance-claim/v1"
const RunType = attestation.VerifyRunType

type AssessmentAttestor struct {
	Claim       ConformanceClaim
	rawEvidence RawEvidence
	evidenceRef string
}

func NewAttestor(evidence RawEvidence) *AssessmentAttestor {
	return &AssessmentAttestor{
		Claim:       ConformanceClaim{},
		rawEvidence: evidence,
	}
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
		ResourceRef:    a.rawEvidence.Resource,
		RawEvidenceRef: a.evidenceRef,
	}
	claim.PopulateAssessment(a.rawEvidence)
	return nil
}

func (a *AssessmentAttestor) Schema() *jsonschema.Schema {
	return jsonschema.Reflect(a.Claim)
}

// Export exports attestations to remote storage
func Export(ctx context.Context, attestor attestation.Attestor, archivistaURL string) error {
	// using this purposefully for not because I just want the one envelope for now
	runResults, err := gowitness.Run("comply",
		gowitness.RunWithAttestors([]attestation.Attestor{attestor}),
		gowitness.RunWithInsecure(true),
	)
	if err != nil {
		return err
	}

	// export attestations to Archivista
	client := archivista.New(archivistaURL)
	_, err = client.Store(ctx, runResults.SignedEnvelope)
	if err != nil {
		return err
	}
	return nil
}
