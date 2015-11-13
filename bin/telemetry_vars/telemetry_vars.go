package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
)

type varHeader struct {
	Type VarType // VarType
	Name string
	Desc string
	Unit string // something like "kg/m^2"
}

type VarType int32

const (
	// 1 byte
	CharType VarType = iota
	BoolType VarType = iota

	// 4 bytes
	IntType      VarType = iota
	BitfieldType VarType = iota
	FloatType    VarType = iota

	// 8 bytes
	DoubleType VarType = iota

	// index, don't use
	ETCount = iota
)

func main() {
	jsonData, e := ioutil.ReadFile("vars.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	vars := []varHeader{}
	json.Unmarshal(jsonData, &vars)

	varsPerType := make(map[string][]varHeader)
	for _, v := range vars {
		switch v.Type {
		case CharType:
			varsPerType["char"] = append(varsPerType["char"], v)
		case BoolType:
			varsPerType["bool"] = append(varsPerType["bool"], v)
		case IntType:
			varsPerType["int"] = append(varsPerType["int"], v)
		case BitfieldType:
			varsPerType["bitfield"] = append(varsPerType["bitfield"], v)
		case FloatType:
			varsPerType["float"] = append(varsPerType["float"], v)
		case DoubleType:
			varsPerType["double"] = append(varsPerType["double"], v)
		default:
			panic("Unknown type")
		}
	}

	tmpl, err := template.New("telemetry_vars.tmpl").
		Funcs(template.FuncMap{"ucFirst": ucFirst}).
		ParseFiles("telemetry_vars.tmpl")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, varsPerType)
	if err != nil {
		panic(err)
	}
}

func ucFirst(s string) string {
	b := []byte(s)
	b[0] = bytes.ToUpper(b[0:1])[0]
	return string(b)
	// r, n := utf8.DecodeRuneInString(s)
	// return string(unicode.ToUpper(r)) + s[n:]
}
