package storageaccount

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
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

type ResourceName struct {
	name string
	azureName string
}

func Create(app Source, ast *resource.Ast, options resource.Options) {
	storageAccounts := app.GetStorageAccounts()
	resourceGroup := app.GetAzureResourceGroup()
	subscription := options.AzureSubscriptionName
	for _, sg := range storageAccounts {
		generateStorageAccount(app, ast, resourceGroup, sg, subscription)
	}
}


func generateStorageAccount(source resource.Source, ast *resource.Ast, rg string, sg *skatteetaten_no_v1alpha1.StorageAccountConfig, subscription string) {
	objectMeta := resource.CreateObjectMeta(source)
	// TODO: With ASO v2 change this to
	//   objectMeta.Name = resourceName.Name
	//   objectMeta.AzureName = resourceName.azureName
	resourceName := generateName(subscription, source.GetNamespace(), source.GetName(), sg.Name)
	objectMeta.Name = resourceName.azureName

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
	envVar := createConnectionStringEnvVar(sg, resourceName.azureName)
	ast.Env = append(ast.Env, envVar...)

	config :=skatteetaten_no_v1alpha1.ExternalEgressConfig{
		Host:  fmt.Sprintf("%s.blob.core.windows.net", resourceName.azureName),
		Ports: []skatteetaten_no_v1alpha1.PortConfig{{
			Name:     "https",
			Port:     443,
			Protocol: "TCP",
		}}}
	seName:= fmt.Sprintf("sg-%s", sg.Name)
	service_entry.GenerateServiceEntry(source, ast, seName, config)
}

func createConnectionStringEnvVar(sg *skatteetaten_no_v1alpha1.StorageAccountConfig, azureName string) []corev1.EnvVar {
	secretName := fmt.Sprintf("storageaccount-%s", azureName)

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

func generateName(subscription string, namespace string, appName string, sgName string) ResourceName {
	// Storage account names must be between 3 and 24 characters in length and may contain numbers
	// and lowercase letters only. The storage account name must be unique within Azure.
	k8sName := fmt.Sprintf("%s-%s", appName, sgName)

	// Generate SHA1 from full name and extract the first 7 chars
	fullName := fmt.Sprintf("%s-%s-%s", subscription, namespace, k8sName)
	h := sha1.New()
	h.Write([]byte(fullName))

	sha1String := hex.EncodeToString(h.Sum(nil))
	sha1String = sha1String[0:7]

	// Filter all non-alphanumeric chars and leave max 14 chars from k8s name as prefix
	// to azure name. Full name is only used for generating the hash.
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	name := []rune(reg.ReplaceAllString(k8sName, ""))

	// azureName: sg<k8s name><hash of full name>
	azureName := fmt.Sprintf("sg%s%s", string(name[0:14]), sha1String)

	return ResourceName{
		name: k8sName,
		azureName: azureName,
	}
}
