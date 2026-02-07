# fluent-bit-output-otel-record

This is an Fluent Bit output plugin for experimenting with OpenTelemetry record format. 
The intent is to be able to cope with fluent-bit record groups and to be able to export 
them in a format that can be easily consumed by OpenTelemetry Logs SDK.
Ideally, this plugin shall be able to transform incoming fluent-bit records to 
[otel sdk log.Record](https://pkg.go.dev/go.opentelemetry.io/otel/sdk/log#Record) format,
carrying all scope and resource attributes from the log group.

## Running the plugin
The example requires fluent-bit binary and plugin to be built and available.
1. Build the plugin and fluent-bit binary
```bash
make run
```

The fluent-bit configuration is carried by [fluent-bit.yaml](./fluent-bit.yaml) file. 
It is uses [fluent-bit opentelemetry_envelop](https://docs.fluentbit.io/manual/data-pipeline/processors/opentelemetry-envelope)
to generate log groups with scope and resource attributes.

Ot generates 1 log group with single log record.
```text
[2026/02/07 10:04:40.152650000] [ info] [engine] Shutdown Grace Period=5, Shutdown Input Grace Period=2
[0] test: [2106-02-07 07:28:15 +0100 CET, {"resource": {"attributes": {"service.name": "random"}} "scope": {"name": "gstdout", "version": "0.0.0-dev"} }
[1] test: [2026-02-07 10:04:41.153944969 +0100 CET, {"log": "Running output plugin" "level": "info" "tag": "test" }
[2] test: [2106-02-07 07:28:14 +0100 CET, {}
```

The plugin shall generate output in the following format:
```json
{
  "resource": {
    "attributes": {
      "service.name": "random"
    }
  },
  "scope": {
    "name": "gstdout",
    "version": "0.0.0-dev"
  },
  "log_records": [
    {
      "time_unix_nano": 1707294495000000000,
      "observed_time_unix_nano": 1707294495000000000,
      "severity_number": 9,
      "severity_text": "INFO",
      "body": {
        "string_value": "{\"log\":\"Running output plugin\",\"level\":\"info\",\"tag\":\"test\"}"
      },
      "attributes": []
    }
  ]
}
```