package main

import (
	"os"

	"github.com/bizflycloud/bizflycloud-certmanager-dns-webhook/bizflycloud"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"k8s.io/klog"
)

func main() {
	var GroupName = os.Getenv("GROUP_NAME")
	if GroupName == "" {
		klog.Fatal("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName, bizflycloud.NewSolver())
}
