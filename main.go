/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	assembleFlag = flag.Bool("a", false, "Immediately invoke asm2plan9s")
	stripFlag    = flag.Bool("s", false, "Strip comments")
	compactFlag  = flag.Bool("c", false, "Compact byte codes")
	formatFlag   = flag.Bool("f", false, "Format using asmfmt")
	compilerFlag = flag.String("m", "", "Optional C/CXX compiler options")
)

var (
	cCompileOptions = []string{
		"-masm=intel",
		"-mno-red-zone",
		"-mstackrealign",
		"-mllvm",
		"-inline-threshold=1000",
		"-fno-asynchronous-unwind-tables",
		"-fno-exceptions",
		"-fno-rtti",
	}
	cCompiler string = "clang"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("usage: goasmc [option] SOURCE_FILE")
		flag.PrintDefaults()
		return
	}
	var err error
	srcFile := flag.Arg(0)                                // A/B/C/src.{s,c}
	srcFileBase := filepath.Base(srcFile)                 // src.{s,c}
	ext := strings.ToLower(filepath.Ext(srcFileBase))     // .{s,c}
	srcFileBase = srcFileBase[:len(srcFileBase)-len(ext)] // src
	assemblyFile := srcFileBase + "_amd64.s"              // src_amd64.s
	goCompanion := srcFileBase + "_amd64.go"              // src_amd64.go

	switch ext {
	default:
		fmt.Printf("unsupported source file: %s\n", srcFile)
		os.Exit(1)
	case ".s":
	case ".c", ".cc", ".cpp":
		if *compilerFlag != "" {
			for _, s := range strings.Split(*compilerFlag, ",") {
				cCompileOptions = append(cCompileOptions, s)
			}
		}
		fmt.Println("Compiling", srcFile)
		invoke(cCompiler, append(cCompileOptions, "-S", srcFile)...)
	}

	srcFile = srcFileBase + ".s"
	if _, err = os.Stat(goCompanion); os.IsNotExist(err) {
		fmt.Printf("go file not found: %s\n", goCompanion)
		os.Exit(1)
	}

	fmt.Println("Processing", srcFile)
	lines, err := readLines(srcFile)
	if err != nil {
		fmt.Printf("readLines: %s\n", err)
		os.Exit(1)
	}
	result, err := process(lines, goCompanion)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}
	err = writeLines(result, assemblyFile, true)
	if err != nil {
		fmt.Printf("writeLines: %s\n", err)
		os.Exit(1)
	}
	if *assembleFlag {
		if err = invoke("asm2plan9s", assemblyFile); err != nil {
			fmt.Printf("asm2plan9s: %s\n", err)
			os.Exit(1)
		}
	}
	if *stripFlag {
		if err = stripGoasmComments(assemblyFile); err != nil {
			fmt.Printf("stripComments: %s\n", err)
			os.Exit(1)
		}
	}
	if *compactFlag {
		if err = compactOpcodes(assemblyFile); err != nil {
			fmt.Printf("compactOpcodes: %s\n", err)
			os.Exit(1)
		}
	}
	if *formatFlag {
		if err = invoke("asmfmt", "-w", assemblyFile); err != nil {
			fmt.Printf("asmfmt: %s\n", err)
			os.Exit(1)
		}
	}
}

func process(assembly []string, goCompanionFile string) ([]string, error) {
	// Split out the assembly source into subroutines
	subroutines := segmentSource(assembly)
	tables := segmentConstTables(assembly)

	var result []string
	for isubroutine, sub := range subroutines { // Iterate over all subroutines
		golangArgs, golangReturns := parseCompanionFile(goCompanionFile, sub.name)
		stackArgs := argumentsOnStack(sub.body)
		if len(golangArgs) > 6 && len(golangArgs)-6 < stackArgs.Number {
			return nil, fmt.Errorf("not enough arguments on stack (%d) but needed %d", len(golangArgs)-6, stackArgs.Number)
		}
		if table := getCorrespondingTable(sub.body, tables); table.isPresent() { // Check for constants table
			// Output constants table
			result = append(result, strings.Split(table.Constants, "\n")...)
			result = append(result, "") // append empty line
			sub.table = table
		}

		// Create object to get offsets for stack pointer
		stack := NewStack(sub.epilogue, len(golangArgs), scanBodyForCalls(sub))
		// Write header for subroutine in go assembly
		result = append(result, writeGoasmPrologue(sub, stack, golangArgs, golangReturns)...)
		// Write body of code
		assembly, err := writeGoasmBody(sub, stack, stackArgs, golangArgs, golangReturns)
		if err != nil {
			return nil, err
		}
		result = append(result, assembly...)
		if isubroutine < len(subroutines)-1 {
			// Empty lines before next subroutine
			result = append(result, "\n", "\n")
		}
	}
	return result, nil
}
