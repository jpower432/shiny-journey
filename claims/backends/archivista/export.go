package archivista

import (
	"context"
	"log"

	gowitness "github.com/in-toto/go-witness"
	"github.com/in-toto/go-witness/archivista"
	"github.com/in-toto/go-witness/attestation"
	"github.com/in-toto/go-witness/cryptoutil"
)

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
