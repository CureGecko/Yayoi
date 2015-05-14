package main

import (
	"crypto/rand"
)

//Characters that are allowed to be used in the random string function. Because upper case characters are included, you must do a comparison in the database using "COLLATE utf8_bin"
const allowedCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890=-_!"

//Creates a random string using the allowed characters to the given length.
func randomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		character := int(b) % len(allowedCharacters)
		bytes[i] = allowedCharacters[character]
	}
	return string(bytes)
}

//Removes empty strings from an array.
func Filter(a []string) []string {
	var result []string
	for _, c := range a {
		if c != "" {
			result = append(result, c)
		}
	}
	return result
}
