package crypto

type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type NoOpEncryptor struct{}

func (e *NoOpEncryptor) Encrypt(plaintext string) (string, error)  { return plaintext, nil }
func (e *NoOpEncryptor) Decrypt(ciphertext string) (string, error) { return ciphertext, nil }
