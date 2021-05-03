package kafka

import (
	"fmt"

	"github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/nais/liberator/pkg/namegen"
	"k8s.io/apimachinery/pkg/util/validation"
)

func GenerateKafkaSecretName(app *nais_io_v1alpha1.Application) (string, error) {
	secretName, err := namegen.ShortName(fmt.Sprintf("kafka-%s-%s", app.Name, app.Spec.Kafka.Pool), validation.DNS1035LabelMaxLength)

	if err != nil {
		return "", fmt.Errorf("unable to generate kafka secret name: %s", err)
	}
	return secretName, err
}