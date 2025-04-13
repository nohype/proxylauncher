package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "--exit-code" && len(args) > 1 {
			code, err := strconv.Atoi(args[1])
			if err == nil {
				os.Exit(code)
			}
		}
		fmt.Println("Received arguments:", args)
	}
	os.Exit(0)
}
