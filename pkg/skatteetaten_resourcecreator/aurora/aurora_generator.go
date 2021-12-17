package aurora

import (
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	corev1 "k8s.io/api/core/v1"
)

type Source interface {
	resource.Source
	IsAuroraApplication() bool
}

func Create(app Source, ast *resource.Ast) {
	if app.IsAuroraApplication() {
		ast.Volumes = append(ast.Volumes, corev1.Volume{
			Name:         "log-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir:              &corev1.EmptyDirVolumeSource{},
			},
		})

		ast.VolumeMounts=append(ast.VolumeMounts, corev1.VolumeMount{
			Name:             "log-volume",
			MountPath:        "/u01/logs",
		})
	}

}
