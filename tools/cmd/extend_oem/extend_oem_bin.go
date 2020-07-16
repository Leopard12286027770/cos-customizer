package main

import (
	"cos-customizer/tools"
	"log"
	"os"
	"strconv"
)

// main generates binary file to extend the OEM partition.
// Built by Bazel. The binary will be in data/builtin_build_context/.
func main() {
	log.SetOutput(os.Stdout)
	args := os.Args
	if len(args) != 5 {
		log.Fatalln("error: must have 4 arguments: disk string, statePartNum, oemPartNum int, oemSize string")
	}
	statePartNum, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalln("error: the 2nd argument statePartNum must be an int")
	}
	oemPartNum, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatalln("error: the 3rd argument oemPartNum must be an int")
	}
	err = tools.ExtendOEMPartition(args[1], statePartNum, oemPartNum, args[4])
	if err != nil {
		log.Fatalln(err)
	}
}
