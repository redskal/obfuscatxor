package main

import (
	"hash/crc64"
	"math/rand"
)

var hashTable = crc64.MakeTable(crc64.ECMA)

// SetHashTable sets HashTable used for crc64 checksums to specified value.
func SetCRCHashTable(table uint64) {
	hashTable = crc64.MakeTable(table)
}

// RandomiseHashTable sets HashTable used for crc64 checksums to a random value.
func RandomiseCRCHashTable() {
	hashTable = crc64.MakeTable(rand.Uint64())
}

// GetHash returns the crc64 hash of the specified string.
func GetCRCHash(input string) uint64 {
	return crc64.Checksum([]byte(input), hashTable)
}

// StringXOR taken from https://kylewbanks.com/blog/xor-encryption-using-go
func StringXOR(input, key string) (output string) {
	for i := 0; i < len(input); i++ {
		output += string(input[i] ^ key[i%len(key)])
	}
	return
}
