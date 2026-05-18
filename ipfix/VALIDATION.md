# IPFIX Template Validation Guide

## Overview

Template validation is a critical safety feature that helps detect configuration errors, detect when a router has changed its IPFIX export format, and prevent data corruption from desynchronized templates.

The validation system automatically checks templates when they are registered, and also provides programmatic APIs for explicit validation.

## Automatic Template Validation

Templates are automatically validated when registered with the session. To see validation results, enable debug logging:

```bash
export NETFLOWDEBUG=1
./your-app
```

This will output validation results for all received templates:

```
ipfix: register template: id=259 fields=34 (...)
ipfix:   template 259: 34 fields, total size: 152 bytes
ipfix: Template 259 Validation:
ipfix:   Status: ✓ VALID
ipfix:   Record Size: 152 bytes
ipfix:   Field Count: 34
```

If validation detects issues:

```
ipfix: WARNING: Template 259 validation failed:
ipfix: Template 259 Validation:
ipfix:   Status: ✗ INVALID
ipfix:   Record Size: 152 bytes
ipfix:   Field Count: 34
ipfix:   Errors:
ipfix:     ✗ template has no fields
```

## Programmatic Validation

### Basic Validation

```go
import "github.com/xzones2014/netflow/ipfix"

// Validate a specific template
validator := ipfix.NewTemplateValidator(templateID, template)
result := validator.Validate()

if !result.Valid {
    log.Printf("Template %d is invalid: %v\n", templateID, result.Errors)
    return // Don't process data with invalid template
}
```

### Validate from Session

```go
// Validate by template ID from a session
result := ipfix.ValidateTemplateByID(templateID, session)

if !result.Valid {
    log.Printf("Cannot use template %d: %v\n", result.TemplateID, result.Errors)
} else if result.IsSuspicious {
    log.Printf("Warning: Template %d may have issues: %v\n", result.TemplateID, result.Warnings)
}
```

### Inspect Results

```go
result := ipfix.ValidateTemplateByID(templateID, session)

// Print everything
fmt.Println(result.String())

// Or access individual fields
fmt.Printf("Template ID: %d\n", result.TemplateID)
fmt.Printf("Valid: %v\n", result.Valid)
fmt.Printf("Record Size: %d bytes\n", result.RecordSize)
fmt.Printf("Field Count: %d\n", result.FieldCount)

if len(result.Errors) > 0 {
    for _, err := range result.Errors {
        log.Printf("ERROR: %s\n", err)
    }
}

if len(result.Warnings) > 0 {
    for _, warn := range result.Warnings {
        log.Printf("WARNING: %s\n", warn)
    }
}
```

## Validation Checks

The validator performs comprehensive checks:

### 1. Structure Validation
- ✓ Template is not nil
- ✓ Template has at least one field
- ✓ Field count does not exceed 255 (IPFIX limit)

**What to do if it fails:**
- Template definition not properly initialized
- May indicate memory corruption
- Contact support if consistently failing

### 2. Field Type Validation
- ✓ Detects reserved/invalid field types
- ✓ Identifies suspicious type combinations
- ✓ Flags enterprise fields with unusual properties

**What to do if warnings appear:**
- Check your IPFIX configuration
- Verify field types against RFC 5102
- Ensure router firmware is supported

### 3. Field Size Validation
- ✓ Validates against IPFIX RFC 5102 standards
- ✓ Checks for unreasonably large fixed-size fields (>512 bytes)
- ✓ Validates common field types:
  - octetDeltaCount: 8 bytes
  - packetDeltaCount: 8 bytes
  - protocolIdentifier: 1 byte
  - sourceTransportPort: 2 bytes
  - sourceIPv4Address: 4 bytes
  - sourceIPv6Address: 16 bytes
  - And 20+ more RFC 5102 fields

**What to do if validation fails:**
- Field size doesn't match IPFIX specification
- Router may have unusual configuration
- Check router IPFIX settings

### 4. Common Issues Detection
- ✓ Detects very small field counts (<2 fields)
- ✓ Detects MikroTik IPFIX exports (for special handling)
- ✓ Detects missing common required fields
  - Protocol (field 4)
  - Source/Destination ports (fields 7, 11)

**What to do if warnings appear:**
- Review router IPFIX field configuration
- Ensure minimum required fields are enabled
- May indicate recent configuration change

## Interpreting Validation Results

### Valid Template ✓
```
Template 259 Validation:
  Status: ✓ VALID
  Record Size: 152 bytes
  Field Count: 34
```

**Action:** Safe to process data

### Invalid Template ✗
```
Template 259 Validation:
  Status: ✗ INVALID
  Record Size: 0 bytes
  Field Count: 0
  Errors:
    ✗ template has no fields
  ACTION REQUIRED: Template is invalid and will not decode correctly
```

**Action:** Stop processing this template, request new template from router

### Suspicious but Valid Template ⚠️
```
Template 259 Validation:
  Status: ✓ VALID
  Record Size: 152 bytes
  Field Count: 34
  Warnings:
    ⚠️  field 10: unusually large fixed size 512
  NOTE: Template decoded but has suspicious characteristics
```

**Action:** Process data but monitor for issues

## Common Issues and Solutions

### Issue: "template has no fields"
- **Cause:** Router hasn't started exporting data yet
- **Solution:** Wait for IPFIX export to start
- **Workaround:** Check router IPFIX configuration

### Issue: "field count exceeds 255"
- **Cause:** Malformed template or data corruption
- **Solution:** Restart IPFIX collection
- **Workaround:** Check for data line issues

### Issue: "field X: expected Y bytes, got Z"
- **Cause:** Router changed field size (firmware upgrade)
- **Solution:** Reset collector cache and restart
- **Workaround:** Update collector configuration for new router version

### Issue: "unusually large fixed size N"
- **Cause:** Non-standard field size (custom enterprise field)
- **Solution:** This is usually OK, just informational warning
- **Note:** MikroTik sometimes uses custom field sizes

### Issue: "template missing N common fields"
- **Cause:** Router configured to not export certain fields
- **Solution:** Check router IPFIX field configuration
- **Impact:** Some network analysis features may not work

## Best Practices

### 1. Validate on Application Start

```go
// When application starts, validate all known templates
result := ipfix.ValidateTemplateByID(expectedTemplateID, session)
if !result.Valid {
    log.Fatal("Expected template is invalid:", result.Errors)
}
```

### 2. Log Validation Results

```go
result := ipfix.ValidateTemplateByID(templateID, session)
log.Printf("Template %d validation: %s", result.TemplateID, result.String())
```

### 3. Monitor for Changes

```go
// Store initial validation result
initialResult := ipfix.ValidateTemplateByID(templateID, session)

// Later, validate again
newResult := ipfix.ValidateTemplateByID(templateID, session)

// Alert if template changed
if initialResult.RecordSize != newResult.RecordSize {
    log.Printf("ALERT: Template %d record size changed from %d to %d\n",
        templateID, initialResult.RecordSize, newResult.RecordSize)
}
```

### 4. Check Before Processing

```go
result := ipfix.ValidateTemplateByID(templateID, session)

if !result.Valid {
    log.Printf("Skipping invalid template %d\n", templateID)
    return
}

// Safe to process
processTemplateData(result.RecordSize)
```

### 5. Build a Template Inventory

```go
// Track all templates and their status
templateStatus := make(map[uint16]*ipfix.ValidationResult)

result := ipfix.ValidateTemplateByID(templateID, session)
templateStatus[templateID] = result

// Report status
for templateID, result := range templateStatus {
    fmt.Printf("%d: %s\n", templateID, map[bool]string{
        true: "✓ VALID",
        false: "✗ INVALID",
    }[result.Valid])
}
```

## Performance Considerations

- Template validation runs once when template is registered
- Minimal overhead (O(n) where n = number of fields)
- Average template has 20-40 fields = microsecond-scale validation
- Does not impact data record processing performance

## Integration with Desynchronization Detection

Template validation works together with desynchronization detection:

1. **Template registered:** Validated automatically
2. **Data records received:** Compared against template size
3. **Mismatch detected:** Diagnostic information logged
4. **Validation + Diagnostics:** Together identify root cause

Example:

```go
// Template validation shows record size is 152 bytes
// Desynchronization detection shows actual records are 124 bytes
// Together they indicate template mismatch

// Solution: Reset collector cache
```

## References

- [IPFIX Desynchronization Guide](./DESYNCHRONIZATION.md)
- RFC 7011: IPFIX Protocol Specification
- RFC 5102: IPFIX Information Elements
- RFC 5103: Bidirectional Flow Export Using IPFIX
- MikroTik IPFIX: https://wiki.mikrotik.com/wiki/Manual:IP/Traffic_Flow

## Support

For validation issues:
1. Enable `NETFLOWDEBUG=1` to see validation output
2. Collect validation result with `result.String()`
3. Check against RFC 5102 field specifications
4. Review router IPFIX configuration
5. Open issue with validation output attached
