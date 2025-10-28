package mtlsprocessor

import (
	"context"
	"crypto/x509"
	"time"

	"github.com/hkwi/mtlsauthextension"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/xprocessor"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type Config struct {
	Prefix string `mapstructure:"prefix"`
}

func NewFactory() processor.Factory {
	return xprocessor.NewFactory(
		component.MustNewType("mtls"),
		func() component.Config {
			return Config{}
		},
		xprocessor.WithLogs(createLogs, component.StabilityLevelAlpha),
		xprocessor.WithMetrics(createMetrics, component.StabilityLevelAlpha),
		xprocessor.WithTraces(createTraces, component.StabilityLevelAlpha),
	)
}

func peerCertificate(ctx context.Context) *x509.Certificate {
	cli := client.FromContext(ctx)
	if cli.Auth != nil {
		// depends on mtlsauthextension (GRPC and HTTP)
		if cert, ok := cli.Auth.(*mtlsauthextension.PeerInfo); ok {
			return (*x509.Certificate)(cert)
		}
	}
	if peer, ok := peer.FromContext(ctx); ok {
		// GRPC only fallback; no HTTP support
		if peer != nil && peer.AuthInfo != nil {
			if tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo); ok {
				if len(tlsInfo.State.PeerCertificates) > 0 {
					return tlsInfo.State.PeerCertificates[0]
				}
			}
		}
	}
	return nil
}

func putInAttribute(config Config, cert *x509.Certificate, attr pcommon.Map) {
	attr.PutStr(
		config.Prefix+"tls.client.subject",
		cert.Subject.String(),
	)
	attr.PutStr(
		config.Prefix+"tls.client.issuer",
		cert.Issuer.String(),
	)
	attr.PutStr(
		config.Prefix+"tls.client.not_before",
		cert.NotBefore.Format(time.RFC3339),
	)
	attr.PutStr(
		config.Prefix+"tls.client.not_after",
		cert.NotAfter.Format(time.RFC3339),
	)
}

func createLogs(ctx context.Context, settings processor.Settings, cfg component.Config, next consumer.Logs) (processor.Logs, error) {
	config := cfg.(Config)
	return processorhelper.NewLogs(
		ctx,
		settings,
		cfg,
		next,
		func(ctx context.Context, l plog.Logs) (plog.Logs, error) {
			if cert := peerCertificate(ctx); cert != nil {
				for _, rl := range l.ResourceLogs().All() {
					resourceAttrs := rl.Resource().Attributes()
					putInAttribute(config, cert, resourceAttrs)
				}
			}
			return l, nil
		},
	)
}

func createMetrics(ctx context.Context, settings processor.Settings, cfg component.Config, next consumer.Metrics) (processor.Metrics, error) {
	config := cfg.(Config)
	return processorhelper.NewMetrics(
		ctx,
		settings,
		cfg,
		next,
		func(ctx context.Context, m pmetric.Metrics) (pmetric.Metrics, error) {
			if cert := peerCertificate(ctx); cert != nil {
				for _, rm := range m.ResourceMetrics().All() {
					resourceAttrs := rm.Resource().Attributes()
					putInAttribute(config, cert, resourceAttrs)
				}
			}
			return m, nil
		},
	)
}

func createTraces(ctx context.Context, settings processor.Settings, cfg component.Config, next consumer.Traces) (processor.Traces, error) {
	config := cfg.(Config)
	return processorhelper.NewTraces(
		ctx,
		settings,
		cfg,
		next,
		func(ctx context.Context, t ptrace.Traces) (ptrace.Traces, error) {
			if cert := peerCertificate(ctx); cert != nil {
				for _, rs := range t.ResourceSpans().All() {
					resourceAttrs := rs.Resource().Attributes()
					putInAttribute(config, cert, resourceAttrs)
				}
			}
			return t, nil
		},
	)
}
