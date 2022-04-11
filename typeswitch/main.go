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
	default:
		fmt.Println("unknown type", t)
	}
}

func main() {
	var vs1 s1
	var vs2 s2
	vs3 := s3{I: 10}
	vs4 := s4{I: 10}

	printType(nil)
	printType(vs1)
	printType(vs2)
	printType(vs3)
	printType(vs4)

	var i any = s3{I: 10}

	if _, ok := i.(s3); ok {
		fmt.Println("i is s3")
	} else {
		fmt.Println("i is not s3")
	}

	if _, ok := i.(s4); ok {
		fmt.Println("i is s4")
	} else {
		fmt.Println("i is not s4 but it's", reflect.TypeOf(i))
	}
}
