// Package main provides the CLI entry point for ReportGo.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dannyswat/reportgo/pkg/reportgo"
)

var version = "0.1.0"

func main() {
	templatePath := flag.String("template", "", "Path to the XML template file")
	dataPath := flag.String("data", "", "Path to the JSON data file")
	outputPath := flag.String("output", "output.pdf", "Path for the output PDF file")
	validateOnly := flag.Bool("validate", false, "Only validate the template without generating PDF")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help information")

	flag.Parse()

	if *showVersion {
		fmt.Printf("ReportGo version %s\n", version)
		os.Exit(0)
	}

	if *showHelp || *templatePath == "" {
		printUsage()
		os.Exit(0)
	}

	engine := reportgo.New()

	if err := engine.LoadTemplate(*templatePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading template: %v\n", err)
		os.Exit(1)
	}

	if *validateOnly {
		fmt.Println("Template validation successful")
		os.Exit(0)
	}

	if *dataPath != "" {
		if err := engine.LoadDataFromFile(*dataPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
			os.Exit(1)
		}
	}

	if err := engine.Generate(nil, *outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PDF generated successfully: %s\n", *outputPath)
}

func printUsage() {
	fmt.Println("ReportGo - PDF Report Generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  reportgo -template <template.xml> -data <data.json> -output <output.pdf>")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  reportgo -template report.xml -data data.json -output report.pdf")
	fmt.Println("  reportgo -template report.xml -validate")
}
