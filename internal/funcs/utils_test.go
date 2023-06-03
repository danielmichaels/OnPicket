package funcs

import (
	"fmt"
	"os"
)

func ExampleGetEnv() {
	fmt.Println(GetEnv("NOT_EXIST", "default"))
	_ = os.Setenv("DOES_EXIST", "i_exist")
	fmt.Println(GetEnv("DOES_EXIST", "i_exist"))

	// Output:
	// default
	// i_exist
}
