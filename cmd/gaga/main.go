package main

import (
	"fmt"
	"gaga/compiler"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gaga <sourcefile.gaga> [outputfile]")
		os.Exit(1)
	}

	sourceFile := os.Args[1]
	outputFile := "a.out.go"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	c := compiler.NewCompiler(sourceFile, outputFile)
	if err := c.Compile(); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Compilation successful! Output:", outputFile)
}
