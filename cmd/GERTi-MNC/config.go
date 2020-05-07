package main

import (
	"context"
	"fmt"

	"github.com/OmegaRogue/gerte-go"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceConf struct {
	Name    string `yaml:",flow"`
	Address string `yaml:"address"`
}

type NetworkConfig struct {
	Services []ServiceConf
}

type DnsServiceConf struct {
	Name    string
	Uid     types.UID
	Address gerte.GertAddress
	Service *v1.Service
}

type DnsNetworkConfig struct {
	Services []DnsServiceConf
}

func (dns DnsNetworkConfig) ToNetworkConfig() NetworkConfig {
	conf := NetworkConfig{}
	conf.Services = make([]ServiceConf, len(dns.Services))
	for i, service := range dns.Services {
		conf.Services[i] = ServiceConf{
			Name:    service.Name,
			Address: service.Address.PrintAddress(),
		}
	}
	return conf
}

func (net NetworkConfig) ToDnsNetworkConfig(ctx context.Context) (DnsNetworkConfig, error) {
	dns := DnsNetworkConfig{}
	dns.Services = make([]DnsServiceConf, len(net.Services))
	for i, service := range net.Services {
		svc, err := Clientset.CoreV1().Services("gert").Get(ctx, service.Name, metav1.GetOptions{})
		if err != nil {
			return DnsNetworkConfig{}, fmt.Errorf("parseNetworkConf errored on get service: %w", err)
		}

		addr, err := gerte.AddressFromString(service.Address)
		if err != nil {
			return DnsNetworkConfig{}, fmt.Errorf("parseNetworkConf errored on parse address: %w", err)
		}

		dns.Services[i] = DnsServiceConf{
			Name:    service.Name,
			Uid:     svc.UID,
			Address: addr,
			Service: svc,
		}
	}
	return dns, nil
}

func ParseNetworkKey(ctx context.Context) (string, error) {
	key, err := Clientset.CoreV1().Secrets("gert").Get(ctx, "gerte-key", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("parseNetworkKey errored on get secret: %w", err)
	}

	return string(key.Data["key"]), nil
}

func ParseNetworkConf(ctx context.Context) (DnsNetworkConfig, error) {
	con, err := Clientset.CoreV1().ConfigMaps("gert").Get(ctx, "addresses", metav1.GetOptions{})
	if err != nil {
		return DnsNetworkConfig{}, fmt.Errorf("parseNetworkConf errored on get config: %w", err)
	}

	GERTe, err = gerte.AddressFromString(con.Data["gerte"])
	if err != nil {
		return DnsNetworkConfig{}, fmt.Errorf("parseNetworkConf errored on parse GERTe Address: %w", err)
	}

	conf := NetworkConfig{}

	err = yaml.Unmarshal([]byte(con.Data["addresses"]), &conf)
	if err != nil {
		return DnsNetworkConfig{}, fmt.Errorf("parseNetworkConf errored on unmarshal: %w", err)
	}

	dns, err := conf.ToDnsNetworkConfig(ctx)
	if err != nil {
		return DnsNetworkConfig{}, fmt.Errorf("parseNetworkConf errored on convert to DNS Config: %w", err)
	}
	return dns, nil

}

func SaveNetworkConfig(ctx context.Context, config DnsNetworkConfig) error {
	con, err := Clientset.CoreV1().ConfigMaps("gert").Get(ctx, "addresses", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("saveNetworkConfig errored on get config: %w", err)
	}
	conf := config.ToNetworkConfig()
	con.Data["gerte"] = GERTe.PrintAddress()
	data, err := yaml.Marshal(&conf)
	if err != nil {
		return fmt.Errorf("saveNetworkConfig errored on marshal: %w", err)
	}
	con.Data["addresses"] = string(data)
	_, err = Clientset.CoreV1().ConfigMaps("gert").Update(ctx, con, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("saveNetworkConfig errored on update config")
	}

	return nil
}
