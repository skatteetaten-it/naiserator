package postgres

import (
	"fmt"

	azure_microsoft_com_v1alpha1 "github.com/nais/liberator/pkg/apis/azure.microsoft.com/v1alpha1"
	skatteetaten_no_v1alpha1 "github.com/nais/liberator/pkg/apis/nebula.skatteetaten.no/v1alpha1"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	"github.com/nais/naiserator/pkg/skatteetaten_resourcecreator/istio/service_entry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


type Source interface {
	resource.Source
	GetAzureResourceGroup() string
	GetPostgresDatabases() map[string]*skatteetaten_no_v1alpha1.PostgreDatabaseConfig
}

func Create(app Source, ast *resource.Ast) {

	pgd := app.GetPostgresDatabases()
	resourceGroup := app.GetAzureResourceGroup()
	springDataSourceCreated := false
	serverNames := map[string]skatteetaten_no_v1alpha1.ExternalEgressConfig {}
	for _, db := range pgd {
		generatePostgresDatabase(app, ast, resourceGroup, *db)
		for _, user := range db.Users {
			if !springDataSourceCreated {
				secretName := fmt.Sprintf("postgresqluser-pgu-%s-%s", app.GetName(), user.Name)
				dbVars := GenerateDbEnv(user, secretName)
				ast.Env = append(ast.Env, dbVars...)
				springDataSourceCreated=true
			}
			generatePostgresUser(app, ast, resourceGroup, *db, *user)
		}

		_, ok := serverNames[db.Server]
		if !ok {
			serverNames[db.Server] =skatteetaten_no_v1alpha1.ExternalEgressConfig{
				Host:  fmt.Sprintf("%s.postgres.database.azure.com", generateDatabaseServerName(app, db.Server)),
				Ports: []skatteetaten_no_v1alpha1.PortConfig{{
					Name:     "postgres",
					Port:     5432,
					Protocol: "TCP",
				}},
			}
		}
	}

	for name, config := range serverNames {
		service_entry.GenerateServiceEntry(app, ast, fmt.Sprintf("pgs-%s", name), config)
	}
}

func generatePostgresDatabase(source resource.Source, ast *resource.Ast, rg string, database skatteetaten_no_v1alpha1.PostgreDatabaseConfig) {
	objectMeta := resource.CreateObjectMeta(source)
	objectMeta.Name = fmt.Sprintf("pgd-%s-%s-%s", source.GetNamespace(), source.GetName(), database.Name)

	db := &azure_microsoft_com_v1alpha1.PostgreSQLDatabase{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PostgreSQLDatabase",
			APIVersion: "azure.microsoft.com/v1alpha1",
		},
		ObjectMeta: objectMeta,
		Spec: azure_microsoft_com_v1alpha1.PostgreSQLDatabaseSpec{
			ResourceGroup: rg,
			Server:        generateDatabaseServerName(source, database.Server),
		},
	}
	ast.AppendOperation(resource.OperationCreateIfNotExists, db)

}

func generateDatabaseServerName(source resource.Source, server string) string {
	return fmt.Sprintf("pgs-%s-%s", source.GetNamespace(), server)
}

func generatePostgresUser(source resource.Source, ast *resource.Ast, rg string, database skatteetaten_no_v1alpha1.PostgreDatabaseConfig, user skatteetaten_no_v1alpha1.PostgreDatabaseUser) {

	objectMeta := resource.CreateObjectMeta(source)
	objectMeta.Name = fmt.Sprintf("pgu-%s-%s", source.GetName(), user.Name)

	pgu := &azure_microsoft_com_v1alpha1.PostgreSQLUser{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PostgreSQLUser",
			APIVersion: "azure.microsoft.com/v1alpha1",
		},
		ObjectMeta: objectMeta,
		Spec: azure_microsoft_com_v1alpha1.PostgreSQLUserSpec{
			DbName:        fmt.Sprintf("pgd-%s-%s-%s", source.GetNamespace(), source.GetName(), database.Name),
			ResourceGroup: rg,
			Server:        fmt.Sprintf("pgs-%s-%s", source.GetNamespace(), database.Server),
			Roles:         []string{user.Role},
		},
	}

	ast.AppendOperation(resource.OperationCreateIfNotExists, pgu)
}
