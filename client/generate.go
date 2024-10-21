package client

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

//go:generate rm -rf mocks/$GOPACKAGE
//go:generate mockgen -destination=mocks/$GOPACKAGE/$GOPACKAGE.go github.com/Azure/checkaccess-v2-go-sdk/client RemotePDPClient
//go:generate goimports -local=github.com/Azure/checkaccess-v2-go-sdk -e -w mocks/$GOPACKAGE/$GOPACKAGE.go
