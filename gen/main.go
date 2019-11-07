package main

import (
	"flag"
	"fmt"
	"mulbase/gen/parse"
)

func main() {
	in := flag.String("input", "", "Sets the root directory path for parsing SDL files.")
	out := flag.String("output", "", "Sets the root directory path for outputting go files.")
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Something went wrong! Resetting all files.")
			panic(r)
		}
	}()

	parse.Parse(*in, *out)
}
