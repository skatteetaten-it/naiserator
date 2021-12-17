package service_entry

import (
	"fmt"

	skatteetaten_no_v1alpha1 "github.com/nais/liberator/pkg/apis/nebula.skatteetaten.no/v1alpha1"
	networking_istio_io_v1alpha3 "github.com/nais/liberator/pkg/apis/networking.istio.io/v1alpha3"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Source interface {
	resource.Source
	GetEgress() *skatteetaten_no_v1alpha1.EgressConfig
}

func Create(app Source, ast *resource.Ast) {
	egressConfig := app.GetEgress()

	// ServiceEntry
	if egressConfig != nil && egressConfig.External != nil {
		for key, egress := range egressConfig.External {
			GenerateServiceEntry(app, ast, key, egress)
		}
	}
}

func GenerateServiceEntry(source resource.Source, ast *resource.Ast, key string, config skatteetaten_no_v1alpha1.ExternalEgressConfig){

	serviceentry := networking_istio_io_v1alpha3.ServiceEntry{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceEntry",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: resource.CreateObjectMeta(source),
		Spec:       networking_istio_io_v1alpha3.ServiceEntrySpec{
			Hosts:      []string{config.Host},
			Location:   "MESH_EXTERNAL",
			Resolution: "DNS",
			Ports:      generateServiceEntryPorts(config),
		},
	}

	serviceentry.ObjectMeta.Name = fmt.Sprintf("%s-%s-%s", source.GetNamespace(), source.GetName(), key)

	ast.AppendOperation(resource.OperationCreateOrUpdate, &serviceentry)
}

func generateServiceEntryPorts(config skatteetaten_no_v1alpha1.ExternalEgressConfig) []networking_istio_io_v1alpha3.Port {
	ports := []networking_istio_io_v1alpha3.Port{}
	for _, port := range config.Ports {
		ports = append(ports, networking_istio_io_v1alpha3.Port{
			Number:   uint32(port.Port),
			Protocol: port.Protocol,
			Name:     port.Name,
		})
	}

	return ports
}
