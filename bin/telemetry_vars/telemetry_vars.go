package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/egonelbre/slice"
)

type jsonVar struct {
	Type       string
	Name       string
	Desc       string
	Unit       string
	MemMapData bool
	DiskData   bool
}

func main() {
	jsonData, e := ioutil.ReadFile("vars.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	vars := []jsonVar{}
	json.Unmarshal(jsonData, &vars)

	// Group variables per variable type (float, bool, etc.)
	varsPerType := make(map[string][]jsonVar)
	for _, v := range vars {
		varsPerType[v.Type] = append(varsPerType[v.Type], v)
	}

	// Sort variables on name
	for _, vars := range varsPerType {
		slice.Sort(vars, func(i, j int) bool {
			nameA := strings.ToLower(vars[i].Name)
			nameB := strings.ToLower(vars[j].Name)
			return nameA < nameB
		})
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
