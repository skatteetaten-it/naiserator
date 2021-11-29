package resourceGroup

import (
	"fmt"

	azure_microsoft_com_v1alpha1 "github.com/nais/liberator/pkg/apis/azure.microsoft.com/v1alpha1"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Source interface {
	resource.Source
	GetAzureResourceGroup() string
}


func GenerateResourceGroupName(app Source) string {
	resourceGroup := app.GetAzureResourceGroup()
	return fmt.Sprintf("rg-%s-%s", app.GetNamespace(), resourceGroup)
}

/*
   A resource group is a shared resource that the first application in a namespace will provsion, it will not be deleted if an
   application is deleted
 */
func Create(app Source, ast *resource.Ast) {
	objectMeta := resource.CreateObjectMeta(app)
	objectMeta.OwnerReferences=[]metav1.OwnerReference{} //we remove owner reference, this resource can be shared
	objectMeta.Name = GenerateResourceGroupName(app)


	rg := &azure_microsoft_com_v1alpha1.ResourceGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ResourceGroup",
			APIVersion: "azure.microsoft.com/v1alpha1",
		},
		ObjectMeta: objectMeta,
		Spec: azure_microsoft_com_v1alpha1.ResourceGroupSpec{
			Location: "norwayeast",
		},
	}
	ast.AppendOperation(resource.OperationCreateIfNotExists, rg)
}

