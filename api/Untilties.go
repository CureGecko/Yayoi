/*
Utilities.go
Yayoi

Created by Cure Gecko on 5/13/15.
Copyright 2015, Cure Gecko. All rights reserved.

Different global utilities
*/

package main

import (
	"crypto/rand"
	"net/url"
	"regexp"
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

//Checks if a string is hexadecimal.
func isHexadecimal(s string) bool {
	return regexp.MustCompile("^[0-9a-fA-F]+$").MatchString(s)
}

//Checks if a string follows email address standards.
func isEmail(s string) bool {
	return regexp.MustCompile("^([\\w-]+(?:\\.[\\w-]+)*)@((?:[\\w-]+\\.)*\\w[\\w-]{0,66})\\.([a-z]{2,6}(?:\\.[a-z]{2})?)$").MatchString(s)
}

//Generates a URL with the correct site path and hostname.
func genURL(url *url.URL, path string) string {
	return url.Scheme + "://" + url.Host + SitePath + path
}
