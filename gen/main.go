package main

import (
	"flag"
	"fmt"
	"github.com/Vliro/mulbase/gen/parse"
)



func main() {

	var conf parse.Config
	//Parse flags.
	//TODO: Use cobra layout?
	flag.StringVar(&conf.Input,"input", "", "Sets the root directory path for parsing SDL files.")
	flag.StringVar(&conf.Output,"output", "", "Sets the root directory path for outputting go files.")
	flag.StringVar(&conf.Package, "package", "gen", "Sets the package name in outputted files")
	flag.StringVar(&conf.State,"mode", "dgraph", "Sets which schema to use in generation for fields. Values are graphql or dgraph.")
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
	if conf.Input == "" || conf.Output == "" {
		flag.Usage()
		return
	}
	//Run the program.
	parse.Parse(&conf)
}
