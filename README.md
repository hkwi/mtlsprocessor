# mTLS Client Info Processor for OpenTelemetry Collector

This module is an OpenTelemetry Collector processor that extracts and expands receiver mTLS (mutual TLS) client certificate information into OpenTelemetry resource attributes. It is designed to help observability pipelines enrich telemetry data with client identity details derived from mTLS connections.

## Features

- Extracts client certificate information (such as subject, issuer, and SANs) from incoming mTLS connections.
- Adds extracted information as resource attributes to OpenTelemetry data (traces, metrics, logs).
- Supports configurable attribute name prefixing to avoid collisions or namespace attributes.

## Build

To build and generate the OpenTelemetry Collector with this processor, use the [ocb](https://github.com/open-telemetry/opentelemetry-collector-builder) (OpenTelemetry Collector Builder) tool. Follow these steps:

1. Install `ocb` if you haven't already.
2. Run `ocb` in this directory with your configuration file (e.g., `builder-config.yaml`).
3. The resulting custom OpenTelemetry Collector binary will include this processor.

For more details, see the [OpenTelemetry Collector Builder documentation](https://github.com/open-telemetry/opentelemetry-collector-builder).


## Configuration

The processor can be configured in the OpenTelemetry Collector pipeline YAML. Example configuration:

```yaml
processors:
  mtls:
    prefix: "gateway."
```

- `prefix` (optional): String to prefix all added resource attribute names. If omitted, no prefix is used.

## Example Usage

Add the processor to your OpenTelemetry Collector pipeline:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        tls:
          cert_file: ./certs/server.pem
          key_file: ./certs/server-key.pem
          client_ca_file: ./certs/client_ca.pem

processors:
  mtls:
    prefix: "gateway."

exporters:
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [mtls]
      exporters: [logging]
```

## Extracted Attributes

The processor will add resource attributes, following OpenTelemetry semantic convention:
- `<prefix>tls.client.subject`
- `<prefix>tls.client.issuer`
- `<prefix>tls.client.not_before`
- `<prefix>tls.client.not_after`

https://opentelemetry.io/docs/specs/semconv/registry/attributes/tls/

