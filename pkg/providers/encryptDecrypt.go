package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/shubhindia/encrypted-secrets/pkg/providers/utils"

	secretsv1alpha1 "github.com/shubhindia/encrypted-secrets/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DecodeAndDecrypt(encryptedSecret *secretsv1alpha1.EncryptedSecret) (*secretsv1alpha1.DecryptedSecret, error) {

	// get the provider
	provider := encryptedSecret.GetAnnotations()["secrets.shubhindia.xyz/provider"]

	// init a decryptedSecret to hold everything
	decryptedSecret := &secretsv1alpha1.DecryptedSecret{
		ObjectMeta: encryptedSecret.ObjectMeta,
		TypeMeta: v1.TypeMeta{
			Kind:       "DecryptedSecret",
			APIVersion: "secrets.shubhindia.xyz/v1alpha1",
		},
	}

	// map to hold the decrypted values
	decryptedMap := make(map[string]string)

	switch provider {
	case "static":
		keyPhrase := os.Getenv("KEYPHRASE")
		if keyPhrase == "" {
			return nil, fmt.Errorf("keyphrase not found")
		}

		for key, value := range encryptedSecret.Data {
			decoded, err := staticDecodeAndDecrypt(value, keyPhrase)
			if err != nil {
				return nil, err
			}
			decryptedMap[key] = decoded
		}

		// add the decrypted values to decryptedSecret
		decryptedSecret.Data = decryptedMap

		return decryptedSecret, nil

	case "k8s":
		k8sClient, err := utils.GetKubeClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubeclient %v", err)
		}

		// define the namespace and secret name to retrieve
		// for now I am keeping this to default since the default service account exists for all the namespaces
		// due to which we don't have to create anyhing apart from the secret for the SA using the provided yaml in the docs
		secretName := "default"
		namespace := encryptedSecret.Namespace

		// Retrieve the secret from the Kubernetes cluster
		secret, err := k8sClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get the secret %v", err)
		}

		// Access the secret data"
		keyPhrase := string(secret.Data["token"])

		// ToDo: plan is to eventually mode this repeated code to a separate util or something.
		// need to figure out a way to avoid this code repetition
		for key, value := range encryptedSecret.Data {
			decoded, err := staticDecodeAndDecrypt(value, keyPhrase)
			if err != nil {
				return nil, err
			}
			decryptedMap[key] = decoded
		}

		// add the decrypted values to decryptedSecret
		decryptedSecret.Data = decryptedMap

		return decryptedSecret, nil

	}

	return nil, nil

}

func EncryptAndEncode(decryptedSecret secretsv1alpha1.DecryptedSecret) (*secretsv1alpha1.EncryptedSecret, error) {

	// get the provider
	provider := decryptedSecret.GetAnnotations()["secrets.shubhindia.xyz/provider"]

	// init a encryptedSecret to hold everything
	encryptedSecret := &secretsv1alpha1.EncryptedSecret{
		ObjectMeta: decryptedSecret.ObjectMeta,
		TypeMeta: v1.TypeMeta{
			Kind:       "EncryptedSecret",
			APIVersion: "secrets.shubhindia.xyz/v1alpha1",
		},
	}

	// map to hold the encrypted values
	encryptedMap := make(map[string]string)

	switch provider {
	case "static":
		keyPhrase := os.Getenv("KEYPHRASE")
		if keyPhrase == "" {
			return nil, fmt.Errorf("keyphrase not found")
		}
		for key, value := range decryptedSecret.Data {
			encrypted, err := staticEncryptAndEncode(value, keyPhrase)
			if err != nil || encrypted == "" {
				return nil, fmt.Errorf("failed to encrypt the data %s", err.Error())

			}
			encryptedMap[key] = encrypted
		}
		encryptedSecret.Data = encryptedMap
		return encryptedSecret, nil

	case "k8s":
		k8sClient, err := utils.GetKubeClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubeclient %v", err)
		}

		// define the namespace and secret name to retrieve
		// for now I am keeping this to default since the default service account exists for all the namespaces
		// due to which we don't have to create anyhing apart from the secret for the SA using the provided yaml in the docs
		secretName := "default"
		namespace := decryptedSecret.Namespace

		// Retrieve the secret from the Kubernetes cluster
		secret, err := k8sClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get the secret %v", err)
		}

		// Access the secret data"
		keyPhrase := string(secret.Data["token"])

		// ToDo: plan is to eventually mode this repeated code to a separate util or something.
		// need to figure out a way to avoid this code repetition
		for key, value := range decryptedSecret.Data {
			encrypted, err := staticEncryptAndEncode(value, keyPhrase)
			if err != nil || encrypted == "" {
				return nil, fmt.Errorf("failed to encrypt the data %s", err.Error())

			}
			encryptedMap[key] = encrypted
		}
		encryptedSecret.Data = encryptedMap
		return encryptedSecret, nil

	}
	return nil, nil

}
