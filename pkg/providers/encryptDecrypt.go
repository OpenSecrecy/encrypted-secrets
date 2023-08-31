package providers

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/shubhindia/encrypted-secrets/pkg/providers/utils"

	secretsv1alpha1 "github.com/shubhindia/encrypted-secrets/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	oldDecryptedData map[string]string
	oldEncryptedData map[string]string
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
	oldDecryptedData = decryptedMap
	oldEncryptedData = encryptedSecret.Data

	switch provider {

	case "k8s":
		k8sClient, err := utils.GetKubeClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubeclient %v", err)
		}

		namespace := encryptedSecret.Namespace

		// Retrieve the secret from the Kubernetes cluster
		secret, err := k8sClient.CoreV1().Secrets(namespace).Get(context.TODO(), "cryptctl-key", v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get the secret %v", err)
		}

		// Access the secret data
		keyPhrase := string(secret.Data["tls.crt"])

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

	case "aws-kms":
		// credentials from the shared credentials file ~/.aws/credentials.
		// ToDo: commonize the client creation code
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}
		client := kms.NewFromConfig(cfg)

		for key, value := range encryptedSecret.Data {
			ciphered, _ := base64.StdEncoding.DecodeString(value)
			decoded, err := client.Decrypt(context.TODO(), &kms.DecryptInput{
				CiphertextBlob: ciphered,
			})
			if err != nil {
				return nil, err
			}
			decryptedMap[key] = string(decoded.Plaintext)
		}
		decryptedSecret.Data = decryptedMap

		return decryptedSecret, nil

	}

	return nil, nil

}

func EncryptAndEncode(decryptedSecret secretsv1alpha1.DecryptedSecret, reEncrypt bool) (*secretsv1alpha1.EncryptedSecret, error) {

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

	case "k8s":
		k8sClient, err := utils.GetKubeClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubeclient %v", err)
		}

		namespace := decryptedSecret.Namespace

		// Retrieve the secret from the Kubernetes cluster
		secret, err := k8sClient.CoreV1().Secrets(namespace).Get(context.TODO(), "cryptctl-key", v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get the secret %v", err)
		}

		// Access the secret data
		keyPhrase := string(secret.Data["tls.crt"])

		// ToDo: plan is to eventually mode this repeated code to a separate util or something.
		// need to figure out a way to avoid this code repetition

		// check if we need to re-encrypt
		if !reEncrypt {
			for key, value := range decryptedSecret.Data {
				oldValue, exists := oldDecryptedData[key]

				if exists && value == oldValue {
					encryptedMap[key] = oldEncryptedData[key]
				} else {
					encrypted, err := staticEncryptAndEncode(value, keyPhrase)
					if err != nil || encrypted == "" {
						return nil, fmt.Errorf("failed to encrypt the data %s", err.Error())

					}
					encryptedMap[key] = encrypted
				}

			}
			encryptedSecret.Data = encryptedMap

		} else {
			for key, value := range oldDecryptedData {
				encrypted, err := staticEncryptAndEncode(value, keyPhrase)
				if err != nil || encrypted == "" {
					return nil, fmt.Errorf("failed to encrypt the data %s", err.Error())

				}
				encryptedMap[key] = encrypted
			}
			encryptedSecret.Data = encryptedMap

		}

		return encryptedSecret, nil

	case "aws-kms":
		// credentials from the shared credentials file ~/.aws/credentials.
		// ToDo: commonize the client creation code
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}
		client := kms.NewFromConfig(cfg)

		if !reEncrypt {
			for key, value := range decryptedSecret.Data {
				oldValue, exists := oldDecryptedData[key]

				if exists && value == oldValue {
					encryptedMap[key] = oldEncryptedData[key]
				} else {
					encrypted, err := client.Encrypt(context.TODO(), &kms.EncryptInput{
						KeyId:     aws.String("alias/cryptctl-key"),
						Plaintext: []byte(value),
					})
					if err != nil {
						return nil, err
					}

					encryptedMap[key] = base64.StdEncoding.EncodeToString(encrypted.CiphertextBlob)
				}

			}
			encryptedSecret.Data = encryptedMap

		} else {

			for key, value := range decryptedSecret.Data {
				encrypted, err := client.Encrypt(context.TODO(), &kms.EncryptInput{
					KeyId:     aws.String("alias/cryptctl-key"),
					Plaintext: []byte(value),
				})
				if err != nil {
					return nil, err
				}

				encryptedMap[key] = base64.StdEncoding.EncodeToString(encrypted.CiphertextBlob)
			}
			encryptedSecret.Data = encryptedMap
		}
		return encryptedSecret, nil

	}
	return nil, nil

}
