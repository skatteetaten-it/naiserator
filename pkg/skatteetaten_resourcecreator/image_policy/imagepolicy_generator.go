package image_policy

import (
	"fmt"
	"strings"
	"time"

	fluxcd_io_image_reflector_v1beta1 "github.com/nais/liberator/pkg/apis/fluxcd.io/image-reflector/v1beta1"
	skatteetaten_no_v1alpha1 "github.com/nais/liberator/pkg/apis/nebula.skatteetaten.no/v1alpha1"
	"github.com/nais/naiserator/pkg/resourcecreator/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Source interface {
	resource.Source
	GetImagePolicy() *skatteetaten_no_v1alpha1.ImagePolicyConfig
	GetImageName() string
}

func Create(app Source, ast *resource.Ast) error {
	imagePolicy := app.GetImagePolicy()

	if imagePolicy == nil  {
		return nil
	}


	//TODO: verifisere dette
	tenantNamespace :=  strings.Split(app.GetNamespace(), "-")[0]

	hasBranch := imagePolicy.Branch != ""
	hasVersion := imagePolicy.Semver != ""

	if hasBranch && hasVersion {
		return fmt.Errorf("specify either version or branch, not both")
	}

	if !hasBranch && !hasVersion  {
		return fmt.Errorf("invalid specification, specify either branchName or semVer range or disable imagePolicy")
	}

	var tags *fluxcd_io_image_reflector_v1beta1.TagFilter
	var choice fluxcd_io_image_reflector_v1beta1.ImagePolicyChoice

	if imagePolicy.Branch != "" {
		choice = fluxcd_io_image_reflector_v1beta1.ImagePolicyChoice{
			Numerical: &fluxcd_io_image_reflector_v1beta1.NumericalPolicy{
				Order: "asc",
			},
		}
		//TODO: validate branch name?
		tags = &fluxcd_io_image_reflector_v1beta1.TagFilter{
			Extract: "$time",
			Pattern: fmt.Sprintf(`^%s-([0-9a-z]+)-(?P<time>[0-9]+)`, imagePolicy.Branch),
		}
	} else if imagePolicy.Semver != "" {
		choice = fluxcd_io_image_reflector_v1beta1.ImagePolicyChoice{
			//TODO: validate semver range?
			SemVer: &fluxcd_io_image_reflector_v1beta1.SemVerPolicy{
				Range: imagePolicy.Semver,
			},
		}
	}

	imagePolicyName := fmt.Sprintf("%s-%s", app.GetName(), imagePolicy.NameSuffix)

	policy := &fluxcd_io_image_reflector_v1beta1.ImagePolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "image.toolkit.fluxcd.io/v1beta1",
			Kind:       "ImagePolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      imagePolicyName,
			Namespace: tenantNamespace,
		},
		Spec: fluxcd_io_image_reflector_v1beta1.ImagePolicySpec{
			ImageRepositoryRef: fluxcd_io_image_reflector_v1beta1.LocalObjectReference{Name: app.GetName()},
			Policy:             choice,
			FilterTags:         tags,
		},
	}
	ast.AppendOperation(resource.OperationCreateIfNotExists, policy)

	repository := &fluxcd_io_image_reflector_v1beta1.ImageRepository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "image.toolkit.fluxcd.io/v1beta1",
			Kind:       "ImageRepository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.GetName(),
			Namespace: tenantNamespace,
		},
		Spec: fluxcd_io_image_reflector_v1beta1.ImageRepositorySpec{
			Image:         app.GetImageName(),
			Interval:      metav1.Duration{
				Duration: 1 * time.Minute,
			},
			SecretRef:     &fluxcd_io_image_reflector_v1beta1.LocalObjectReference{
				Name: "gh-docker-credentials",
			},
		},
	}
	ast.AppendOperation(resource.OperationCreateIfNotExists, repository)
	return nil
}
