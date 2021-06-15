package fixtures

import (
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Minimal naisjob returns the absolute minimum job that might live in a Kubernetes cluster.
func MinimalFailingNaisJob() *nais_io_v1.Naisjob {
	job := &nais_io_v1.Naisjob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Naisjob",
			APIVersion: nais_io_v1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ApplicationName,
			Namespace: ApplicationNamespace,
		},
	}
	err := job.ApplyDefaults()
	if err != nil {
		panic(err)
	}
	return job
}

// MinimalApplication returns the absolute minimum configuration needed to create a full set of Kubernetes resources.
func MinimalNaisJob() *nais_io_v1.Naisjob {
	job := &nais_io_v1.Naisjob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Naisjob",
			APIVersion: nais_io_v1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ApplicationName,
			Namespace: ApplicationNamespace,
			Labels: map[string]string{
				"team": ApplicationTeam,
			},
		},
		Spec: nais_io_v1.NaisjobSpec{
			Image: "example",
		},
	}
	err := job.ApplyDefaults()
	if err != nil {
		panic(err)
	}
	return job
}
