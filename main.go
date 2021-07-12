// This file is part of bizfly-certmanager-dns-webhook
//
// Copyright (C) 2021  BizFly Cloud
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>

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
