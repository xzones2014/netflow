# IPFIX Template Desynchronization Detection & Fix

## Problem Summary

Template desynchronization occurs when the IPFIX collector's cached template definition doesn't match the actual data records being sent by the exporter (e.g., MikroTik router).

### What Happens

1. **Template Mismatch**: The template specifies N bytes per record, but actual records are M bytes (N ≠ M)
2. **Cascading Misalignment**: Each record is parsed starting at the wrong offset
3. **Data Corruption**: Fields shift into incorrect positions, causing:
   - Impossible timestamps
   - MAC addresses appearing in wrong fields
   - IP addresses split across fields
   - All downstream records corrupted

### Root Causes

- **Missed Template Packet**: UDP packet loss causes collector to use stale template
- **Firmware Upgrade**: RouterOS update changed exported fields without updating collector
- **Configuration Change**: IPFIX tracking settings modified without flushing collector cache
- **Cache Not Cleared**: Collector restarted but didn't clear old template definitions

## Detection Strategies

The fixed decoder now implements several detection mechanisms:

### 1. Record Size Validation

```
During DataSet.Unmarshal():
- Calculate expected record size from template
- Track actual bytes consumed per record
- Log warning if they don't match
```

### 2. Template Validation on Registration

When a template is registered with the session, it's automatically validated for:
- **Structure**: Proper field count, valid types
- **Field Sizes**: Checks against IPFIX standards for known field types
- **Common Issues**: Detects suspicious configurations
- **Consistency**: Verifies field sizes match RFC specifications

### 3. Buffer Alignment Checking

```
- Monitor remaining buffer size
- Detect if remaining bytes < expected record size
- Log warnings about potential padding vs corruption
```

### 4. Field-Level Error Logging

```
- Log exact position of any read failures
- Include field type, expected length, and bytes consumed
- Helps pinpoint where desynchronization begins
```

### 5. Diagnostic Output

```
DiagnoseTemplateDesynch(templateID, expectedSize, actualDataSize)
- Shows record count calculation
- Identifies remainder bytes
- Suggests likely causes
```

## Template Validation API

### Programmatic Validation

```go
// Validate a template
validator := NewTemplateValidator(templateID, template)
result := validator.Validate()

if !result.Valid {
    log.Println("Template validation failed:")
    log.Println(result.String())
}

// Validate by ID from session
result := ValidateTemplateByID(templateID, session)
```

### Validation Result Fields

- `Valid`: Whether template passed all validation checks
- `RecordSize`: Calculated size of each data record
- `FieldCount`: Number of fields in template
- `Errors`: Critical validation failures
- `Warnings`: Non-critical issues detected
- `IsSuspicious`: True if warnings present

### Example Output

```
Template 259 Validation:
  Status: ✓ VALID
  Record Size: 152 bytes
  Field Count: 34
  Warnings:
    ⚠️  field 10: unusually large fixed size 512
  NOTE: Template decoded but has suspicious characteristics
```

## Enabling Debug Output

Set environment variable to see detailed logging:

```bash
export NETFLOWDEBUG=1
./your-netflow-app
```

This will output:
- Template registration with validation results
- Set processing details  
- Record unmarshaling progress
- Warning messages about mismatches
- Validation errors and suspicious patterns

## Example Debug Output

```
ipfix: register template: id=259 fields=34 (...)
ipfix:   template 259: 34 fields, total size: 152 bytes
ipfix: received set of 4960 bytes
ipfix: warning: template 259 consumed 124 bytes but expected 152 bytes at record 0
ipfix:   possible template desynchronization: data records may be misaligned
```

## Validation Checks

The validator performs these checks:

### Basic Structure
- Template is not nil
- Template has at least one field
- Field count does not exceed 255

### Field Types
- Detects reserved/invalid types
- Identifies suspicious type combinations

### Field Sizes
- Validates against IPFIX RFC 5102 standards
- Checks for unreasonably large fixed-size fields
- Common field type validation:
  - octetDeltaCount: 8 bytes
  - packetDeltaCount: 8 bytes
  - protocolIdentifier: 1 byte
  - sourceIPv4Address: 4 bytes
  - sourceIPv6Address: 16 bytes
  - Etc.

### Common Issues
- Missing common required fields
- Very small field count
- Detects MikroTik IPFIX exports (special handling)

## How to Fix

### Immediate Fix: Reset Collector Cache

1. **Stop the collector**
2. **Clear any cached template definitions** (location depends on your collector)
3. **Restart the collector**
4. The collector will receive fresh template definitions on next export

### Long-term Fix: Verify Template Configuration

1. **Check MikroTik Router**: Verify IPFIX export settings
   ```
   /ip traffic-flow/ipfix
   print detail
   ```

2. **Verify Collector Template Cache**: Ensure it matches router exports
   
3. **Check for Recent Changes**:
   - RouterOS version changes
   - Traffic Flow configuration changes
   - IPFIX field selection changes

4. **Monitor Logs**: Use NETFLOWDEBUG=1 to catch future mismatches

5. **Validate Templates**: Use the programmatic validation API to check templates before processing data

## Code Changes

The decoder now includes:

- `ValidateTemplateSize(template)`: Calculate expected record size
- `NewTemplateValidator(templateID, template)`: Create validator instance
- `Validate()`: Perform comprehensive validation
- `ValidationResult`: Detailed result object
- `ValidateTemplateByID(templateID, session)`: Validate from session
- Enhanced error messages with field/template details
- `DiagnoseTemplateDesynch()`: Detailed diagnostic output
- Debug logging at field level
- Automatic validation during template registration

### Usage Example

```go
template, found := session.GetTemplate(templateID)
if !found {
    return errTemplateNotFound(templateID)
}

expectedSize := ValidateTemplateSize(template)
// expectedSize can be used to validate incoming data

// Validate programmatically
validator := NewTemplateValidator(templateID, template)
result := validator.Validate()
if !result.Valid {
    log.Println(result.String())
    // Handle invalid template
}

// Or validate by ID
result := ValidateTemplateByID(templateID, session)
if result.IsSuspicious {
    log.Println("WARNING:", result.String())
}

// If issues detected:
diagnostic := DiagnoseTemplateDesynch(templateID, expectedSize, dataSize)
log.Println(diagnostic)
```

## Prevention

To prevent future desynchronization:

1. **Monitor Template Changes**: Log template updates and validation results
2. **Validate on Start**: Compare first records to template size
3. **Set Alerts**: Alert on validation failures or suspicious templates
4. **Test Updates**: Before applying firmware updates, test in lab
5. **Document Configuration**: Keep templates backed up and documented
6. **Use Validation API**: Integrate template validation into your collector

## References

- RFC 7011: IPFIX Protocol Specification
- RFC 5102: IPFIX Information Elements
- MikroTik IPFIX Configuration: `/ip traffic-flow/ipfix`
- Common field sizes: RFC 5102, RFC 7022
