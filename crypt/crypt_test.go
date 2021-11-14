package crypt

import (
	"log"
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	text := "91fPLgXt3tJPZGyDSLEFnD4btsZ9UZ86ibUtShVPsPMJxP15qJP"
	key := "sixteenbytesword"
	encryptedFilepath := "./encypted-private-keys"
	got := Encrypt(text, key)

	if Decrypt(got, key) != text {
		log.Fatal("not equal")
	}

	_ = os.Remove(encryptedFilepath)
	err := os.WriteFile(encryptedFilepath, []byte(got), 0644)
	check(err)

	gotFile, err := os.ReadFile(encryptedFilepath)
	check(err)
	if Decrypt(string(gotFile), key) != text {
		log.Fatal("not equal")
	}
}
