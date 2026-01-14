package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderMetadata(t *testing.T) {
	p := New("test-version")()

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "trueform" {
		t.Errorf("TypeName = %v, want trueform", resp.TypeName)
	}
	if resp.Version != "test-version" {
		t.Errorf("Version = %v, want test-version", resp.Version)
	}
}

func TestProviderSchema(t *testing.T) {
	p := New("test")()

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", resp.Diagnostics)
	}

	// Check required attributes exist
	schema := resp.Schema

	if _, ok := schema.Attributes["host"]; !ok {
		t.Error("Schema missing 'host' attribute")
	}
	if _, ok := schema.Attributes["api_key"]; !ok {
		t.Error("Schema missing 'api_key' attribute")
	}
	if _, ok := schema.Attributes["verify_ssl"]; !ok {
		t.Error("Schema missing 'verify_ssl' attribute")
	}
}

func TestProviderResources(t *testing.T) {
	p := New("test")()

	resources := p.Resources(context.Background())

	expectedResources := []string{
		"pool",
		"dataset",
		"snapshot",
		"share_smb",
		"share_nfs",
		"user",
		"vm",
		"vm_device",
		"app",
		"cronjob",
		"iscsi_portal",
		"iscsi_target",
		"iscsi_extent",
		"iscsi_initiator",
		"iscsi_targetextent",
		"certificate",
		"static_route",
	}

	if len(resources) != len(expectedResources) {
		t.Errorf("Expected %d resources, got %d", len(expectedResources), len(resources))
	}

	// Verify each resource can be instantiated
	for i, resourceFunc := range resources {
		resource := resourceFunc()
		if resource == nil {
			t.Errorf("Resource %d returned nil", i)
		}
	}
}

func TestProviderDataSources(t *testing.T) {
	p := New("test")()

	dataSources := p.DataSources(context.Background())

	expectedDataSources := []string{
		"pool",
		"dataset",
		"user",
		"vm",
	}

	if len(dataSources) != len(expectedDataSources) {
		t.Errorf("Expected %d data sources, got %d", len(expectedDataSources), len(dataSources))
	}

	// Verify each data source can be instantiated
	for i, dsFunc := range dataSources {
		ds := dsFunc()
		if ds == nil {
			t.Errorf("Data source %d returned nil", i)
		}
	}
}
