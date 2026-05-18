// Example: Template Validation
//
// This example demonstrates how to use the template validation API
// to verify IPFIX templates before processing data records.
//
// Usage:
//   export NETFLOWDEBUG=1  # Optional: enable debug logging
//   go run example_validation.go

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xzones2014/netflow/ipfix"
	"github.com/xzones2014/netflow/session"
)

func main() {
	// Example 1: Direct template validation
	exampleDirectValidation()

	// Example 2: Validate template from session
	exampleSessionValidation()

	// Example 3: Check validation results
	exampleCheckResults()
}

// exampleDirectValidation shows how to validate a template directly
func exampleDirectValidation() {
	fmt.Println("=== Example 1: Direct Template Validation ===\n")

	// In real code, you would get this template from the decoder
	// For this example, we'll show the validation API usage pattern

	// Create a validator for a specific template
	// templateID := uint16(259)
	// template := ... // obtained from session
	// validator := ipfix.NewTemplateValidator(templateID, template)
	// result := validator.Validate()

	fmt.Println("Pattern:")
	fmt.Println(`
  validator := ipfix.NewTemplateValidator(templateID, template)
  result := validator.Validate()
  
  if !result.Valid {
      log.Println("Template validation failed!")
      log.Println(result.String())
  }
`)
	fmt.Println()
}

// exampleSessionValidation shows how to validate a template by ID from session
func exampleSessionValidation() {
	fmt.Println("=== Example 2: Validate Template from Session ===\n")

	// In real code:
	// sess := ... // your session
	// templateID := uint16(259)
	// result := ipfix.ValidateTemplateByID(templateID, sess)

	fmt.Println("Pattern:")
	fmt.Println(`
  result := ipfix.ValidateTemplateByID(templateID, session)
  
  fmt.Println("Template ID:", result.TemplateID)
  fmt.Println("Valid:", result.Valid)
  fmt.Println("Record Size:", result.RecordSize, "bytes")
  fmt.Println("Field Count:", result.FieldCount)
  
  if len(result.Errors) > 0 {
      fmt.Println("Errors:", result.Errors)
  }
  if len(result.Warnings) > 0 {
      fmt.Println("Warnings:", result.Warnings)
  }
`)
	fmt.Println()
}

// exampleCheckResults shows how to interpret validation results
func exampleCheckResults() {
	fmt.Println("=== Example 3: Check Validation Results ===\n")

	fmt.Println("Validation Results have these fields:")
	fmt.Println()
	fmt.Println("  TemplateID   uint16   - ID of the template")
	fmt.Println("  Valid        bool     - Passed all validation checks")
	fmt.Println("  RecordSize   int      - Size of each data record in bytes")
	fmt.Println("  FieldCount   int      - Number of fields in template")
	fmt.Println("  Errors       []string - Critical validation failures")
	fmt.Println("  Warnings     []string - Non-critical issues")
	fmt.Println("  IsSuspicious bool     - True if warnings present")
	fmt.Println()

	fmt.Println("Use them like this:")
	fmt.Println(`
  result := ipfix.ValidateTemplateByID(templateID, sess)
  
  // Check if template is valid for processing
  if !result.Valid {
      log.Printf("Cannot process template %d: %v\n", result.TemplateID, result.Errors)
      return
  }
  
  // Warn about suspicious templates
  if result.IsSuspicious {
      log.Printf("Warning: Template %d has issues: %v\n", result.TemplateID, result.Warnings)
  }
  
  // Use record size for validation
  log.Printf("Template %d expects %d-byte records\n", result.TemplateID, result.RecordSize)
`)
	fmt.Println()
}
