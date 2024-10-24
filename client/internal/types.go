package internal

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.
import (
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/golang-jwt/jwt/v4"
)

type Custom struct {
	ObjectId   string                 `json:"oid"`
	ClaimNames map[string]interface{} `json:"_claim_names"`
	Groups     []string               `json:"groups"`
	jwt.RegisteredClaims
}

type MockTransport struct {
	statusCode int
}

func (m *MockTransport) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Header:     make(http.Header),
	}, nil
}

func CreatePipelineWithServer(returnedHttpCode int) runtime.Pipeline {
	return runtime.NewPipeline(
		"remotepdpclient_test",
		"v0.1.0",
		runtime.PipelineOptions{},
		&policy.ClientOptions{
			Transport: &MockTransport{
				statusCode: returnedHttpCode,
			},
		})
}
