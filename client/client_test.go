package client

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"errors"
	"net/http"
	"testing"

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
		cred        *azidentity.ClientSecretCredential
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
			cred:        nil,
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
		_, err := NewRemotePDPClient(c.endpoint, c.scope, c.cred, nil)
		if c.expectedErr && err == nil {
			t.Errorf("%s: expected error to be 'non-nil' but got 'nil'", c.desc)
		}
		if !c.expectedErr && err != nil {
			t.Errorf("%s: expected error to be 'nil' but got '%v'", c.desc, err)
		}
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
		decision, err := client.CheckAccess(context.Background(), AuthorizationRequest{})
		if decision != c.expectedDecision && err != c.expectedErr {
			t.Errorf("%s: expected decision to be %v; and error to be %s. Got %v and %s",
				c.desc, c.expectedDecision, c.expectedErr, decision, err)
		}
		close()
	}
}

func TestCreateAuthorizationRequest(t *testing.T) {
	subject := "subject123"
	resourceId := "resource456"
	claimName := "claim789"
	actions := []string{"read", "write"}

	expected := AuthorizationRequest{
		Subject: SubjectInfo{
			Attributes: SubjectAttributes{
				ObjectId:  subject,
				ClaimName: claimName,
			},
		},
		Actions: []ActionInfo{
			{Id: "read"},
			{Id: "write"},
		},
		Resource: ResourceInfo{
			Id: resourceId,
		},
	}

	result := CreateAuthorizationRequest(subject, resourceId, claimName, actions...)

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
