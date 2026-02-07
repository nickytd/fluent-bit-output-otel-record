package main

import (
	"C"
	"fmt"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var count int
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}

	// Create Fluent Bit decoder
	dec := output.NewDecoder(data, int(length))

	// Fluent-Bit now supports record groups
	// Each group starts wth scope and resource keys, followed by the actual records and ends with an empty record.
	// [0] test: [2106-02-07 07:28:15 +0100 CET, {"scope": "map[]" "resource": "map[]" }
	// [1] test: [2026-02-06 14:48:02.347278118 +0100 CET, {"tag": "test" "rand_value": "9223372036854775807" }
	// [2] test: [2106-02-07 07:28:14 +0100 CET, {}

	// Track current record group state (optional - provides resource/scope context)
	var currentResource map[interface{}]interface{}
	var currentScope map[interface{}]interface{}

	// Iterate Records
	count = 0
	for {
		// Extract Record
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		var timestamp time.Time
		switch t := ts.(type) {
		case output.FLBTime:
			timestamp = ts.(output.FLBTime).Time
		case uint64:
			timestamp = time.Unix(int64(t), 0)
		default:
			fmt.Println("time provided invalid, defaulting to now.")
			timestamp = time.Now()
		}

		// Print record keys and values
		fmt.Printf("[%d] %s: [%s, {", count, C.GoString(tag), timestamp.String())
		for k, v := range record {
			fmt.Printf("\"%s\": ", k)
			printValue(v)
			fmt.Printf(" ")
		}
		fmt.Printf("}\n")
		count++

		// Generate otel sdk logs.Record for each record
		// Record groups are optional - if present, they provide resource/scope context
		// A record group starts with {"resource": map{}, "scope": map{}}
		// followed by actual records and ends with an empty record {}

		if isRecordGroupStart(record) {
			// Extract resource and scope for this group
			currentResource, currentScope = extractResourceAndScope(record)
			fmt.Printf("  -> Record group started (resource: %v, scope: %v)\n", currentResource != nil, currentScope != nil)
			continue
		}

		if isRecordGroupEnd(record) {
			// End of record group, reset state
			currentResource = nil
			currentScope = nil
			fmt.Printf("  -> Record group ended\n")
			continue
		}

		// Generate OTel log record for all records (record groups are optional)
		entry := LogEntry{
			Timestamp: timestamp,
			Record:    record,
		}
		otelRecord := convertToOTelRecord(entry, currentResource, currentScope)

		// Print the generated OTel record for debugging
		fmt.Printf("  -> OTel Record: timestamp=%s, severity=%s, body=%s\n",
			otelRecord.Timestamp().String(),
			otelRecord.SeverityText(),
			logValueToString(otelRecord.Body()),
		)
	}

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	return output.FLBPluginRegister(def, "gstdout", "Stdout GO!")
}

// (fluentbit will call this)
// plugin (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
