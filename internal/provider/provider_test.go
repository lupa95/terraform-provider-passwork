// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `provider "passwork" {}`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"passwork": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PASSWORK_API_KEY"); v == "" {
		t.Fatal("PASSWORK_API_KEY must be set for acceptance tests")
	}

	if v := os.Getenv("PASSWORK_HOST"); v == "" {
		t.Fatal("PASSWORK_HOST must be set for acceptance tests")
	}
	
	if v := os.Getenv("PASSWORK_VAULT_ID"); v == "" {
		t.Fatal("PASSWORK_VAULT_ID must be set for acceptance tests")
	}
}
