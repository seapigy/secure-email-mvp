package pqc

import (
	"crypto/rand"
	"fmt"

	"github.com/cloudflare/circl/kem/kyber/kyber1024"
	"github.com/cloudflare/circl/kem/kyber/kyber512"
	"github.com/cloudflare/circl/kem/kyber/kyber768"
	"github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cloudflare/circl/sign/dilithium/mode3"
	"github.com/cloudflare/circl/sign/dilithium/mode5"
)

// LibOQSWrapper provides a unified interface for PQC operations
type LibOQSWrapper struct {
	kemAlgorithms map[string]string
	sigAlgorithms map[string]string
}

// NewLibOQSWrapper creates a new PQC library wrapper
func NewLibOQSWrapper() (*LibOQSWrapper, error) {
	wrapper := &LibOQSWrapper{
		kemAlgorithms: make(map[string]string),
		sigAlgorithms: make(map[string]string),
	}

	// Initialize KEM algorithms
	wrapper.kemAlgorithms["kyber512"] = "kyber512"
	wrapper.kemAlgorithms["kyber768"] = "kyber768"
	wrapper.kemAlgorithms["kyber1024"] = "kyber1024"

	// Initialize signature algorithms
	wrapper.sigAlgorithms["dilithium2"] = "dilithium2"
	wrapper.sigAlgorithms["dilithium3"] = "dilithium3"
	wrapper.sigAlgorithms["dilithium5"] = "dilithium5"

	return wrapper, nil
}

// ValidateKEMAlgorithm checks if a KEM algorithm is supported
func (w *LibOQSWrapper) ValidateKEMAlgorithm(algorithm string) bool {
	_, exists := w.kemAlgorithms[algorithm]
	return exists
}

// ValidateSignatureAlgorithm checks if a signature algorithm is supported
func (w *LibOQSWrapper) ValidateSignatureAlgorithm(algorithm string) bool {
	_, exists := w.sigAlgorithms[algorithm]
	return exists
}

// GenerateKEMKeyPair generates a KEM key pair
func (w *LibOQSWrapper) GenerateKEMKeyPair(algorithm string) ([]byte, []byte, error) {
	switch algorithm {
	case "kyber512":
		pub, priv, err := kyber512.GenerateKeyPair(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Kyber512 key pair: %w", err)
		}
		pubBytes, err := pub.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Kyber512 public key: %w", err)
		}
		privBytes, err := priv.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Kyber512 private key: %w", err)
		}
		return pubBytes, privBytes, nil
	case "kyber768":
		pub, priv, err := kyber768.GenerateKeyPair(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Kyber768 key pair: %w", err)
		}
		pubBytes, err := pub.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Kyber768 public key: %w", err)
		}
		privBytes, err := priv.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Kyber768 private key: %w", err)
		}
		return pubBytes, privBytes, nil
	case "kyber1024":
		pub, priv, err := kyber1024.GenerateKeyPair(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Kyber1024 key pair: %w", err)
		}
		pubBytes, err := pub.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Kyber1024 public key: %w", err)
		}
		privBytes, err := priv.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Kyber1024 private key: %w", err)
		}
		return pubBytes, privBytes, nil
	default:
		return nil, nil, fmt.Errorf("unsupported KEM algorithm: %s", algorithm)
	}
}

// GenerateSignatureKeyPair generates a signature key pair
func (w *LibOQSWrapper) GenerateSignatureKeyPair(algorithm string) ([]byte, []byte, error) {
	switch algorithm {
	case "dilithium2":
		pub, priv, err := mode2.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Dilithium2 key pair: %w", err)
		}
		pubBytes, err := pub.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Dilithium2 public key: %w", err)
		}
		privBytes, err := priv.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Dilithium2 private key: %w", err)
		}
		return pubBytes, privBytes, nil
	case "dilithium3":
		pub, priv, err := mode3.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Dilithium3 key pair: %w", err)
		}
		pubBytes, err := pub.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Dilithium3 public key: %w", err)
		}
		privBytes, err := priv.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Dilithium3 private key: %w", err)
		}
		return pubBytes, privBytes, nil
	case "dilithium5":
		pub, priv, err := mode5.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Dilithium5 key pair: %w", err)
		}
		pubBytes, err := pub.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Dilithium5 public key: %w", err)
		}
		privBytes, err := priv.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal Dilithium5 private key: %w", err)
		}
		return pubBytes, privBytes, nil
	default:
		return nil, nil, fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}
}

// Encapsulate performs KEM encapsulation
func (w *LibOQSWrapper) Encapsulate(algorithm string, publicKeyBytes []byte) ([]byte, []byte, error) {
	switch algorithm {
	case "kyber512":
		pub := &kyber512.PublicKey{}
		pub.Unpack(publicKeyBytes)
		ciphertext := make([]byte, kyber512.CiphertextSize)
		sharedSecret := make([]byte, kyber512.SharedKeySize)
		pub.EncapsulateTo(ciphertext, sharedSecret, nil)
		return ciphertext, sharedSecret, nil
	case "kyber768":
		pub := &kyber768.PublicKey{}
		pub.Unpack(publicKeyBytes)
		ciphertext := make([]byte, kyber768.CiphertextSize)
		sharedSecret := make([]byte, kyber768.SharedKeySize)
		pub.EncapsulateTo(ciphertext, sharedSecret, nil)
		return ciphertext, sharedSecret, nil
	case "kyber1024":
		pub := &kyber1024.PublicKey{}
		pub.Unpack(publicKeyBytes)
		ciphertext := make([]byte, kyber1024.CiphertextSize)
		sharedSecret := make([]byte, kyber1024.SharedKeySize)
		pub.EncapsulateTo(ciphertext, sharedSecret, nil)
		return ciphertext, sharedSecret, nil
	default:
		return nil, nil, fmt.Errorf("unsupported KEM algorithm: %s", algorithm)
	}
}

// Decapsulate performs KEM decapsulation
func (w *LibOQSWrapper) Decapsulate(algorithm string, ciphertext []byte, privateKeyBytes []byte) ([]byte, error) {
	switch algorithm {
	case "kyber512":
		priv := &kyber512.PrivateKey{}
		priv.Unpack(privateKeyBytes)
		sharedSecret := make([]byte, kyber512.SharedKeySize)
		priv.DecapsulateTo(sharedSecret, ciphertext)
		return sharedSecret, nil
	case "kyber768":
		priv := &kyber768.PrivateKey{}
		priv.Unpack(privateKeyBytes)
		sharedSecret := make([]byte, kyber768.SharedKeySize)
		priv.DecapsulateTo(sharedSecret, ciphertext)
		return sharedSecret, nil
	case "kyber1024":
		priv := &kyber1024.PrivateKey{}
		priv.Unpack(privateKeyBytes)
		sharedSecret := make([]byte, kyber1024.SharedKeySize)
		priv.DecapsulateTo(sharedSecret, ciphertext)
		return sharedSecret, nil
	default:
		return nil, fmt.Errorf("unsupported KEM algorithm: %s", algorithm)
	}
}

// Sign performs digital signature
func (w *LibOQSWrapper) Sign(algorithm string, message []byte, privateKeyBytes []byte) ([]byte, error) {
	switch algorithm {
	case "dilithium2":
		priv := &mode2.PrivateKey{}
		if err := priv.UnmarshalBinary(privateKeyBytes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Dilithium2 private key: %w", err)
		}
		signature := make([]byte, mode2.SignatureSize)
		mode2.SignTo(priv, message, signature)
		return signature, nil
	case "dilithium3":
		priv := &mode3.PrivateKey{}
		if err := priv.UnmarshalBinary(privateKeyBytes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Dilithium3 private key: %w", err)
		}
		signature := make([]byte, mode3.SignatureSize)
		mode3.SignTo(priv, message, signature)
		return signature, nil
	case "dilithium5":
		priv := &mode5.PrivateKey{}
		if err := priv.UnmarshalBinary(privateKeyBytes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Dilithium5 private key: %w", err)
		}
		signature := make([]byte, mode5.SignatureSize)
		mode5.SignTo(priv, message, signature)
		return signature, nil
	default:
		return nil, fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}
}

// Verify performs signature verification
func (w *LibOQSWrapper) Verify(algorithm string, message []byte, signature []byte, publicKeyBytes []byte) error {
	switch algorithm {
	case "dilithium2":
		pub := &mode2.PublicKey{}
		if err := pub.UnmarshalBinary(publicKeyBytes); err != nil {
			return fmt.Errorf("failed to unmarshal Dilithium2 public key: %w", err)
		}
		if !mode2.Verify(pub, message, signature) {
			return fmt.Errorf("Dilithium2 signature verification failed")
		}
		return nil
	case "dilithium3":
		pub := &mode3.PublicKey{}
		if err := pub.UnmarshalBinary(publicKeyBytes); err != nil {
			return fmt.Errorf("failed to unmarshal Dilithium3 public key: %w", err)
		}
		if !mode3.Verify(pub, message, signature) {
			return fmt.Errorf("Dilithium3 signature verification failed")
		}
		return nil
	case "dilithium5":
		pub := &mode5.PublicKey{}
		if err := pub.UnmarshalBinary(publicKeyBytes); err != nil {
			return fmt.Errorf("failed to unmarshal Dilithium5 public key: %w", err)
		}
		if !mode5.Verify(pub, message, signature) {
			return fmt.Errorf("Dilithium5 signature verification failed")
		}
		return nil
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}
}

// GetSupportedKEMAlgorithms returns a list of supported KEM algorithms
func (w *LibOQSWrapper) GetSupportedKEMAlgorithms() []string {
	algorithms := make([]string, 0, len(w.kemAlgorithms))
	for algo := range w.kemAlgorithms {
		algorithms = append(algorithms, algo)
	}
	return algorithms
}

// GetSupportedSignatureAlgorithms returns a list of supported signature algorithms
func (w *LibOQSWrapper) GetSupportedSignatureAlgorithms() []string {
	algorithms := make([]string, 0, len(w.sigAlgorithms))
	for algo := range w.sigAlgorithms {
		algorithms = append(algorithms, algo)
	}
	return algorithms
}
