/*
  For processing the strings to be added to output as CRC64 hashes
*/
package main

type CRCHash struct {
	Plaintext string
	Hash      string
	Varname   string
}

func (c *CRCHash) GetPlaintext() string {
	return c.Plaintext
}

func (c *CRCHash) GetHash() string {
	return c.Hash
}

func (c *CRCHash) GetVarName() string {
	return c.Varname
}
