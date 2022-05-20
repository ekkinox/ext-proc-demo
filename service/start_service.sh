#!/bin/sh
/service &
envoy -c /src/envoy.yaml --service-cluster "${CLUSTER_NAME}"
