package main

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateSecureKey(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(key), nil
}

// func main() {
// 	// 64 byte (512 bit)
// 	key, err := generateSecureKey(64)
// 	if err != nil {
// 		log.Fatalf("Key üretilirken hata oluştu: %v", err)
// 	}

// 	fmt.Println("Oluşturulan JWT Secret Key:")
// 	fmt.Println(key)
// }
