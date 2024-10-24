package token

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"

	"github.com/Azure/checkaccess-v2-go-sdk/client/internal"
	"github.com/Azure/checkaccess-v2-go-sdk/client/internal/test"
)

func TestExtractClaims(t *testing.T) {
	dummyObjectId := "1234567890"
	claims := &internal.Custom{
		ObjectId: dummyObjectId,
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
	validTestToken, err := test.CreateTestToken(dummyObjectId, claims)
	if err != nil {
		t.Errorf("Error creating test token: %v", err)
	}

	tests := []struct {
		name       string
		token      string
		wantClaims *internal.Custom
		wantErr    bool
	}{
		{
			name:       "Can extract oid from a valid token",
			token:      validTestToken,
			wantClaims: claims,
			wantErr:    false,
		},
		{
			name:       "Return an error when given an invalid jwt",
			token:      "invalid",
			wantClaims: nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractClaims(tt.token)
			t.Log(got, err)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expect an error but got nothing")
				}
			} else {
				if diff := cmp.Diff(got, tt.wantClaims); diff != "" {
					t.Errorf("Got: %q, want %q", got, tt.wantClaims)
				}
				if err != nil {
					t.Errorf("Expect no error but got one")
				}
			}
		})
	}
}
