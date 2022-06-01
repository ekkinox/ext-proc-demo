package main

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	configPb "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	filterPb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

func (grpcServer *GRPCServer) Process(stream extProcPb.ExternalProcessor_ProcessServer) error {

	log.Debug().Msg("- Started process -")

	ctx := stream.Context()

	csrf := ""
	contentType := ""

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()

		if err == io.EOF {
			log.Debug().Msg("received stream client termination")
			return nil
		}

		if err != nil {
			msg := fmt.Sprintf("received stream error: %v", err)
			log.Error().Msg(msg)
			return status.Error(codes.Unknown, msg)
		}

		resp := &extProcPb.ProcessingResponse{}

		switch v := req.Request.(type) {

		case *extProcPb.ProcessingRequest_RequestHeaders:

			log.Debug().Msg("--- In RequestHeaders processing ---")
			r := req.Request
			h := r.(*extProcPb.ProcessingRequest_RequestHeaders)

			for _, n := range h.RequestHeaders.Headers.Headers {
				if strings.ToLower(n.Key) == "x-csrf" {
					csrf = n.Value
				}
				if strings.ToLower(n.Key) == "content-type" {
					contentType = n.Value
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

			log.Debug().Msg("--- In RequestBody processing ---")
			r := req.Request
			b := r.(*extProcPb.ProcessingRequest_RequestBody)

			t := http.Request{
				Method: "POST",
				Header: http.Header{"Content-Type": {contentType}},
				Body:   ioutil.NopCloser(bytes.NewBuffer(b.RequestBody.Body)),
			}

			err := t.ParseMultipartForm(100000)
			if err != nil {
				log.Error().Msgf("parse error %v", err)
			}

			for key, value := range t.Form {
				log.Debug().Msgf("Form key: %v, form value %v\n", key, value)
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

			log.Debug().Msg("--- In ResponseHeaders processing ---")

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
								},
							},
						},
					},
				},
			}

			break

		default:
			log.Error().Msgf("unknown Request type %v", v)
		}

		if err := stream.Send(resp); err != nil {
			log.Error().Msgf("stream sending error %v", err)
		}
	}
}
