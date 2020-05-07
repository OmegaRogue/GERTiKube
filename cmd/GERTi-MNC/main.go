package main

import (
	"context"
	"log"

	"github.com/OmegaRogue/gerte-go"
	"k8s.io/client-go/kubernetes"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

var Dns DnsNetworkConfig

var Key string

var Clientset kubernetes.Clientset

var GERTe gerte.GertAddress

func main() {
	var err error
	Clientset, err = ClusterSetup()

	Key, err = ParseNetworkKey(context.TODO())
	if err != nil {
		log.Fatalf("error on ParseNetworkKey: %+v", err)
	}
	log.Printf("key: %v\n", Key)

	Dns, err = ParseNetworkConf(context.TODO())
	log.Printf("GERTe: %+v\n", GERTe)
	if err != nil {
		log.Fatalf("error on ParseNetworkConf: %+v", err)
	}
	log.Println("old:")

	for _, service := range Dns.Services {
		log.Printf("%v: %+v\n", service.Name, service.Address)
	}
	Dns.Services[0].Address.Upper += 1

	log.Println("new:")
	for _, service := range Dns.Services {
		log.Printf("%v: %+v\n", service.Name, service.Address)
	}

	err = SaveNetworkConfig(context.TODO(), Dns)
	if err != nil {
		log.Fatalf("error on SaveNetworkConfig: %+v", err)
	}

}
