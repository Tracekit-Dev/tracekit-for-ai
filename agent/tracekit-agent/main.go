package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 1
	}
	switch args[0] {
	case "auth":
		return runAuth(args[1:])
	case "mcp":
		return runMCP()
	case "version", "--version", "-v":
		fmt.Println(versionString())
		return 0
	case "help", "--help", "-h":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		printUsage()
		return 1
	}
}

func versionString() string { return version }

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: tracekit-agent <auth|mcp|version>")
}
