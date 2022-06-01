package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ekkinox/ext-proc-demo/ext-proc/utils"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	healthPb "google.golang.org/grpc/health/grpc_health_v1"
)

var config utils.Config

type GRPCServer struct{}

func main() {

	// init config
	config = utils.InitConfig()

	// init logger
	utils.InitLogger(config)

	// run
	if config.Flag.HealthCheckMode == false {
		runGRPCServer()
	} else {
		runCLICheck()
	}

	// shutdown
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-exit
}

func runGRPCServer() {

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Env.GRPCPort))
		if err != nil {
			log.Error().Msgf("failed to listen tcp on port %d: %v", config.Env.GRPCPort, err)
		}

		grpcServer := grpc.NewServer()
		healthPb.RegisterHealthServer(grpcServer, &GRPCServer{})
		extProcPb.RegisterExternalProcessorServer(grpcServer, &GRPCServer{})

		if config.Env.GRPCReflection {
			log.Info().Msg("activating gRPC reflection")
			reflection.Register(grpcServer)
		}

		if err := grpcServer.Serve(lis); err != nil {
			log.Error().Msgf("failed to serve: %v", err)
		}
	}()

	log.Info().Msgf("%v gRPC service started on port %v", config.Name, config.Env.GRPCPort)
}

func runCLICheck() {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf(":%d", config.Env.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Error().Msgf("cannot connect to %s gRPC health check service", config.Name)
		os.Exit(1)
	}
	defer conn.Close()

	c := healthPb.NewHealthClient(conn)

	resp, err := c.Check(ctx, &healthPb.HealthCheckRequest{
		Service: fmt.Sprintf("%s CLI health checker", config.Name),
	})
	if err != nil {
		log.Error().Msgf("error during CLI gRPC health check: %v", err)
		os.Exit(1)
	}

	if resp.Status != healthPb.HealthCheckResponse_SERVING {
		log.Error().Msgf("unexpected CLI gRPC health check response status: %v", resp.Status)
		os.Exit(1)
	}

	log.Info().Msgf("success of %v CLI gRPC health check", config.Name)
	os.Exit(0)
}
