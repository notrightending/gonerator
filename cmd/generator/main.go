package main

import (
	"fmt"
	"log"
	"os"

	"github.com/notrightending/gonerator/internal/generator"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: generator <input_file> <output_file>")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	err := generator.Generate(inputFile, outputFile)
	if err != nil {
		log.Fatalf("Error generating handlers: %v", err)
	}

	fmt.Printf("Generated handlers written to %s\n", outputFile)
}
