package getproviders

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform/addrs"
)

func TestFilesystemMirrorSourceAllAvailablePackages(t *testing.T) {
	source := NewFilesystemMirrorSource("testdata/filesystem-mirror")
	got, err := source.AllAvailablePackages()
	if err != nil {
		t.Fatal(err)
	}

	want := map[addrs.Provider][]PackageMeta{
		nullProvider:       {},
		randomProvider:     {},
		happycloudProvider: {},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("incorrect result\n%s", diff)
	}
}

var nullProvider = addrs.Provider{
	Hostname:  svchost.Hostname("registry.terraform.io"),
	Namespace: "hashicorp",
	Type:      "null",
}
var randomProvider = addrs.Provider{
	Hostname:  svchost.Hostname("registry.terraform.io"),
	Namespace: "hashicorp",
	Type:      "random",
}
var happycloudProvider = addrs.Provider{
	Hostname:  svchost.Hostname("tfe.example.com"),
	Namespace: "awesomecorp",
	Type:      "happycloud",
}
