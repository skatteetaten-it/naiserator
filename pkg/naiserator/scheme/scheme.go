package naiserator_scheme

import (
	iam_cnrm_cloud_google_com_v1beta1 "github.com/nais/liberator/pkg/apis/iam.cnrm.cloud.google.com/v1beta1"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	networking_istio_io_v1alpha3 "github.com/nais/liberator/pkg/apis/networking.istio.io/v1alpha3"
	sql_cnrm_cloud_google_com_v1beta1 "github.com/nais/liberator/pkg/apis/sql.cnrm.cloud.google.com/v1beta1"
	storage_cnrm_cloud_google_com_v1beta1 "github.com/nais/liberator/pkg/apis/storage.cnrm.cloud.google.com/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func Scheme(schemes ...func(*runtime.Scheme) error) (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	for _, fn := range schemes {
		err := fn(scheme)
		if err != nil {
			return nil, err
		}
	}
	return scheme, nil
}

func All() (*runtime.Scheme, error) {
	return Scheme(
		nais_io_v1alpha1.AddToScheme,
		nais_io_v1.AddToScheme,
		iam_cnrm_cloud_google_com_v1beta1.AddToScheme,
		sql_cnrm_cloud_google_com_v1beta1.AddToScheme,
		storage_cnrm_cloud_google_com_v1beta1.AddToScheme,
		networking_istio_io_v1alpha3.AddToScheme,
		clientgoscheme.AddToScheme,
	)
}