// SPDX-License-Identifier: MIT

package collector

import (
	"context"

	"github.com/czerwonk/ovirt_api/api"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type clientTracingAdapter struct {
	client *api.Client
	tracer trace.Tracer
}

// GetAndParse implements Client.GetAndParse
func (cta *clientTracingAdapter) GetAndParse(ctx context.Context, path string, v interface{}) error {
	_, span := cta.tracer.Start(ctx, "Client.RunCommandAndParseWithParser", trace.WithAttributes(
		attribute.String("path", path),
	))
	defer span.End()

	err := cta.client.GetAndParse(path, v)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}
