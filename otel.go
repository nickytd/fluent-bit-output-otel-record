package main

import (
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// LogEntry represents a single log entry with timestamp and record data.
type LogEntry struct {
	Timestamp time.Time
	Record    map[interface{}]interface{}
}

// logValueToString formats a log.Value for display, handling all value types.
func logValueToString(v log.Value) string {
	switch v.Kind() {
	case log.KindBool:
		return fmt.Sprintf("%v", v.AsBool())
	case log.KindInt64:
		return fmt.Sprintf("%d", v.AsInt64())
	case log.KindFloat64:
		return fmt.Sprintf("%f", v.AsFloat64())
	case log.KindString:
		return v.AsString()
	case log.KindBytes:
		return string(v.AsBytes())
	case log.KindSlice:
		var parts []string
		for _, elem := range v.AsSlice() {
			parts = append(parts, logValueToString(elem))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case log.KindMap:
		var parts []string
		for _, kv := range v.AsMap() {
			parts = append(parts, fmt.Sprintf("%s: %s", kv.Key, logValueToString(kv.Value)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case log.KindEmpty:
		return "<empty>"
	default:
		return "<unknown>"
	}
}

// convertToOTelRecord converts a Fluent Bit record to an OpenTelemetry log record.
func convertToOTelRecord(entry LogEntry, resource, scope map[interface{}]interface{}) sdklog.Record {
	var record sdklog.Record

	// Set timestamp
	record.SetTimestamp(entry.Timestamp)
	record.SetObservedTimestamp(time.Now())

	// Set severity based on record content
	severity := determineSeverity(entry.Record)
	record.SetSeverity(severity)
	record.SetSeverityText(severityToText(severity))

	// Convert record body
	body := convertToLogValue(entry.Record)
	record.SetBody(body)

	// Add resource attributes
	if resource != nil {
		attrs := extractAttributes(resource, "resource.")
		record.AddAttributes(attrs...)
	}

	// Add scope attributes
	if scope != nil {
		attrs := extractAttributes(scope, "scope.")
		record.AddAttributes(attrs...)
	}

	return record
}

// determineSeverity determines the log severity based on record content.
func determineSeverity(record map[interface{}]interface{}) log.Severity {
	// Look for common level/severity keys
	levelKeys := []string{"level", "severity", "log_level", "loglevel", "lvl"}

	for _, key := range levelKeys {
		if val, ok := getMapValue(record, key); ok {
			levelStr := strings.ToLower(interfaceToString(val))
			return parseSeverity(levelStr)
		}
	}

	return log.SeverityInfo
}

// parseSeverity converts a string level to OpenTelemetry severity.
func parseSeverity(level string) log.Severity {
	switch level {
	case "trace":
		return log.SeverityTrace
	case "debug":
		return log.SeverityDebug
	case "info", "information":
		return log.SeverityInfo
	case "warn", "warning":
		return log.SeverityWarn
	case "error", "err":
		return log.SeverityError
	case "fatal", "critical", "panic", "emergency":
		return log.SeverityFatal
	default:
		return log.SeverityInfo
	}
}

// severityToText converts severity enum to string representation.
func severityToText(severity log.Severity) string {
	switch severity {
	case log.SeverityTrace, log.SeverityTrace2, log.SeverityTrace3, log.SeverityTrace4:
		return "TRACE"
	case log.SeverityDebug, log.SeverityDebug2, log.SeverityDebug3, log.SeverityDebug4:
		return "DEBUG"
	case log.SeverityInfo, log.SeverityInfo2, log.SeverityInfo3, log.SeverityInfo4:
		return "INFO"
	case log.SeverityWarn, log.SeverityWarn2, log.SeverityWarn3, log.SeverityWarn4:
		return "WARN"
	case log.SeverityError, log.SeverityError2, log.SeverityError3, log.SeverityError4:
		return "ERROR"
	case log.SeverityFatal, log.SeverityFatal2, log.SeverityFatal3, log.SeverityFatal4:
		return "FATAL"
	default:
		return "INFO"
	}
}

// getMapValue retrieves a value from a map by string key.
func getMapValue(m map[interface{}]interface{}, key string) (interface{}, bool) {
	// Try string key
	if val, ok := m[key]; ok {
		return val, true
	}
	// Try []byte key
	if val, ok := m[[]byte(key)]; ok {
		return val, true
	}
	// Iterate and compare string representations
	for k, v := range m {
		if interfaceToString(k) == key {
			return v, true
		}
	}
	return nil, false
}

// extractAttributes converts a map to OpenTelemetry key-value attributes.
func extractAttributes(m map[interface{}]interface{}, prefix string) []log.KeyValue {
	var attrs []log.KeyValue

	for k, v := range m {
		keyStr := prefix + interfaceToString(k)
		attr := convertToKeyValue(keyStr, v)
		attrs = append(attrs, attr)
	}

	return attrs
}

// convertToKeyValue converts an interface value to an OpenTelemetry KeyValue.
func convertToKeyValue(key string, value interface{}) log.KeyValue {
	logValue := convertToLogValue(value)
	return log.KeyValue{Key: key, Value: logValue}
}

// convertToLogValue converts an interface value to an OpenTelemetry log.Value.
func convertToLogValue(v interface{}) log.Value {
	switch val := v.(type) {
	case string:
		return log.StringValue(val)
	case []byte:
		return log.StringValue(string(val))
	case bool:
		return log.BoolValue(val)
	case int:
		return log.Int64Value(int64(val))
	case int8:
		return log.Int64Value(int64(val))
	case int16:
		return log.Int64Value(int64(val))
	case int32:
		return log.Int64Value(int64(val))
	case int64:
		return log.Int64Value(val)
	case uint:
		return log.Int64Value(int64(val))
	case uint8:
		return log.Int64Value(int64(val))
	case uint16:
		return log.Int64Value(int64(val))
	case uint32:
		return log.Int64Value(int64(val))
	case uint64:
		return log.Int64Value(int64(val))
	case float32:
		return log.Float64Value(float64(val))
	case float64:
		return log.Float64Value(val)
	case map[interface{}]interface{}:
		return convertMapToLogValue(val)
	case map[string]interface{}:
		return convertStringMapToLogValue(val)
	case []interface{}:
		return convertSliceToLogValue(val)
	case []map[interface{}]interface{}:
		return convertMapSliceToLogValue(val)
	default:
		return log.StringValue(interfaceToString(v))
	}
}

// convertMapToLogValue converts a map to an OpenTelemetry map value.
func convertMapToLogValue(m map[interface{}]interface{}) log.Value {
	kvs := make([]log.KeyValue, 0, len(m))
	for k, v := range m {
		keyStr := interfaceToString(k)
		kvs = append(kvs, log.KeyValue{Key: keyStr, Value: convertToLogValue(v)})
	}
	return log.MapValue(kvs...)
}

// convertStringMapToLogValue converts a string map to an OpenTelemetry map value.
func convertStringMapToLogValue(m map[string]interface{}) log.Value {
	kvs := make([]log.KeyValue, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, log.KeyValue{Key: k, Value: convertToLogValue(v)})
	}
	return log.MapValue(kvs...)
}

// convertSliceToLogValue converts a slice to an OpenTelemetry slice value.
func convertSliceToLogValue(s []interface{}) log.Value {
	// Check if it's a byte array
	if isBytes(s) {
		return log.StringValue(interfaceSliceToString(s))
	}

	values := make([]log.Value, len(s))
	for i, v := range s {
		values[i] = convertToLogValue(v)
	}
	return log.SliceValue(values...)
}

// convertMapSliceToLogValue converts a slice of maps to an OpenTelemetry slice value.
func convertMapSliceToLogValue(s []map[interface{}]interface{}) log.Value {
	values := make([]log.Value, len(s))
	for i, m := range s {
		values[i] = convertMapToLogValue(m)
	}
	return log.SliceValue(values...)
}

// interfaceToString converts an interface to its string representation.
func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case []interface{}:
		if isBytes(val) {
			return interfaceSliceToString(val)
		}
		return ""
	default:
		return ""
	}
}

// isRecordGroupStart checks if a record marks the start of a record group.
func isRecordGroupStart(record map[interface{}]interface{}) bool {
	hasResource := false
	hasScope := false

	for k := range record {
		keyStr := interfaceToString(k)
		if keyStr == "resource" {
			hasResource = true
		}
		if keyStr == "scope" {
			hasScope = true
		}
	}

	return hasResource && hasScope
}

// isRecordGroupEnd checks if a record marks the end of a record group (empty record).
func isRecordGroupEnd(record map[interface{}]interface{}) bool {
	return len(record) == 0
}

// extractResourceAndScope extracts resource and scope maps from a group start record.
func extractResourceAndScope(record map[interface{}]interface{}) (resource, scope map[interface{}]interface{}) {
	for k, v := range record {
		keyStr := interfaceToString(k)
		switch keyStr {
		case "resource":
			if m, ok := v.(map[interface{}]interface{}); ok {
				resource = m
			}
		case "scope":
			if m, ok := v.(map[interface{}]interface{}); ok {
				scope = m
			}
		}
	}
	return resource, scope
}
