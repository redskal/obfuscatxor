/*
 The template used to generate the Go source code.
*/
package main

const outTemplate = `
{{define "main"}}// Code generated by 'go generate'; DO NOT EDIT

package {{packagename}}

var (
{{range .String}}// Key: "{{.GetKey}}", String: "{{.GetPlaintext}}"
{{.GetVarName}} = {{.GetPretty}} 
{{end}}

{{range .Hash}}{{.GetVarName}} uint64 = {{.GetHash}} // String: "{{.GetPlaintext}}"
{{end}}
)
{{end}}
`
