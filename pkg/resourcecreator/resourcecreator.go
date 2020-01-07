// package resourcecreator converts the Kubernetes custom resource definition
// `nais.io.Applications` into standard Kubernetes resources such as Deployment,
// Service, Ingress, and so forth.

package resourcecreator

import (
	"encoding/base64"
	"fmt"
	"strings"

	nais "github.com/nais/naiserator/pkg/apis/nais.io/v1alpha1"
)

// Create takes an Application resource and returns a slice of Kubernetes resources
// along with information about what to do with these resources.
func Create(app *nais.Application, resourceOptions ResourceOptions) (ResourceOperations, error) {
	var operation Operation

	team, ok := app.Labels["team"]
	if !ok || len(team) == 0 {
		return nil, fmt.Errorf("the 'team' label needs to be set in the application metadata")
	}

	ops := ResourceOperations{
		{Service(app), OperationCreateOrUpdate},
		{ServiceAccount(app, resourceOptions), OperationCreateOrUpdate},
		{HorizontalPodAutoscaler(app), OperationCreateOrUpdate},
	}

	leRole := LeaderElectionRole(app)
	leRoleBinding := LeaderElectionRoleBinding(app)

	if app.Spec.LeaderElection {
		ops = append(ops, ResourceOperation{leRole, OperationCreateOrUpdate})
		ops = append(ops, ResourceOperation{leRoleBinding, OperationCreateOrRecreate})
	} else {
		ops = append(ops, ResourceOperation{leRole, OperationDeleteIfExists})
		ops = append(ops, ResourceOperation{leRoleBinding, OperationDeleteIfExists})
	}

	if len(resourceOptions.GoogleProjectId) > 0 {
		googleServiceAccount := GoogleServiceAccount(app)
		googleServiceAccountBinding := GoogleServiceAccountBinding(app, &googleServiceAccount, resourceOptions.GoogleProjectId)
		ops = append(ops, ResourceOperation{&googleServiceAccount, OperationCreateOrUpdate})
		ops = append(ops, ResourceOperation{&googleServiceAccountBinding, OperationCreateOrUpdate})

		if len(app.Spec.GCP.Buckets) > 0 {
			buckets := GoogleStorageBuckets(app)
			for _, bucket := range buckets {
				bucketBac := GoogleStorageBucketAccessControl(app, bucket.Name, resourceOptions.GoogleProjectId, googleServiceAccount.Name)
				ops = append(ops, ResourceOperation{bucket, OperationCreateIfNotExists})
				ops = append(ops, ResourceOperation{bucketBac, OperationCreateOrUpdate})
			}
		}

		for i, sqlInstance := range app.Spec.GCP.SqlInstances {
			if i > 0 {
				return nil, fmt.Errorf("only one sql instance is supported")
			}

			if len(sqlInstance.Name) == 0 {
				app.Spec.GCP.SqlInstances[i].Name = app.Name
				sqlInstance.Name = app.Name
			}

			instance, err := GoogleSqlInstance(app, sqlInstance)
			if err != nil {
				return nil, fmt.Errorf("unable to create sqlinstance: %s", err)
			}

			ops = append(ops, ResourceOperation{instance, OperationCreateOrUpdate})

			for _, db := range GoogleSqlDatabases(app, sqlInstance) {
				ops = append(ops, ResourceOperation{db, OperationCreateOrUpdate})
			}

			key, err := Keygen(32)
			if err != nil {
				return nil, fmt.Errorf("unable to generate secret for sql user: %s", err)
			}
			password := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(key)

			ops = append(ops, ResourceOperation{GoogleSqlUser(app, instance.Name, sqlInstance.CascadingDelete, password), OperationCreateOrUpdate})
			ops = append(ops, ResourceOperation{OpaqueSecret(app, GCPSqlInstanceSecretName(instance.Name), map[string]string{
				fmt.Sprintf("GCP_SQLINSTANCE_%s_PASSWORD", strings.ReplaceAll(strings.ToUpper(instance.Name), "-", "_")): password,
				fmt.Sprintf("GCP_SQLINSTANCE_%s_USERNAME", strings.ReplaceAll(strings.ToUpper(instance.Name), "-", "_")): instance.Name}),
				OperationCreateOrUpdate})
		}
	}

	if resourceOptions.AccessPolicy {
		ops = append(ops, ResourceOperation{NetworkPolicy(app, resourceOptions.AccessPolicyNotAllowedCIDRs), OperationCreateOrUpdate})
		vses, err := VirtualServices(app)

		if err != nil {
			return nil, fmt.Errorf("unable to create VirtualServices: %s", err)
		}

		operation = OperationCreateOrUpdate
		if len(app.Spec.Ingresses) == 0 {
			operation = OperationDeleteIfExists
		}

		for _, vs := range vses {
			ops = append(ops, ResourceOperation{vs, operation})
		}

		// Applies to ServiceRoles and ServiceRoleBindings
		operation = OperationCreateOrUpdate
		if len(app.Spec.AccessPolicy.Inbound.Rules) == 0 && len(app.Spec.Ingresses) == 0 {
			operation = OperationDeleteIfExists
		}

		serviceRole := ServiceRole(app)
		if serviceRole != nil {
			ops = append(ops, ResourceOperation{serviceRole, operation})
		}

		serviceRoleBinding := ServiceRoleBinding(app)
		if serviceRoleBinding != nil {
			ops = append(ops, ResourceOperation{serviceRoleBinding, operation})
		}

		serviceRolePrometheus := ServiceRolePrometheus(app)
		if serviceRolePrometheus != nil {
			ops = append(ops, ResourceOperation{serviceRolePrometheus, OperationCreateOrUpdate})
		}

		serviceRoleBindingPrometheus := ServiceRoleBindingPrometheus(app)
		operation = OperationCreateOrUpdate
		if !app.Spec.Prometheus.Enabled {
			operation = OperationDeleteIfExists
		}

		if serviceRoleBindingPrometheus != nil {
			ops = append(ops, ResourceOperation{serviceRoleBindingPrometheus, operation})
		}

		serviceEntry := ServiceEntry(app)
		operation = OperationCreateOrUpdate
		if len(app.Spec.AccessPolicy.Outbound.External) == 0 {
			operation = OperationDeleteIfExists
		}
		if serviceEntry != nil {
			ops = append(ops, ResourceOperation{serviceEntry, operation})
		}

	} else {

		ingress, err := Ingress(app)
		if err != nil {
			return nil, fmt.Errorf("while creating ingress: %s", err)
		}

		// Kubernetes doesn't support ingress resources without any rules. This means we must
		// delete the old resource if it exists.
		operation = OperationCreateOrUpdate
		if len(app.Spec.Ingresses) == 0 {
			operation = OperationDeleteIfExists
		}

		ops = append(ops, ResourceOperation{ingress, operation})
	}

	deployment, err := Deployment(app, resourceOptions) // TODO ...resourceOptions, additionalEnvs) //
	if err != nil {
		return nil, fmt.Errorf("while creating deployment: %s", err)
	}
	ops = append(ops, ResourceOperation{deployment, OperationCreateOrUpdate})

	return ops, nil
}

func int32p(i int32) *int32 {
	return &i
}

func CascadingDeleteAnnotation(cascadingDelete bool) map[string]string {
	if cascadingDelete {
		return nil
	}

	return map[string]string{"cnrm.cloud.google.com/deletion-policy": "abandon"}
}
