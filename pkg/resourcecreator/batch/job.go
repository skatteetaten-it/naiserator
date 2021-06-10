package batch

import (
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/nais/naiserator/pkg/resourcecreator/pod"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	"github.com/nais/naiserator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateJobSpec(naisjob *nais_io_v1.Naisjob, ast *resource.Ast, resourceOptions resource.Options, batch *batchv1.Job) (batchv1.JobSpec, error) {
	podSpec, err := pod.CreateSpec(ast, resourceOptions, naisjob.GetName(), corev1.RestartPolicyNever)
	if err != nil {
		return batchv1.JobSpec{}, err
	}

	jobSpec := addJobSpec(naisjob, ast, podSpec, batch)

	return jobSpec, nil
}

func addJobSpec(naisjob *nais_io_v1.Naisjob, ast *resource.Ast, podSpec *corev1.PodSpec, batch *batchv1.Job) batchv1.JobSpec {
	jobSpec := toJobSpec(naisjob, ast, podSpec, batch)
	if naisjob.Spec.ManualSelector {
		jobSpec = addManualSelector(naisjob, jobSpec)
	}
	return jobSpec
}

func toJobSpec(naisjob *nais_io_v1.Naisjob, ast *resource.Ast, podSpec *corev1.PodSpec, batch *batchv1.Job) batchv1.JobSpec {
	naisJobObjectMeta := pod.CreateNaisjobObjectMeta(naisjob, ast)
	if batch.Name != "" {
		naisJobObjectMeta.Labels["app"] = naisjob.Name
	}

	return batchv1.JobSpec{
		ActiveDeadlineSeconds: naisjob.Spec.ActiveDeadlineSeconds,
		BackoffLimit:          util.Int32p(naisjob.Spec.BackoffLimit),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: pod.CreateNaisjobObjectMeta(naisjob, ast),
			Spec:       *podSpec,
		},
		TTLSecondsAfterFinished: naisjob.Spec.TTLSecondsAfterFinished,
	}
}

func addManualSelector(naisjob *nais_io_v1.Naisjob, jobSpec batchv1.JobSpec) batchv1.JobSpec {
	manualSelector := true
	jobSpec.ManualSelector = &manualSelector
	labels := make(map[string]string)
	labels["app"] = naisjob.GetName()
	jobSpec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	return jobSpec
}

func CreateJob(naisjob *nais_io_v1.Naisjob, ast *resource.Ast, resourceOptions resource.Options, batch *batchv1.Job) error {

	objectMeta := resource.CreateObjectMeta(naisjob)

	if val, ok := naisjob.GetAnnotations()["kubernetes.io/change-cause"]; ok {
		objectMeta.Annotations["kubernetes.io/change-cause"] = val
	}

	jobSpec, err := CreateJobSpec(naisjob, ast, resourceOptions, batch)
	if err != nil {
		return err
	}

	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: objectMeta,
		Spec:       jobSpec,
	}

	ast.AppendOperation(resource.OperationCreateOrUpdate, &job)
	return nil
}
