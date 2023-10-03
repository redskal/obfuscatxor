/*
 For structuring the strings to be processed and included into
 the generated source code file.
*/
package main

type ObsString struct {
	Plaintext       string
	Key             string
	Varname         string
	Encrypted       []byte
	EncryptedPretty string
}

func (s *ObsString) GetPlaintext() string {
	return s.Plaintext
}

func (s *ObsString) GetKey() string {
	return s.Key
}

func (s *ObsString) GetVarName() string {
	return s.Varname
}

func (s *ObsString) GetEncrypted() []byte {
	return s.Encrypted
}

func (s *ObsString) GetPretty() string {
	return s.EncryptedPretty
}
