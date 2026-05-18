package ipfix

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

var debug = os.Getenv("NETFLOWDEBUG") != ""
var debugLog = log.New(os.Stderr, "ipfix: ", log.Lmicroseconds|log.Lmicroseconds)

func hexdump(data []byte) {
	fmt.Fprint(os.Stderr, hex.Dump(data))
}

// DiagnoseTemplateDesynch provides detailed information about potential template desynchronization.
// It compares expected template size with actual data to help identify misalignment issues.
func DiagnoseTemplateDesynch(templateID uint16, expectedSize, actualDataSize int) string {
	if actualDataSize == 0 {
		return fmt.Sprintf("Template %d: No data available for diagnosis", templateID)
	}
	
	if expectedSize == 0 {
		return fmt.Sprintf("Template %d: Invalid template (size=0)", templateID)
	}
	
	recordCount := actualDataSize / expectedSize
	remainder := actualDataSize % expectedSize
	
	msg := fmt.Sprintf("Template %d Analysis:\n", templateID)
	msg += fmt.Sprintf("  Expected record size: %d bytes\n", expectedSize)
	msg += fmt.Sprintf("  Available data: %d bytes\n", actualDataSize)
	msg += fmt.Sprintf("  Calculated records: %d\n", recordCount)
	msg += fmt.Sprintf("  Remaining bytes: %d\n", remainder)
	
	if remainder != 0 {
		msg += fmt.Sprintf("  ⚠️  WARNING: Data size %d is not divisible by template size %d\n", 
			actualDataSize, expectedSize)
		msg += fmt.Sprintf("  ⚠️  This indicates TEMPLATE DESYNCHRONIZATION or corrupted data\n")
		msg += fmt.Sprintf("  ⚠️  Possible causes:\n")
		msg += fmt.Sprintf("       - Template definition changed but collector cache not updated\n")
		msg += fmt.Sprintf("       - Missed template packet due to UDP packet loss\n")
		msg += fmt.Sprintf("       - RouterOS firmware upgrade changed exported fields\n")
	}
	
	return msg
}
