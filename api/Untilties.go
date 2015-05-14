package main

import (
	"crypto/rand"
)

const allowedCharacters = "abcdefghijklmnopqrstuvwxyz1234567890=-_!"

func randomString(count int) string {
	bytes := make([]byte, count)
	rand.Read(bytes)
	for i, b := range bytes {
		character := int(b) % len(allowedCharacters)
		bytes[i] = allowedCharacters[character]
	}
	return string(bytes)
}
