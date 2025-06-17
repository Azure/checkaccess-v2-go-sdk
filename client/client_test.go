package client

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/Azure/checkaccess-v2-go-sdk/client/internal"
	"github.com/Azure/checkaccess-v2-go-sdk/client/internal/test"
)

func TestClientCreate(t *testing.T) {
	endpoint := "https://westus.authorization.azure.net/providers/Microsoft.Authorization/checkAccess?api-version=2021-06-01-preview"
	scope := "https://authorization.azure.net/.default"
	emptyString := "   "
	cred, err := azidentity.NewClientSecretCredential("888988bf-86f1-31ea-91cd-2d7cd011db48", "clientID", "clientSecret", nil)
	if err != nil {
		t.Error("Unable to create a new client secret credential")
	}

	cases := []struct {
		desc        string
		endpoint    string
		scope       string
		cred        azcore.TokenCredential
		expectedErr bool
	}{
		{
			desc:        "fail - Invalid endpoint",
			endpoint:    emptyString,
			scope:       scope,
			cred:        cred,
			expectedErr: true,
		}, {
			desc:        "fail - Invalid scope",
			endpoint:    endpoint,
			scope:       emptyString,
			cred:        cred,
			expectedErr: true,
		}, {
			desc:        "fail - Invalid credential",
			endpoint:    endpoint,
			scope:       scope,
			expectedErr: true,
		}, {
			desc:        "success - successful creation of client",
			endpoint:    endpoint,
			scope:       scope,
			cred:        cred,
			expectedErr: false,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			_, err := NewRemotePDPClient(c.endpoint, c.scope, c.cred, nil)
			if c.expectedErr && err == nil {
				t.Errorf("expected error to be 'non-nil' but got '%v'", err)
			}
			if !c.expectedErr && err != nil {
				t.Errorf("expected error to be 'nil' but got '%v'", err)
			}
		})
	}
}

func TestCheckAccess(t *testing.T) {
	t.Parallel()
	endpoint := "https://westus.authorization.azure.net/providers/Microsoft.Authorization/checkAccess?api-version=2021-06-01-preview"

	cases := []struct {
		desc             string
		returnedHttpCode int
		expectedDecision *AuthorizationDecisionResponse
		expectedErr      error
	}{
		{
			desc:             "Successful calls should return an access decision",
			returnedHttpCode: http.StatusOK,
			expectedDecision: &AuthorizationDecisionResponse{},
			expectedErr:      nil,
		}, {
			desc:             "Call resulting in a failure should return an error",
			returnedHttpCode: http.StatusUnauthorized,
			expectedDecision: nil,
			expectedErr:      errors.New("An error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.desc, func(t *testing.T) {
			mockPipeline := test.CreatePipelineWithServer(tt.returnedHttpCode)
			client := &remotePDPClient{endpoint, mockPipeline}
			decision, err := client.CheckAccess(context.Background(), AuthorizationRequest{})
			if decision != tt.expectedDecision && !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected decision to be %v; and error to be %s. Got %v and %s",
					tt.expectedDecision, tt.expectedErr, decision, err)
			}
		})
	}
}

func TestCreateAuthorizationRequest(t *testing.T) {
	t.Parallel()
	endpoint := "https://westus.authorization.azure.net/providers/Microsoft.Authorization/checkAccess?api-version=2021-06-01-preview"
	scope := "https://authorization.azure.net/.default"
	actionInfo := []ActionInfo{{Id: "read"}, {Id: "write"}}
	actions := []string{"read", "write"}
	dummyObjectId := "1234567890"
	resourceId := "https://management.azure.com/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm"
	cred, err := azidentity.NewClientSecretCredential("888988bf-86f1-31ea-91cd-2d7cd011db48", "clientID", "clientSecret", nil)
	if err != nil {
		t.Error("Unable to create a new client secret credential")
	}

	client, err := NewRemotePDPClient(endpoint, scope, cred, nil)
	if err != nil {
		t.Error("Unable to create a new PDP client")
	}

	for _, tt := range []struct {
		name                     string
		claims                   *internal.Custom
		wantAuthorizationRequest *AuthorizationRequest
		wantErr                  string
	}{
		{
			name:    "fail - invalid token",
			wantErr: "need token in creating AuthorizationRequest",
		},
		{
			name: "pass - don't set claimName or groups when both exist",
			claims: &internal.Custom{
				ObjectId: dummyObjectId,
				ClaimNames: map[string]interface{}{
					"groups": "src1",
				},
				Groups: []string{"group1", "group2"},
			},
			wantAuthorizationRequest: &AuthorizationRequest{
				Subject: SubjectInfo{
					Attributes: SubjectAttributes{ObjectId: dummyObjectId},
				},
				Actions: actionInfo,
				Resource: ResourceInfo{
					Id: resourceId,
				},
			},
		},
		{
			name: "pass - don't set claimName or groups when both don't exist",
			claims: &internal.Custom{
				ObjectId: dummyObjectId,
			},
			wantAuthorizationRequest: &AuthorizationRequest{
				Subject: SubjectInfo{
					Attributes: SubjectAttributes{ObjectId: dummyObjectId},
				},
				Actions: actionInfo,
				Resource: ResourceInfo{
					Id: resourceId,
				},
			},
		},
		{
			name: "pass - set claimName when claimName exits and groups don't exist",
			claims: &internal.Custom{
				ObjectId: dummyObjectId,
				ClaimNames: map[string]interface{}{
					"groups": "src1",
				},
			},
			wantAuthorizationRequest: &AuthorizationRequest{
				Subject: SubjectInfo{
					Attributes: SubjectAttributes{ObjectId: dummyObjectId, ClaimName: GroupExpansion},
				},
				Actions: actionInfo,
				Resource: ResourceInfo{
					Id: resourceId,
				},
			},
		},
		{
			name: "pass - set groups when groups exits and claimName don't exist",
			claims: &internal.Custom{
				ObjectId: dummyObjectId,
				Groups:   []string{"group1", "group2"},
			},
			wantAuthorizationRequest: &AuthorizationRequest{
				Subject: SubjectInfo{
					Attributes: SubjectAttributes{ObjectId: dummyObjectId, Groups: []string{"group1", "group2"}},
				},
				Actions: actionInfo,
				Resource: ResourceInfo{
					Id: resourceId,
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			testtoken := ""
			if tt.claims != nil {
				testtoken, err = test.CreateTestToken(dummyObjectId, tt.claims)
				if err != nil {
					t.Errorf("Error creating test token: %v", err)
				}
			}

			result, err := client.CreateAuthorizationRequest(resourceId, actions, testtoken)
			if tt.wantErr != "" && err == nil {
				t.Errorf("expected error to be '%s' but got '%s'", tt.wantErr, err)
			}
			if tt.wantErr == "" && err != nil {
				t.Errorf("expected error to be 'nil' but got '%s'", err)
			}
			if diff := cmp.Diff(result, tt.wantAuthorizationRequest); diff != "" {
				t.Errorf("incorrect authorization request: %v", diff)
			}
		})
	}
}
