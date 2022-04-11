package main

import (
	"fmt"
	"reflect"
)

type s1 struct{}

type s2 struct{}

type s3 struct {
	I int
}

type s4 struct {
	I int
}

type runner interface {
	Run() error
}

type walker interface {
	Walk() error
}

type thing struct {
}

func (t *thing) Run() error {
	return nil
}

type thing2 struct {
}

func (t *thing2) Run() error {
	return nil
}

func (t thing2) Walk() error {
	return nil
}

func testRunner(name string, r runner) {
	if _, ok := r.(walker); ok {
		fmt.Println(name, "runner passsed to testRunner compat with walker")
	} else {
		fmt.Println(name, "runner passsed to testRunner not compat with walker")
	}
}

func printType(v any) {
	switch t := v.(type) {
	case nil:
		fmt.Println("nil", t)
	case s1:
		fmt.Println("s1", t)
	case s2:
		fmt.Println("s2", t)
	case s3:
		fmt.Println("s3", t)
	case s4:
		fmt.Println("s4", t)
	case runner:
		fmt.Println("runner", t)
	case thing:
		fmt.Println("thing", t)
	default:
		fmt.Println("unknown type", t)
	}
}

func main() {
	var vs1 s1
	var vs2 s2
	vs3 := s3{I: 10}
	vs4 := s4{I: 10}
	vthing := thing{}
	vthing2 := thing2{}

	printType(nil)
	printType(vs1)
	printType(vs2)
	printType(vs3)
	printType(vs4)
	printType(vthing)

	testRunner("vthing", &vthing)
	testRunner("vthing2", &vthing2)

	var is3 any = s3{I: 10}

	if _, ok := is3.(s3); ok {
		fmt.Println("is3 is s3")
	} else {
		fmt.Println("is3 is not s3")
	}

	if _, ok := is3.(s4); ok {
		fmt.Println("is3 is s4")
	} else {
		fmt.Println("is3 is not s4 but it's", reflect.TypeOf(is3))
	}

	var ithing any = thing{}

	if _, ok := ithing.(thing); ok {
		fmt.Println("ithing is thing")
	} else {
		fmt.Println("ithing is not thing but it's", reflect.TypeOf(ithing))
	}

	if _, ok := ithing.(runner); ok {
		fmt.Println("ithing is runner")
	} else {
		fmt.Println("ithing is not runner but it's", reflect.TypeOf(ithing))
	}

	if _, ok := ithing.(thing2); ok {
		fmt.Println("ithing is thing2")
	} else {
		fmt.Println("ithing is not thing2 but it's", reflect.TypeOf(ithing))
	}
}
