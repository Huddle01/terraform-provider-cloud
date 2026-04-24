package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/huddle01/terraform-provider-cloud/internal/provider"
)

// Version is the provider version reported to Terraform (provider metadata).
// Overridden at build time via: -ldflags "-X main.Version=<version>".
var Version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/huddle01/cloud",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(Version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
