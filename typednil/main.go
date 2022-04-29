package main

import "fmt"

type MyInterface interface {
	Method() error
}

type MyStruct struct{}

func (m *MyStruct) Method() error {
	return nil
}

func main() {
	testI := func(m MyInterface) {
		fmt.Println(m == nil)
	}
	testS := func(m *MyStruct) {
		fmt.Println(m == nil)
	}
	testS(nil)                                        // true
	testI(nil)                                        // true
	testS((*MyStruct)(nil))                           // true
	testI((*MyStruct)(nil))                           // false --> typed nil pointer
	fmt.Println(MyInterface((*MyStruct)(nil)) == nil) // false (same as using func)
}
