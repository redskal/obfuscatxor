/*
For generating the source code file from the structs that have
been populated.
*/
package main

import (
	"bufio"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const (
	OBS_STRING = iota + 1
	OBS_HASH
)

type Source struct {
	String      []*ObsString
	Hash        []*CRCHash
	PackageName string
}

// ParseFiles parses files into a Source instance
func ParseFiles(files []string) (*Source, error) {
	// instantiate a new Source struct
	src := &Source{
		String: make([]*ObsString, 0),
	}

	// process the files provided
	for _, file := range files {
		if err := src.ParseFile(file); err != nil {
			return nil, err
		}
	}

	return src, nil
}

// ParseFile adds additional string entries to Source
func (src *Source) ParseFile(path string) error {
	// open the file provided
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// new scanner, innit
	s := bufio.NewScanner(file)
	for s.Scan() {
		t := strings.TrimSpace(s.Text())

		// not the droid you are looking for
		if !strings.HasPrefix(t, prefixObfuscate) && !strings.HasPrefix(t, prefixHash) {
			continue
		}

		// we're not interested in the prefix
		var obfuscate bool // used to judge between xor and crc hash
		if strings.HasPrefix(t, prefixObfuscate) {
			t = t[len(prefixObfuscate):]
			obfuscate = true
		} else if strings.HasPrefix(t, prefixHash) {
			t = t[len(prefixHash):]
			obfuscate = false
		}

		// a weird false positive?
		if !(t[0] == ' ' || t[0] == '\t') {
			continue
		}
		// no thanks, whitespace
		t = strings.TrimSpace(t)

		// process the string we have either through xor or hashing
		if obfuscate {
			str, err := newString(t)
			if err != nil {
				return err
			}
			src.String = append(src.String, str)
		} else {
			hsh, err := newHash(t)
			if err != nil {
				return err
			}
			src.Hash = append(src.Hash, hsh)
		}
	}
	if err := s.Err(); err != nil {
		return err
	}

	// get the package name
	fset := token.NewFileSet()
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	pkg, err := parser.ParseFile(fset, "", file, parser.PackageClauseOnly)
	if err != nil {
		return err
	}
	src.PackageName = pkg.Name.Name

	return nil
}

// Generate output source file
func (src *Source) Generate(w io.Writer) error {
	// create our function map
	funcMap := template.FuncMap{
		"packagename": src.GetPackageName,
	}

	// process template and write to w
	t := template.Must(template.New("main").Funcs(funcMap).Parse(outTemplate))
	err := t.Execute(w, src)
	if err != nil {
		return errors.New("Failed to execute template: " + err.Error())
	}
	return nil
}

// GetPackageName returns the package name of src
func (src *Source) GetPackageName() string {
	return src.PackageName
}

// newHash parses string s and returns CRCHash
func newHash(s string) (*CRCHash, error) {
	s = strings.TrimSpace(s)

	crc := &CRCHash{}

	var p string
	var b string
	var found bool
	for i := 0; i <= 1; i++ {
		p, b, s, found = extractSection(s, '(', ')')
		if !found {
			return nil, errors.New("Could not extract information from \"" + s + "\"")
		}
		switch strings.ToLower(p) {
		case "phrase":
			crc.Plaintext = b
		case "varname":
			crc.Varname = b
		}

	}
	fmt.Printf("DEBUG: hash(phrase=\"%s\" varname=\"%s\")\n", crc.Plaintext, crc.Varname)

	// Can't hash a non-existant string or delcare it if name is not known
	if crc.Plaintext == "" || crc.Varname == "" {
		return nil, errors.New("variable name or plaintext string not supplied")
	}

	// hash the string using ECMA CRC64 hash table
	uintHash := GetCRCHash(crc.Plaintext)
	crc.Hash = strconv.FormatUint(uintHash, 10)

	return crc, nil
}

// newString parses string s and returns an ObsString.
func newString(s string) (*ObsString, error) {
	s = strings.TrimSpace(s)

	str := &ObsString{}

	// extract the 3 strings
	var p string
	var b string
	var found bool
	for i := 0; i <= 2; i++ {
		p, b, s, found = extractSection(s, '(', ')')
		if !found || !containsWord(p) {
			return nil, errors.New("Could not extract information from \"" + s + "\".")
		}
		switch strings.ToLower(p) {
		case "key":
			str.Key = b
		case "phrase":
			str.Plaintext = b
		case "varname":
			str.Varname = b
		}
	}
	fmt.Printf("DEBUG: obfuscate(key=\"%s\" phrase=\"%s\" varname=\"%s\")\n", str.GetKey(), str.GetPlaintext(), str.GetVarName())

	// we can't encrypt the string and produce output without these.
	if str.Key == "" || str.Plaintext == "" || str.Varname == "" {
		return nil, errors.New("key, variable name, or plaintext string not supplied")
	}

	// encrypt the string and get a pretty version
	str.Encrypted = []byte(StringXOR(str.Plaintext, str.Key))
	str.EncryptedPretty = prettifyBytes(str.Encrypted)

	return str, nil
}

// prettifyBytes makes a printable version of the encrypted string for adding
// to the output source file
func prettifyBytes(b []byte) (pretty string) {
	for i, v := range b {
		if (i % 15) == 0 {
			pretty += "\n"
		}
		if i != (len(b) - 1) {
			pretty += fmt.Sprintf("0x%02x", v) + ", "
		} else {
			pretty += fmt.Sprintf("0x%02x", v)
		}
	}
	pretty = fmt.Sprintf("[]byte{%s }", pretty)
	return
}

// containsWord is a helper function for checking the strings extracted
// for the correct formatting. Makes newString less ugly to read.
func containsWord(s string) bool {
	return (strings.ToLower(s) == "key" || strings.ToLower(s) == "phrase" ||
		strings.ToLower(s) == "varname")
}

// extractSection extracts text out of string s starting after start
// and ending just before end. found return value will indicate success,
// and prefix, body and suffix will contain correspondent parts of string s.
func extractSection(s string, start, end rune) (prefix, body, suffix string, found bool) {
	// ripped straight from https://github.com/C-Sto/BananaPhone/blob/master/cmd/mkdirectwinsyscall/function.go
	// which was ripped from https://cs.opensource.google/go/x/sys/+/5a0f0661:windows/mkwinsyscall/mkwinsyscall.go;l=403
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, string(start)) {
		// no prefix
		body = s[1:]
	} else {
		a := strings.SplitN(s, string(start), 2)
		if len(a) != 2 {
			suffix = s
			found = false
			return
		}
		prefix = a[0]
		body = a[1]
	}
	a := strings.SplitN(body, string(end), 2)
	if len(a) != 2 {
		//has no end marker. suffix won't be set, but body/prefix may be
		found = false
		return
	}
	return prefix, a[0], a[1], true
}
