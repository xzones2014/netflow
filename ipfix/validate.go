package ipfix

import (
	"fmt"
	"strings"

	"github.com/xzones2014/netflow/session"
)

// TemplateValidator provides comprehensive validation of IPFIX templates.
type TemplateValidator struct {
	templateID uint16
	template   session.Template
	errors     []string
	warnings   []string
}

// ValidationResult contains the results of template validation.
type ValidationResult struct {
	TemplateID   uint16
	Valid        bool
	RecordSize   int
	FieldCount   int
	Errors       []string
	Warnings     []string
	IsSuspicious bool
}

// NewTemplateValidator creates a new validator for the given template.
func NewTemplateValidator(templateID uint16, template session.Template) *TemplateValidator {
	return &TemplateValidator{
		templateID: templateID,
		template:   template,
		errors:     []string{},
		warnings:   []string{},
	}
}

// Validate performs comprehensive validation of the template and returns results.
func (tv *TemplateValidator) Validate() *ValidationResult {
	tv.validateBasicStructure()
	tv.validateFieldTypes()
	tv.validateFieldSizes()
	tv.validateForCommonIssues()

	result := &ValidationResult{
		TemplateID:   tv.templateID,
		Valid:        len(tv.errors) == 0,
		RecordSize:   ValidateTemplateSize(tv.template),
		FieldCount:   len(tv.template.GetFields()),
		Errors:       tv.errors,
		Warnings:     tv.warnings,
		IsSuspicious: len(tv.warnings) > 0,
	}

	return result
}

// validateBasicStructure checks basic template structure.
func (tv *TemplateValidator) validateBasicStructure() {
	if tv.template == nil {
		tv.errors = append(tv.errors, "template is nil")
		return
	}

	fields := tv.template.GetFields()
	if len(fields) == 0 {
		tv.errors = append(tv.errors, "template has no fields")
		return
	}

	if len(fields) > 255 {
		tv.errors = append(tv.errors, fmt.Sprintf("template has %d fields, but maximum is 255", len(fields)))
	}
}

// validateFieldTypes checks for valid field types and configurations.
func (tv *TemplateValidator) validateFieldTypes() {
	fields := tv.template.GetFields()

	for i, field := range fields {
		fieldType := field.GetType()
		fieldLen := field.GetLength()

		// Check for reserved field types (0-1, 65535 - variable length indicator)
		if fieldType == 0 || fieldType == 1 {
			tv.warnings = append(tv.warnings, fmt.Sprintf("field %d: reserved type %d", i, fieldType))
		}

		// Check for suspicious combinations
		if fieldType > 300 && fieldType < 32768 {
			// Check if it looks like an enterprise field
			if fieldLen == 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (type %d): zero-length field", i, fieldType))
			}
		}
	}
}

// validateFieldSizes checks field sizes against IPFIX standards.
func (tv *TemplateValidator) validateFieldSizes() {
	fields := tv.template.GetFields()
	totalSize := 0

	for i, field := range fields {
		fieldLen := field.GetLength()
		fieldType := field.GetType()

		// Track total size
		if fieldLen != VariableLength {
			totalSize += int(fieldLen)
		}

		// Check for unreasonably large fixed-size fields
		if fieldLen > 0 && fieldLen < VariableLength && fieldLen > 512 {
			tv.warnings = append(tv.warnings, fmt.Sprintf("field %d: unusually large fixed size %d", i, fieldLen))
		}

		// Common field type checks
		switch fieldType {
		case 1: // octetDeltaCount
			if fieldLen != 8 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (octetDeltaCount): expected 8 bytes, got %d", i, fieldLen))
			}
		case 2: // packetDeltaCount
			if fieldLen != 8 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (packetDeltaCount): expected 8 bytes, got %d", i, fieldLen))
			}
		case 4: // protocolIdentifier
			if fieldLen != 1 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (protocolIdentifier): expected 1 byte, got %d", i, fieldLen))
			}
		case 5: // ipClassOfService
			if fieldLen != 1 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (ipClassOfService): expected 1 byte, got %d", i, fieldLen))
			}
		case 7: // sourceTransportPort
			if fieldLen != 2 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (sourceTransportPort): expected 2 bytes, got %d", i, fieldLen))
			}
		case 8: // sourceIPv4Address
			if fieldLen != 4 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (sourceIPv4Address): expected 4 bytes, got %d", i, fieldLen))
			}
		case 11: // destinationTransportPort
			if fieldLen != 2 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (destinationTransportPort): expected 2 bytes, got %d", i, fieldLen))
			}
		case 12: // destinationIPv4Address
			if fieldLen != 4 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (destinationIPv4Address): expected 4 bytes, got %d", i, fieldLen))
			}
		case 27: // sourceIPv6Address
			if fieldLen != 16 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (sourceIPv6Address): expected 16 bytes, got %d", i, fieldLen))
			}
		case 28: // destinationIPv6Address
			if fieldLen != 16 && fieldLen != 0 {
				tv.warnings = append(tv.warnings, fmt.Sprintf("field %d (destinationIPv6Address): expected 16 bytes, got %d", i, fieldLen))
			}
		}
	}

	if totalSize == 0 && len(fields) > 0 {
		tv.errors = append(tv.errors, "template has fields but total size is 0 (all fields are variable-length)")
	}

	if totalSize > 65535 {
		tv.errors = append(tv.errors, fmt.Sprintf("template total size %d exceeds maximum record size", totalSize))
	}
}

// validateForCommonIssues checks for known problematic configurations.
func (tv *TemplateValidator) validateForCommonIssues() {
	fields := tv.template.GetFields()

	if len(fields) < 2 {
		tv.warnings = append(tv.warnings, "template has very few fields (< 2)")
	}

	// Check for likely MikroTik template (field 22, 21, 160 combination)
	hasMikrotikMarkers := false
	for _, field := range fields {
		fieldType := field.GetType()
		if fieldType == 22 || fieldType == 21 || fieldType == 160 {
			hasMikrotikMarkers = true
			break
		}
	}

	if hasMikrotikMarkers && len(fields) > 20 {
		if debug {
			debugLog.Printf("template %d: appears to be MikroTik IPFIX export\n", tv.templateID)
		}
	}

	// Check for missing common required fields
	fieldTypes := make(map[uint16]bool)
	for _, field := range fields {
		fieldTypes[field.GetType()] = true
	}

	// These are commonly expected in NetFlow
	commonFields := []uint16{4, 7, 11} // protocol, srcPort, dstPort
	missingCommon := 0
	for _, fieldType := range commonFields {
		if !fieldTypes[fieldType] {
			missingCommon++
		}
	}

	if missingCommon > 1 {
		tv.warnings = append(tv.warnings, fmt.Sprintf("template missing %d common fields (protocol, ports)", missingCommon))
	}
}

// String returns a human-readable representation of the validation result.
func (vr *ValidationResult) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Template %d Validation:\n", vr.TemplateID))
	sb.WriteString(fmt.Sprintf("  Status: "))
	if vr.Valid {
		sb.WriteString("✓ VALID\n")
	} else {
		sb.WriteString("✗ INVALID\n")
	}

	sb.WriteString(fmt.Sprintf("  Record Size: %d bytes\n", vr.RecordSize))
	sb.WriteString(fmt.Sprintf("  Field Count: %d\n", vr.FieldCount))

	if len(vr.Errors) > 0 {
		sb.WriteString("  Errors:\n")
		for _, err := range vr.Errors {
			sb.WriteString(fmt.Sprintf("    ✗ %s\n", err))
		}
	}

	if len(vr.Warnings) > 0 {
		sb.WriteString("  Warnings:\n")
		for _, warn := range vr.Warnings {
			sb.WriteString(fmt.Sprintf("    ⚠️  %s\n", warn))
		}
	}

	if !vr.Valid {
		sb.WriteString("  ACTION REQUIRED: Template is invalid and will not decode correctly\n")
	} else if vr.IsSuspicious {
		sb.WriteString("  NOTE: Template decoded but has suspicious characteristics\n")
	}

	return sb.String()
}

// ValidateTemplateByID validates a template retrieved from a session.
func ValidateTemplateByID(templateID uint16, sess session.Session) *ValidationResult {
	if sess == nil {
		return &ValidationResult{
			TemplateID: templateID,
			Valid:      false,
			Errors:     []string{"session is nil"},
		}
	}

	template, found := sess.GetTemplate(templateID)
	if !found {
		return &ValidationResult{
			TemplateID: templateID,
			Valid:      false,
			Errors:     []string{"template not found in session"},
		}
	}

	validator := NewTemplateValidator(templateID, template)
	return validator.Validate()
}
