package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/coocood/freecache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	configPb "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	filterPb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	healthPb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	cache    *freecache.Cache
	cacheKey = []byte("key")
)

type server struct{}
type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthPb.HealthCheckRequest) (*healthPb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request + %s", in.String())
	return &healthPb.HealthCheckResponse{Status: healthPb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthPb.HealthCheckRequest, srv healthPb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func (s *server) Process(srv extProcPb.ExternalProcessor_ProcessServer) error {

	log.Println(" ")
	log.Println(" ")
	log.Println("Started process:  -->  ")

	ctx := srv.Context()

	contentType := ""
	csrf := ""

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		log.Println(" ")
		log.Println(" ")
		log.Println("Got stream:  -->  ")

		resp := &extProcPb.ProcessingResponse{}

		switch v := req.Request.(type) {

		case *extProcPb.ProcessingRequest_RequestHeaders:

			log.Println("--- In RequestHeaders processing ...")
			r := req.Request
			h := r.(*extProcPb.ProcessingRequest_RequestHeaders)

			log.Printf("Request: %+v\n", r)
			log.Printf("Headers: %+v\n", h)
			log.Printf("EndOfStream: %v\n", h.RequestHeaders.EndOfStream)

			for _, n := range h.RequestHeaders.Headers.Headers {
				if strings.ToLower(n.Key) == "content-type" {
					contentType = n.Value
				}
				if strings.ToLower(n.Key) == "x-csrf" {
					csrf = n.Value
				}
				if strings.ToLower(n.Key) == "x-cache" {
					_, err := cache.Get(cacheKey)
					if err != nil {
						cache.Set(cacheKey, []byte(n.Value), 1)
					}
				}
			}

			bodyMode := filterPb.ProcessingMode_NONE
			if csrf == "" {
				bodyMode = filterPb.ProcessingMode_BUFFERED
			}

			resp = &extProcPb.ProcessingResponse{
				Response: &extProcPb.ProcessingResponse_RequestHeaders{
					RequestHeaders: &extProcPb.HeadersResponse{
						Response: &extProcPb.CommonResponse{
							HeaderMutation: &extProcPb.HeaderMutation{
								SetHeaders: []*configPb.HeaderValueOption{
									{
										Header: &configPb.HeaderValue{
											Key:   "x-went-into-req-headers",
											Value: "true",
										},
									},
								},
							},
						},
					},
				},
				ModeOverride: &filterPb.ProcessingMode{
					ResponseHeaderMode: filterPb.ProcessingMode_SEND,
					RequestBodyMode:    bodyMode,
				},
			}

			break

		case *extProcPb.ProcessingRequest_RequestBody:

			log.Println("--- In RequestBody processing")
			r := req.Request
			b := r.(*extProcPb.ProcessingRequest_RequestBody)

			log.Printf("Request: %+v\n", r)
			log.Printf("Body: %+v\n", b)
			log.Printf("EndOfStream: %v\n", b.RequestBody.EndOfStream)
			log.Printf("Content Type: %v\n", contentType)

			t := http.Request{
				Method: "POST",
				Header: http.Header{"Content-Type": {contentType}},
				Body:   ioutil.NopCloser(bytes.NewBuffer(b.RequestBody.Body)),
			}

			err := t.ParseMultipartForm(100000)
			if err != nil {
				log.Printf("parse error %v", err)
			}

			for key, value := range t.Form {
				log.Printf("Form key: %v, form value %v\n", key, value)
				if key == "csrf" {
					csrf = value[0]
				}
			}

			resp = &extProcPb.ProcessingResponse{
				Response: &extProcPb.ProcessingResponse_RequestBody{
					RequestBody: &extProcPb.BodyResponse{
						Response: &extProcPb.CommonResponse{
							HeaderMutation: &extProcPb.HeaderMutation{
								SetHeaders: []*configPb.HeaderValueOption{
									{
										Header: &configPb.HeaderValue{
											Key:   "x-went-into-req-body",
											Value: "true",
										},
									},
								},
							},
						},
					},
				},
			}

			break

		case *extProcPb.ProcessingRequest_ResponseHeaders:

			log.Println("--- In ResponseHeaders processing")
			r := req.Request
			h := r.(*extProcPb.ProcessingRequest_ResponseHeaders)

			cacheEntry, _ := cache.Get(cacheKey)

			log.Printf("Request: %+v\n", r)
			log.Printf("Headers: %+v\n", h)
			log.Printf("Content Type: %v\n", contentType)
			log.Printf("Cache entry: %v\n", cacheEntry)

			resp = &extProcPb.ProcessingResponse{
				Response: &extProcPb.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: &extProcPb.HeadersResponse{
						Response: &extProcPb.CommonResponse{
							HeaderMutation: &extProcPb.HeaderMutation{
								SetHeaders: []*configPb.HeaderValueOption{
									{
										Header: &configPb.HeaderValue{
											Key:   "x-went-into-resp-headers",
											Value: "true",
										},
									},
									{
										Header: &configPb.HeaderValue{
											Key:   "x-extracted-csrf",
											Value: csrf,
										},
									},
									{
										Header: &configPb.HeaderValue{
											Key:   "x-extracted-cache",
											Value: string(cacheEntry),
										},
									},
								},
							},
						},
					},
				},
			}

			break

		default:
			log.Printf("Unknown Request type %+v\n", v)
		}

		if err := srv.Send(resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}

func main() {

	// cache init
	cache = freecache.NewCache(1024)
	debug.SetGCPercent(20)

	// grpc server init
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	extProcPb.RegisterExternalProcessorServer(s, &server{})
	healthPb.RegisterHealthServer(s, &healthServer{})

	log.Println("Starting gRPC server on port :50051")

	// shutdown
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		log.Printf("caught sig: %+v", sig)
		log.Println("Wait for 1 second to finish processing")
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	s.Serve(lis)
}
