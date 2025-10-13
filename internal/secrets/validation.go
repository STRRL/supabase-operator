package secrets

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func ValidateDatabaseSecret(secret *corev1.Secret) error {
	requiredKeys := []string{"host", "port", "database", "username", "password"}

	for _, key := range requiredKeys {
		if _, ok := secret.Data[key]; !ok {
			return fmt.Errorf("missing required key '%s'", key)
		}
	}

	return nil
}

func ValidateStorageSecret(secret *corev1.Secret) error {
	requiredKeys := []string{"endpoint", "region", "bucket", "accessKeyId", "secretAccessKey"}

	for _, key := range requiredKeys {
		if _, ok := secret.Data[key]; !ok {
			return fmt.Errorf("missing required key '%s'", key)
		}
	}

	return nil
}

func ValidateBasicAuthSecret(secret *corev1.Secret) error {
	requiredKeys := []string{"username", "password"}

	for _, key := range requiredKeys {
		if _, ok := secret.Data[key]; !ok {
			return fmt.Errorf("missing required key '%s'", key)
		}
	}

	return nil
}
