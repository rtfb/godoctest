package testpkg

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
	switch param {
	case "":
		return ""
	case "foo":
		return "bar"
	case "baz":
		return "quux"
	}
	return ""
}

func multiline(aVeryLongParameterOne string, aVeryLongParameterTwo float64,
	aVeryLongParameterThree string) (veryValuableReturnValue string,
	errDamn error) {
	/*
		@test = {
			{"param1", 3.14, "param2", "return value", nil},
		}
	*/
	return "return value", nil
}

func fibonacci(n int) int {
	/*
		@test = {
			{1, 1},
			{2, 1},
			{3, 2},
			{7, 13},
			{11, 89},
		}
	*/
	if n == 1 {
		return 1
	}
	if n == 2 {
		return 1
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	fmt.Println("vim-go")
}
