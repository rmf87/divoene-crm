package crypto

import (
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := DeriveKey("test-secret")
	plaintext := []byte("sk_live_abc123xyz")

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if encrypted == "" {
		t.Fatal("encrypted string is empty")
	}
	// Encrypted output should not contain the plaintext
	if encrypted == "sk_live_abc123xyz" {
		t.Fatal("encrypted value must not be plaintext")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("round-trip failed: got %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestDifferentKeysProduceDifferentCiphertext(t *testing.T) {
	key1 := DeriveKey("secret-one")
	key2 := DeriveKey("secret-two")
	plaintext := []byte("same-data")

	enc1, _ := Encrypt(plaintext, key1)
	enc2, _ := Encrypt(plaintext, key2)

	if enc1 == enc2 {
		t.Fatal("different keys should produce different ciphertexts")
	}

	// Cross-decryption should fail
	if _, err := Decrypt(enc1, key2); err == nil {
		t.Fatal("decrypting with wrong key should fail")
	}
}

func TestEncryptSameKeyDifferentCiphertextEachCall(t *testing.T) {
	key := DeriveKey("fixed-key")
	plaintext := []byte("same-plaintext")

	enc1, _ := Encrypt(plaintext, key)
	enc2, _ := Encrypt(plaintext, key)

	// Each encryption should produce different ciphertext (random nonce)
	if enc1 == enc2 {
		t.Fatal("same plaintext should produce different ciphertext due to random nonce")
	}

	// Both should decrypt to the same plaintext
	dec1, _ := Decrypt(enc1, key)
	dec2, _ := Decrypt(enc2, key)
	if string(dec1) != string(plaintext) || string(dec2) != string(plaintext) {
		t.Fatal("both should decrypt to same plaintext")
	}
}

func TestDecryptTamperedData(t *testing.T) {
	key := DeriveKey("test-key")
	plaintext := []byte("tamper-me")

	enc, _ := Encrypt(plaintext, key)

	// Flip a byte in the middle of the hex string
	tampered := []byte(enc)
	tampered[10] ^= 0xff

	if _, err := Decrypt(string(tampered), key); err == nil {
		t.Fatal("tampered ciphertext should fail decryption")
	}
}

func TestDecryptShortData(t *testing.T) {
	key := DeriveKey("test-key")
	// Too short to contain nonce + data
	if _, err := Decrypt("aabb", key); err == nil {
		t.Fatal("short ciphertext should fail")
	}
}

func TestDecryptInvalidHex(t *testing.T) {
	key := DeriveKey("test-key")
	if _, err := Decrypt("not-hex-zzz!", key); err == nil {
		t.Fatal("invalid hex should fail")
	}
}

func TestDeriveKeyIsDeterministic(t *testing.T) {
	k1 := DeriveKey("same-secret")
	k2 := DeriveKey("same-secret")
	if string(k1) != string(k2) {
		t.Fatal("DeriveKey must be deterministic")
	}
}

func TestDeriveKeyLength(t *testing.T) {
	key := DeriveKey("any-secret")
	if len(key) != 32 {
		t.Fatalf("DeriveKey must return 32 bytes (AES-256), got %d", len(key))
	}
}
