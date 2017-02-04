package main

import "fmt"

func untestFunc() {
	/*
		this is something other, not a doctest comment
	*/
}

func testFunc1(param string) string {
	//
	//  @test = {
	//  	{"", ""},
	//  }
	//
	return ""
}

func testFunc2(param string) string {
	/*
		@test = {
			{"", ""},
			{"foo", "bar"},
			{"baz", "quux"},
		}
	*/
	return ""
}

func main() {
	fmt.Println("vim-go")
}
