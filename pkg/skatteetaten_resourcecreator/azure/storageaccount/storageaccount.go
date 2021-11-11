package storageaccount

import (
	"fmt"
	"strings"

	azure_microsoft_com_v1alpha1 "github.com/nais/liberator/pkg/apis/azure.microsoft.com/v1alpha1"
	skatteetaten_no_v1alpha1 "github.com/nais/liberator/pkg/apis/nebula.skatteetaten.no/v1alpha1"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	"github.com/nais/naiserator/pkg/skatteetaten_resourcecreator/istio/service_entry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type Source interface {
	resource.Source
	GetAzureResourceGroup() string
	GetStorageAccounts() map[string]*skatteetaten_no_v1alpha1.StorageAccountConfig
}

func Create(app Source, ast *resource.Ast) {
	storageAccounts := app.GetStorageAccounts()
	resourceGroup := app.GetAzureResourceGroup()
	for _, sg := range storageAccounts {
		generateStorageAccount(app, ast, resourceGroup, sg)
	}
}


func generateStorageAccount(source resource.Source, ast *resource.Ast, rg string, sg *skatteetaten_no_v1alpha1.StorageAccountConfig) {
	objectMeta := resource.CreateObjectMeta(source)
	objectMeta.Name = strings.ReplaceAll(fmt.Sprintf("sg%s%s%s", source.GetNamespace(), source.GetName(), sg.Name), "-", "")

	object := &azure_microsoft_com_v1alpha1.StorageAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageAccount",
			APIVersion: "azure.microsoft.com/v1alpha1",
		},
		ObjectMeta: objectMeta,
		Spec: azure_microsoft_com_v1alpha1.StorageAccountSpec{
			Location:               "norwayeast",
			ResourceGroup:          rg,
			Sku:                    azure_microsoft_com_v1alpha1.StorageAccountSku{
				Name: "Standard_LRS",
			},
			Kind:                   "StorageV2",
			AccessTier:             "Hot",
			EnableHTTPSTrafficOnly:  pointer.BoolPtr(true),
			//TODO SKYK-1047 og SKYK-1048
			/*NetworkRule:            &azure_microsoft_com_v1alpha1.StorageNetworkRuleSet{
				Bypass:              "AzureServices",
				VirtualNetworkRules: &[]azure_microsoft_com_v1alpha1.VirtualNetworkRule{{
					SubnetId: nil,
				}},
				DefaultAction:       "Deny",
			}, */
		},
	}

	ast.AppendOperation(resource.OperationCreateIfNotExists, object)
	envVar := createConnectionStringEnvVar(objectMeta, sg)
	ast.Env = append(ast.Env, envVar...)

	config :=skatteetaten_no_v1alpha1.ExternalEgressConfig{
		Host:  fmt.Sprintf("%s.blob.core.windows.net", objectMeta.Name),
		Ports: []skatteetaten_no_v1alpha1.PortConfig{{
			Name:     "https",
			Port:     443,
			Protocol: "TCP",
		}}}
	seName:= fmt.Sprintf("sg-%s", sg.Name)
	service_entry.GenerateServiceEntry(source, ast, seName, config)

}

func createConnectionStringEnvVar(objectMeta metav1.ObjectMeta, sg *skatteetaten_no_v1alpha1.StorageAccountConfig) []corev1.EnvVar {
	secretName := fmt.Sprintf("storageaccount-%s", objectMeta.Name)

	var envs []corev1.EnvVar
	if sg.Primary {
		envs = append(envs, corev1.EnvVar{
			Name: fmt.Sprintf("%s_CONNECTIONSTRING",strings.ToUpper(sg.Prefix)),
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "connectionString0",
				},
			},
		})

	}
	envs = append(envs, corev1.EnvVar{
		Name: fmt.Sprintf("%s_%s_CONNECTIONSTRING",strings.ToUpper(sg.Prefix), strings.ToUpper(sg.Name)),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: "connectionString0",
			},
		},
	})

	return envs
}
