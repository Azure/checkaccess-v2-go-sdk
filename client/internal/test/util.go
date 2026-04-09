package test

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"github.com/Azure/checkaccess-v2-go-sdk/client/internal"
)

type mockTransport struct {
	statusCode int
}

func CreateTestToken(oid string, fakeClaims *internal.Custom) (string, error) {
	// Define the signing key
	signingKey := []byte("test-secret-key")

	// Create the custom claims
	claims := internal.Custom{
		ObjectId: oid,
		ClaimNames: map[string]interface{}{
			"example_claim": "example_value",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   "test-subject",
			Audience:  []string{"test-audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "unique-id",
		},
	}

	if fakeClaims != nil {
		claims = *fakeClaims
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}

func (m *mockTransport) Do(req *http.Request) (*http.Response, error) {
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
			Transport: &mockTransport{
				statusCode: returnedHttpCode,
			},
		})
}
