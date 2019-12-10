package main

import (
	"flag"
	"github.com/Vliro/humus/gen/parse"
)

func main() {

	var conf parse.Config
	//Parse flags.
	//TODO: Use cobra layout?
	flag.StringVar(&conf.Input,"input", "", "Sets the root directory path for parsing SDL files.")
	flag.StringVar(&conf.Output,"output", "", "Sets the root directory path for outputting go files.")
	flag.StringVar(&conf.Package, "package", "gen", "Sets the package name in outputted files")
	flag.BoolVar(&conf.EasyJson, "json", true, "Sets whether to generate easyjson preamble for code generation.")
	flag.Parse()

	if conf.Input == "" || conf.Output == "" {
		flag.Usage()
		return
	}
	//Run the program.
	parse.Parse(&conf)
}


