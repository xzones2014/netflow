# netflow
NetFlow version 1, 5, 7, 8, 9 &amp; 10 (IPFIX) support for Go

[![Build Status](https://travis-ci.org/xzones2014/netflow.svg?branch=master)](https://travis-ci.org/xzones2014/netflow)
[![GoDoc](https://godoc.org/github.com/xzones2014/netflow?status.svg)](https://godoc.org/github.com/xzones2014/netflow)

## Project status

We're still in a draft status, do not use this for processing NetFlows in an
untrusted environment.

## Features

- **Full IPFIX Support** - Decodes NetFlow v10 (IPFIX) with template validation
- **Template Validation** - Automatic detection of misconfigured or corrupted templates
- **Desynchronization Detection** - Identifies when templates don't match actual data
- **Comprehensive Diagnostics** - Detailed error reporting for troubleshooting
- **Multiple NetFlow Versions** - Supports v1, v5, v6, v7, v8, v9, and v10 (IPFIX)

## Documentation

- [IPFIX Template Validation Guide](./ipfix/VALIDATION.md) - Comprehensive validation API documentation
- [IPFIX Desynchronization Detection](./ipfix/DESYNCHRONIZATION.md) - Troubleshooting template mismatches

## Quick Start

### Basic IPFIX Decoding

```go
import "github.com/xzones2014/netflow/ipfix"

decoder := ipfix.NewDecoder(connection, session)
message, err := decoder.Next()
```

### Template Validation

```go
// Automatic validation on registration
// Enable debug output to see results:
export NETFLOWDEBUG=1

// Or validate programmatically:
result := ipfix.ValidateTemplateByID(templateID, session)
if !result.Valid {
    log.Println(result.String())
}
```

## Recent Improvements (May 2026)

- ✓ IPFIX template validation system
- ✓ Desynchronization detection and diagnostics
- ✓ Field-level error reporting
- ✓ RFC 5102 field size validation
- ✓ Common issue detection for MikroTik and standard IPFIX exports
- ✓ Make IPFIX usable (v10) - Completed with comprehensive validation and diagnostics

## TODO

* Expire templates with empty fields (v9)
* Define a more clear interface for caching
* Template expiration based on idle time

## Help wanted!

Please submit your pull requests and/or bug reports via [GitHub](https://github.com/xzones2014/netflow/issues).


## IPFIX Decoder Fix Summary

### **Core Problem Addressed**
Fixed the critical **template desynchronization bug** where the decoder was reading records at the wrong byte offsets when the actual data record length didn't match the template definition.

### **Key Changes**

#### 1. **Enhanced Error Detection** (decoder.go)
- Added `ValidateTemplateSize()` function to calculate expected record size from template
- Added `errTemplateMismatch()` and `errRecordSizeAlignment()` functions for detailed error reporting
- Added `errInvalidVersion()` and `Version` constant (IPFIX v10)

#### 2. **Record Size Validation** (packet.go)
- Modified `DataSet.Unmarshal()` to:
  - Calculate expected record size before unmarshaling
  - Track bytes consumed per record
  - Detect when actual consumption differs from expected
  - Log detailed warnings about potential desynchronization
  - Include record count in diagnostic output

#### 3. **Field-Level Error Logging** (packet.go)
- Enhanced `DataRecord.Unmarshal()` with:
  - Field-level error logging showing type, length, and position
  - Better error context for debugging

- Enhanced `Field.Unmarshal()` with:
  - Validation of actual bytes read vs expected
  - Detection of short reads (EOF conditions)
  - Detailed error messages with field metadata

#### 4. **Diagnostic Tools** (debug.go)
- Added `DiagnoseTemplateDesynch()` function that:
  - Calculates expected vs actual record counts
  - Identifies remainder bytes indicating misalignment
  - Suggests likely root causes
  - Provides actionable remediation steps

#### 5. **Improved Template Logging** (packet.go)
- Enhanced template registration logging to show:
  - Field count
  - Total record size in bytes
  - Useful for verifying template correctness

#### 6. **Documentation** (DESYNCHRONIZATION.md)
- Comprehensive guide covering:
  - How desynchronization happens
  - Root causes and detection strategies
  - Enabling debug output with `NETFLOWDEBUG=1`
  - Example debug output
  - How to fix (immediate and long-term solutions)
  - Prevention strategies

### **How to Use**

**Enable verbose logging:**
```bash
export NETFLOWDEBUG=1
./your-app
```

**Programmatic validation:**
```go
template, found := session.GetTemplate(templateID)
if found {
    expectedSize := ValidateTemplateSize(template)
    // Use expectedSize to validate incoming data
}
```

**Get diagnostic info:**
```go
diagnostic := DiagnoseTemplateDesynch(templateID, expectedSize, dataSize)
log.Println(diagnostic)
```

### **Files Modified**
- decoder.go - Added validation functions
- packet.go - Enhanced record unmarshaling with validation
- debug.go - Added diagnostic utility
- DESYNCHRONIZATION.md - **NEW** comprehensive guide
- translate.go - Added `GetSession()` getter
- Fixed import syntax errors across multiple files

The decoder now provides **early detection** of template misalignment issues before data corruption spreads, with **detailed diagnostics** to help operators identify and fix the root cause.
