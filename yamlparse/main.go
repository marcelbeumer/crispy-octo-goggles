package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Root struct {
	Strlist []string `yaml:"strlist"`
	Int     int      `yaml:"int"`
	Nested  *Example `yaml:"nested,omitempty"`
}

type Example struct {
	Root Root `yaml:"root"`
}

func test(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}

	var data Example
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	b, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	fmt.Print(string(b[:]))

	return nil
}

func main() {
	fpath := "./yamlparse/example.yaml"
	if len(os.Args) > 1 {
		fpath = os.Args[1]
	}
	if err := test(fpath); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
