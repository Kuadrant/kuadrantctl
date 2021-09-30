package utils

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func ValidateOAS3(docRaw []byte) error {
	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData(docRaw)
	if err != nil {
		return err
	}

	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		return fmt.Errorf("OpenAPI validation error: %w", err)
	}

	return nil
}
