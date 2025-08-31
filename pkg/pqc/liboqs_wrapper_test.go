package pqc

import (
	"testing"
)

func TestNewLibOQSWrapper(t *testing.T) {
	wrapper, err := NewLibOQSWrapper()
	if err != nil {
		t.Fatalf("Failed to create LibOQSWrapper: %v", err)
	}

	if wrapper == nil {
		t.Fatal("Wrapper is nil")
	}

	// Test KEM algorithm validation
	if !wrapper.ValidateKEMAlgorithm("kyber512") {
		t.Error("kyber512 should be a valid KEM algorithm")
	}
	if !wrapper.ValidateKEMAlgorithm("kyber768") {
		t.Error("kyber768 should be a valid KEM algorithm")
	}
	if !wrapper.ValidateKEMAlgorithm("kyber1024") {
		t.Error("kyber1024 should be a valid KEM algorithm")
	}
	if wrapper.ValidateKEMAlgorithm("invalid") {
		t.Error("invalid should not be a valid KEM algorithm")
	}

	// Test signature algorithm validation
	if !wrapper.ValidateSignatureAlgorithm("dilithium2") {
		t.Error("dilithium2 should be a valid signature algorithm")
	}
	if !wrapper.ValidateSignatureAlgorithm("dilithium3") {
		t.Error("dilithium3 should be a valid signature algorithm")
	}
	if !wrapper.ValidateSignatureAlgorithm("dilithium5") {
		t.Error("dilithium5 should be a valid signature algorithm")
	}
	if wrapper.ValidateSignatureAlgorithm("invalid") {
		t.Error("invalid should not be a valid signature algorithm")
	}
}

func TestLibOQSWrapper_GenerateKEMKeyPair(t *testing.T) {
	wrapper, err := NewLibOQSWrapper()
	if err != nil {
		t.Fatalf("Failed to create LibOQSWrapper: %v", err)
	}

	algorithms := []string{"kyber512", "kyber768", "kyber1024"}

	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			pubKey, privKey, err := wrapper.GenerateKEMKeyPair(algo)
			if err != nil {
				t.Fatalf("Failed to generate %s key pair: %v", algo, err)
			}

			if len(pubKey) == 0 {
				t.Errorf("Public key for %s is empty", algo)
			}
			if len(privKey) == 0 {
				t.Errorf("Private key for %s is empty", algo)
			}

			// Test encapsulation and decapsulation
			ciphertext, sharedSecret1, err := wrapper.Encapsulate(algo, pubKey)
			if err != nil {
				t.Fatalf("Failed to encapsulate with %s: %v", algo, err)
			}

			if len(ciphertext) == 0 {
				t.Errorf("Ciphertext for %s is empty", algo)
			}
			if len(sharedSecret1) == 0 {
				t.Errorf("Shared secret for %s is empty", algo)
			}

			sharedSecret2, err := wrapper.Decapsulate(algo, ciphertext, privKey)
			if err != nil {
				t.Fatalf("Failed to decapsulate with %s: %v", algo, err)
			}

			if len(sharedSecret2) == 0 {
				t.Errorf("Decapsulated shared secret for %s is empty", algo)
			}

			// Verify that both shared secrets are identical
			if len(sharedSecret1) != len(sharedSecret2) {
				t.Errorf("Shared secret lengths don't match for %s: %d vs %d", algo, len(sharedSecret1), len(sharedSecret2))
			}

			for i := range sharedSecret1 {
				if sharedSecret1[i] != sharedSecret2[i] {
					t.Errorf("Shared secrets don't match for %s at position %d", algo, i)
					break
				}
			}
		})
	}
}

func TestLibOQSWrapper_GenerateSignatureKeyPair(t *testing.T) {
	wrapper, err := NewLibOQSWrapper()
	if err != nil {
		t.Fatalf("Failed to create LibOQSWrapper: %v", err)
	}

	algorithms := []string{"dilithium2", "dilithium3", "dilithium5"}

	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			pubKey, privKey, err := wrapper.GenerateSignatureKeyPair(algo)
			if err != nil {
				t.Fatalf("Failed to generate %s key pair: %v", algo, err)
			}

			if len(pubKey) == 0 {
				t.Errorf("Public key for %s is empty", algo)
			}
			if len(privKey) == 0 {
				t.Errorf("Private key for %s is empty", algo)
			}

			// Test signing and verification
			message := []byte("Hello, PQC World!")
			signature, err := wrapper.Sign(algo, message, privKey)
			if err != nil {
				t.Fatalf("Failed to sign with %s: %v", algo, err)
			}

			if len(signature) == 0 {
				t.Errorf("Signature for %s is empty", algo)
			}

			err = wrapper.Verify(algo, message, signature, pubKey)
			if err != nil {
				t.Fatalf("Failed to verify signature with %s: %v", algo, err)
			}

			// Test with wrong message
			wrongMessage := []byte("Wrong message")
			err = wrapper.Verify(algo, wrongMessage, signature, pubKey)
			if err == nil {
				t.Errorf("Signature verification should have failed for wrong message with %s", algo)
			}
		})
	}
}

func TestLibOQSWrapper_GetSupportedAlgorithms(t *testing.T) {
	wrapper, err := NewLibOQSWrapper()
	if err != nil {
		t.Fatalf("Failed to create LibOQSWrapper: %v", err)
	}

	kemAlgorithms := wrapper.GetSupportedKEMAlgorithms()
	if len(kemAlgorithms) != 3 {
		t.Errorf("Expected 3 KEM algorithms, got %d", len(kemAlgorithms))
	}

	expectedKEM := map[string]bool{
		"kyber512":  true,
		"kyber768":  true,
		"kyber1024": true,
	}

	for _, algo := range kemAlgorithms {
		if !expectedKEM[algo] {
			t.Errorf("Unexpected KEM algorithm: %s", algo)
		}
	}

	sigAlgorithms := wrapper.GetSupportedSignatureAlgorithms()
	if len(sigAlgorithms) != 3 {
		t.Errorf("Expected 3 signature algorithms, got %d", len(sigAlgorithms))
	}

	expectedSig := map[string]bool{
		"dilithium2": true,
		"dilithium3": true,
		"dilithium5": true,
	}

	for _, algo := range sigAlgorithms {
		if !expectedSig[algo] {
			t.Errorf("Unexpected signature algorithm: %s", algo)
		}
	}
}












