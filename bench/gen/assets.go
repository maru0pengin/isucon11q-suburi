//go:generate go run .
package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	rootPrefix string
)

type Variable struct {
	Hash map[string]string
	Map  map[string]string
}

func hash(path string) string {
	blob, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer blob.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, blob); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func glob(v *Variable, prefix string) {
	matches, err := filepath.Glob(filepath.Join(prefix, "*"))
	if err != nil {
		panic(err)
	}

	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			log.Fatal(err)
		}

		if stat.IsDir() {
			glob(v, match)
			continue
		}

		filename := strings.TrimPrefix(match, rootPrefix+"/")
		v.Hash["/"+filename] = hash(match)
		basename := strings.Split(filename, ".")[0]
		ext := strings.Split(filename, ".")[2]
		v.Map["/"+basename+"."+ext] = "/" + filename
	}
}

func main() {
	prefix := filepath.Join("..", "..", "webapp", "public", "assets")
	rootPrefix = prefix

	v := &Variable{
		Hash: map[string]string{},
		Map:  map[string]string{},
	}
	glob(v, prefix)

	io, err := os.Create(filepath.Join("../scenario", "assets.go"))
	if err != nil {
		log.Fatal(err)
	}

	err = template.Must(template.New("prog").Parse(prog)).Execute(io, v)
	if err != nil {
		log.Fatal(err)
	}
}

const prog = `// Code generated by gen/assets.go; DO NOT EDIT.
package scenario

var (
	resoucesHash = map[string]string{ {{range $key, $value := .Hash}}
		"{{$key}}": "{{$value}}",{{end}}
	}
	resoucesMap = map[string]string{ {{range $key, $value := .Map}}
		"{{$key}}": "{{$value}}",{{end}}
	}
)
`
