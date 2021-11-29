package cosmosdb

import (
	"fmt"
	"strings"

	azure_microsoft_com_v1alpha1 "github.com/nais/liberator/pkg/apis/azure.microsoft.com/v1alpha1"
	skatteetaten_no_v1alpha1 "github.com/nais/liberator/pkg/apis/nebula.skatteetaten.no/v1alpha1"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	"github.com/nais/naiserator/pkg/skatteetaten_resourcecreator/azure/resourceGroup"
	"github.com/nais/naiserator/pkg/skatteetaten_resourcecreator/istio/service_entry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type Source interface {
	resource.Source
	GetAzureResourceGroup() string
	GetCosmosDb() map[string]*skatteetaten_no_v1alpha1.CosmosDBConfig
}

func Create(app Source, ast *resource.Ast) {
	cosmosDb := app.GetCosmosDb()
	resourceGroup := resourceGroup.GenerateResourceGroupName(app)

	for _, db := range cosmosDb {
		generateCosmosDb(app, ast, resourceGroup, db)
	}
}

func generateCosmosDb(source resource.Source, ast *resource.Ast, rg string, db *skatteetaten_no_v1alpha1.CosmosDBConfig) {
	objectMeta := resource.CreateObjectMeta(source)
	objectMeta.Name = fmt.Sprintf("cod-%s-%s-%s", source.GetNamespace(), source.GetName(), db.Name)

	spec := azure_microsoft_com_v1alpha1.CosmosDBSpec{
		Location:      "norwayeast",
		ResourceGroup: rg,
		Properties: azure_microsoft_com_v1alpha1.CosmosDBProperties{
			DatabaseAccountOfferType: "Standard",
		},
	}
	if db.MongoDBVersion != "" {
		spec.Kind = "MongoDB"
		spec.Properties.MongoDBVersion = db.MongoDBVersion
		spec.Properties.Capabilities = &[]azure_microsoft_com_v1alpha1.Capability{{
			Name: pointer.StringPtr("EnableMongo"),
		}}
	} else {
		spec.Kind = "GlobalDocumentDB"
	}

	object := &azure_microsoft_com_v1alpha1.CosmosDB{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CosmosDB",
			APIVersion: "azure.microsoft.com/v1alpha1",
		},
		ObjectMeta: objectMeta,
		Spec:       spec,
	}

	ast.AppendOperation(resource.OperationCreateIfNotExists, object)
	envVar := createConnectionStringEnvVar(objectMeta, db)
	ast.Env = append(ast.Env, envVar...)

	config :=skatteetaten_no_v1alpha1.ExternalEgressConfig{
		Host:  fmt.Sprintf("%s.mongo.cosmos.azure.com", objectMeta.Name),
		Ports: []skatteetaten_no_v1alpha1.PortConfig{{
			Name:     "mongodb",
			Port:     10255,
			Protocol: "TCP",
		}}}
	seName:= fmt.Sprintf("cod-%s", db.Name)
	service_entry.GenerateServiceEntry(source, ast, seName, config)
}


func createConnectionStringEnvVar(objectMeta metav1.ObjectMeta, db *skatteetaten_no_v1alpha1.CosmosDBConfig) []corev1.EnvVar {
	secretName := fmt.Sprintf("cosmosdb-%s", objectMeta.Name)

	var envs []corev1.EnvVar
	if db.Primary {

		envs = append(envs, corev1.EnvVar{
			Name: strings.ToUpper(fmt.Sprintf("%s_URI", db.Prefix)),
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "PrimaryMongoDBConnectionString",
				},
			},
		})

		envs = append(envs, corev1.EnvVar{
			Name:  strings.ToUpper(fmt.Sprintf("%s_DATABASE", db.Prefix)),
			Value: db.Name,
		})
	}

	//TODO: not really sure what to do here if this is not mongodb
	envs = append(envs, corev1.EnvVar{
		Name: strings.ToUpper(fmt.Sprintf("%s_%s_URI", db.Prefix, db.Name)),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: "PrimaryMongoDBConnectionString",
			},
		},
	})

	envs = append(envs, corev1.EnvVar{
		Name:  strings.ToUpper(fmt.Sprintf("%s_%s_DATABASE", db.Prefix, db.Name)),
		Value: db.Name,
	})
	return envs

}
