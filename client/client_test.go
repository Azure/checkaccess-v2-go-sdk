package client

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/go-cmp/cmp"

	testhttp "github.com/Azure/ARO-RP/test/util/http/server"
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

func TestCallingCheckAccess(t *testing.T) {
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
	for _, c := range cases {
		srv, close := testhttp.NewTLSServer()
		srv.SetResponse(testhttp.WithStatusCode(c.returnedHttpCode))
		client := createClientWithServer(srv)
		t.Run(c.desc, func(t *testing.T) {
			decision, err := client.CheckAccess(context.Background(), AuthorizationRequest{})
			if decision != c.expectedDecision && err != c.expectedErr {
				t.Errorf("expected decision to be %v; and error to be %s. Got %v and %s",
					c.expectedDecision, c.expectedErr, decision, err)
			}
		})
		close()
	}
}

func TestCreateAuthorizationRequest(t *testing.T) {
	t.Parallel()
	endpoint := "https://westus.authorization.azure.net/providers/Microsoft.Authorization/checkAccess?api-version=2021-06-01-preview"
	scope := "https://authorization.azure.net/.default"
	cred, err := azidentity.NewClientSecretCredential("888988bf-86f1-31ea-91cd-2d7cd011db48", "clientID", "clientSecret", nil)
	if err != nil {
		t.Error("Unable to create a new client secret credential")
	}

	resourceId := "resource456"
	subjectAttributes := SubjectAttributes{
		ObjectId:  "object123",
		ClaimName: "claim789",
	}
	actions := []string{"read", "write"}

	expected := AuthorizationRequest{
		Subject: SubjectInfo{
			Attributes: subjectAttributes,
		},
		Actions: []ActionInfo{
			{Id: "read"},
			{Id: "write"},
		},
		Resource: ResourceInfo{
			Id: resourceId,
		},
	}
	client, err := NewRemotePDPClient(endpoint, scope, cred, nil)
	if err != nil {
		t.Error("Unable to create a new PDP client")
	}
	result := client.CreateAuthorizationRequest(resourceId, actions, subjectAttributes)

	if diff := cmp.Diff(result, expected); diff != "" {
		t.Errorf("incorrect authorization request: %v", diff)
	}
}

func createClientWithServer(s *testhttp.Server) RemotePDPClient {
	return &remotePDPClient{
		endpoint: s.URL(),
		pipeline: runtime.NewPipeline(
			"remotepdpclient_test",
			"v1.0.0",
			runtime.PipelineOptions{},
			&policy.ClientOptions{Transport: s},
		),
	}
}
