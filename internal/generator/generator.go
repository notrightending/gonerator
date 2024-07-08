package generator

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

// Generate parses the input file, extracts API method information,
// and generates handler code based on the parsed information.
func Generate(inputFile, outputFile string) error {
	// Parse the input file
	methods, err := parseFile(inputFile)
	if err != nil {
		return err
	}

	// Get the package name from the input file
	packageName, err := getPackageName(inputFile)
	if err != nil {
		return err
	}

	// Group methods by receiver type
	groupedMethods := make(map[string][]Method)
	for _, method := range methods {
		groupedMethods[method.ReceiverType] = append(groupedMethods[method.ReceiverType], method)
	}

	// Prepare data for template
	data := struct {
		PackageName string
		Methods     map[string][]Method
	}{
		PackageName: packageName,
		Methods:     groupedMethods,
	}

	// Generate handler code using the template
	var buf bytes.Buffer
	err = handlerTemplate.Execute(&buf, data)
	if err != nil {
		return err
	}

	// Format the generated code
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	// Write the formatted code to the output file
	err = os.WriteFile(outputFile, formattedCode, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getPackageName(filename string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return node.Name.Name, nil
}
