package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

// PublicKeyToID converts a public key to a challenger ID
func PublicKeyToID(publicKey *ecdsa.PublicKey) string {
	// 공개키를 바이트로 변환
	pubKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// SHA256 해시
	hash := sha256.Sum256(pubKeyBytes)

	// 해시의 앞 20바이트를 16진수 문자열로 변환
	return hex.EncodeToString(hash[:20])
}

// SignMessage signs a message with a private key
func SignMessage(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error) {
	// 메시지 해시
	hash := sha256.Sum256(message)

	// 서명 생성
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	// r과 s를 바이트로 변환하여 결합
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature verifies a signature against a message and public key
func VerifySignature(publicKey *ecdsa.PublicKey, message []byte, signature []byte) bool {
	// 메시지 해시
	hash := sha256.Sum256(message)

	// 서명이 64바이트인지 확인 (32바이트 r + 32바이트 s)
	if len(signature) != 64 {
		return false
	}

	// r과 s 분리
	r := signature[:32]
	s := signature[32:]

	// big.Int로 변환
	var rInt, sInt = new(big.Int), new(big.Int)
	rInt.SetBytes(r)
	sInt.SetBytes(s)

	// 서명 검증
	return ecdsa.Verify(publicKey, hash[:], rInt, sInt)
}

// HashData creates a SHA256 hash of the given data
func HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// GenerateNodeID generates a unique node ID
func GenerateNodeID() string {
	// 32바이트 랜덤 데이터 생성
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// 에러 발생 시 현재 시간 기반으로 생성
		return fmt.Sprintf("node_%d", time.Now().UnixNano())
	}

	// SHA256 해시 후 16진수 문자열로 변환
	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:16]) // 앞 16바이트만 사용
}

// ValidatePublicKey validates if a public key is valid
func ValidatePublicKey(publicKey *ecdsa.PublicKey) bool {
	if publicKey == nil {
		return false
	}

	// 곡선이 올바른지 확인
	if publicKey.Curve != elliptic.P256() {
		return false
	}

	// 키가 곡선 위에 있는지 확인
	return publicKey.Curve.IsOnCurve(publicKey.X, publicKey.Y)
}

// PublicKeyToBytes converts a public key to byte slice
func PublicKeyToBytes(publicKey *ecdsa.PublicKey) []byte {
	if publicKey == nil {
		return nil
	}
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
}

// BytesToPublicKey converts byte slice to public key
func BytesToPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	x, y := elliptic.Unmarshal(elliptic.P256(), data)
	if x == nil || y == nil {
		return nil, fmt.Errorf("invalid public key data")
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}, nil
}

// GenerateChallengerID generates a unique challenger ID (string version for discovery)
func GenerateChallengerID(address string, timestamp time.Time) string {
	// Simple version using address and timestamp
	combined := fmt.Sprintf("%s_%d", address, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:20]) // 40 character hex string
}
