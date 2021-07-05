package main

import (
	"os"
	"testing"

	"github.com/jetstack/cert-manager/test/acme/dns"
	"github.com/bizflycloud/bizflycloud-certmanager-dns-webhook/bizflycloud"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	fixture := dns.NewFixture(bizflycloud.NewSolver(),
		dns.SetBinariesPath("_out/kubebuilder/bin"),
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/bizflycloud"),
	)

	fixture.RunConformance(t)
}
