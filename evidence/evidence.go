package evidence

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/in-toto/go-witness/cryptoutil"
)

// RawEvidence represents a simplified raw output from a policy engine.
type RawEvidence struct {
	Metadata `json:,inline`
	Details  json.RawMessage `json:"details"`
	Resource Resource        `json:"resource"`
}

type Metadata struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	PolicyID  string    `json:"policyId"`
	Decision  string    `json:"decision"`
}

type Resource struct {
	Name   string               `json:"name"`
	Digest cryptoutil.DigestSet `json:"digest"`
}

// Export simulates exporting evidence to a backend.
// In a real scenario, this would be client.PutObject(rawEvJSON, rawEvidenceRef) to object storage.
func Export(rawEvidenceRef string, rawEvJSON []byte) error {
	fmt.Printf("\n--- Pushing Raw Evidence to Data Lake (%s) ---\n%s\n", rawEvidenceRef, string(rawEvJSON))
	return nil
}
