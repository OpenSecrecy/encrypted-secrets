package providers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/shubhindia/encrypted-secrets/pkg/providers/utils"
)

func staticDecodeAndDecrypt(encoded string, keyPhrase string) (string, error) {
	ciphered, _ := base64.StdEncoding.DecodeString(encoded)
	hashedPhrase := utils.MdHashing(keyPhrase)

	aesBlock, err := aes.NewCipher([]byte(hashedPhrase))
	if err != nil {
		return "", err
	}

	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", err
	}

	nonceSize := gcmInstance.NonceSize()
	nonce, cipheredText := ciphered[:nonceSize], ciphered[nonceSize:]
	originalText, err := gcmInstance.Open(nil, nonce, cipheredText, nil)
	if err != nil {
		return "", err
	}

	return string(originalText), nil

}

func staticEncryptAndEncode(value string, keyPhrase string) (string, error) {

	aesBlock, err := aes.NewCipher([]byte(utils.MdHashing(keyPhrase)))
	if err != nil {
		return "", err
	}

	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcmInstance.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)

	cipheredText := gcmInstance.Seal(nonce, nonce, []byte(value), nil)

	encoded := base64.StdEncoding.EncodeToString(cipheredText)

	return encoded, nil
}
