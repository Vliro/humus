package main

import (
	"flag"
	"fmt"
	"mulbase/gen/parse"
)

func main() {
	//Parse flags.
	//TODO: Use cobra layout?
	in := flag.String("input", "", "Sets the root directory path for parsing SDL files.")
	out := flag.String("output", "", "Sets the root directory path for outputting go files.")
	flag.Parse()
	/*
		TODO: Fix proper handling if shit goes bad.
	 */
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Something went wrong! Resetting all files.")
			panic(r)
		}
	}()
	if *in == "" || *out == "" {
		panic("missing input and/or output")
	}
	//Run the program.
	parse.Parse(*in, *out)
}
