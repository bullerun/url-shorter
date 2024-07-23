package random

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func GetURL() string {
	randBytes := make([]byte, 6) // 6 байтов даст нам 8 символов после base64
	if _, err := rand.Read(randBytes); err != nil {
		log.Fatalf("Unable to generate random alias: %v", err)
	}
	return base64.URLEncoding.EncodeToString(randBytes)
}
