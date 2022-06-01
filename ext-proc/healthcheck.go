package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	healthPb "google.golang.org/grpc/health/grpc_health_v1"
)

func (grpcServer *GRPCServer) Check(ctx context.Context, request *healthPb.HealthCheckRequest) (*healthPb.HealthCheckResponse, error) {

	log.Info().Msgf("health check invoked with %v", request)

	return &healthPb.HealthCheckResponse{Status: healthPb.HealthCheckResponse_SERVING}, nil
}

func (grpcServer *GRPCServer) Watch(request *healthPb.HealthCheckRequest, watchServer healthPb.Health_WatchServer) error {

	log.Info().Msgf("health watch invoked with %v", request)

	return status.Error(codes.Unimplemented, "watch is not implemented")
}
