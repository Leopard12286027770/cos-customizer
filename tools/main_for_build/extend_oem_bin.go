package main

import (
	"cos-customizer/tools"
	"log"
	"os"
	"strconv"
)

// main generates binary file to extend the OEM partition
// cmd: GOOS=linux GOARCH=amd64 CGO_ENABLE=0 go build -o ./../../data/builtin_build_context/extend-oem.bin extend_oem_bin.go
func main() {
	args := os.Args
	if len(args) != 5 {
		log.Println("error: must have 4 arguments: disk string, statePartNum, oemPartNum int, oemSize string")
		return
	}
	statePartNum, err := strconv.Atoi(args[2])
	if err != nil {
		log.Println("error: the 2nd argument statePartNum must be an int")
		return
	}
	oemPartNum, err := strconv.Atoi(args[3])
	if err != nil {
		log.Println("error: the 3rd argument oemPartNum must be an int")
		return
	}
	err = tools.ExtendOEMPartition(args[1], statePartNum, oemPartNum, args[4])
	if err != nil {
		log.Println(err)
	}
}
