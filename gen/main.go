package main

import (
	"flag"
	"fmt"
	"github.com/Vliro/mulbase/gen/parse"
)

func main() {
	//Parse flags.
	//TODO: Use cobra layout?
	in := flag.String("input", "", "Sets the root directory path for parsing SDL files.")
	out := flag.String("output", "", "Sets the root directory path for outputting go files.")
	mode := flag.String("mode", "dgraph", "Sets which schema to use in generation for fields. Values are graphql or dgraph.")
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
	if *in == "" || *out == "" || *mode == "" {
		flag.Usage()
		return
	}
	//Run the program.
	parse.Parse(*in, *out, *mode)
}
