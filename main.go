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
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	assembleFlag = flag.Bool("a", false, "Immediately invoke asm2plan9s")
	stripFlag    = flag.Bool("s", false, "Strip comments")
	compactFlag  = flag.Bool("c", false, "Compact byte codes")
	formatFlag   = flag.Bool("f", false, "Format using asmfmt")
)

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Printf("error: not enough input files specified\n\n")
		fmt.Println("usage: c2goasm /path/to/c-project/build/SomeGreatCode.cpp.s SomeGreatCode_amd64.s")
		return
	}
	assemblyFile := flag.Arg(1)
	if !strings.HasSuffix(assemblyFile, ".s") {
		fmt.Printf("error: second parameter must have '.s' extension\n")
		return
	}
	goCompanion := assemblyFile[:len(assemblyFile)-2] + ".go"
	if _, err := os.Stat(goCompanion); os.IsNotExist(err) {
		fmt.Printf("error: companion '.go' file is missing for %s\n", flag.Arg(1))
		return
	}
	fmt.Println("Processing", flag.Arg(0))
	lines, err := readLines(flag.Arg(0))
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	result, err := process(lines, goCompanion)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}
	err = writeLines(result, assemblyFile, true)
	if err != nil {
		log.Fatalf("writeLines: %s", err)
	}
	if *assembleFlag {
		fmt.Println("Invoking asm2plan9s on", assemblyFile)
		cmd := exec.Command("asm2plan9s", assemblyFile)
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("asm2plan9s: %v", err)
		}
	}
	if *stripFlag {
		stripGoasmComments(assemblyFile)
	}
	if *compactFlag {
		compactOpcodes(assemblyFile)
	}
	if *formatFlag {
		cmd := exec.Command("asmfmt", "-w", assemblyFile)
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("asmfmt: %v", err)
		}
	}
}

func process(assembly []string, goCompanionFile string) ([]string, error) {
	// Split out the assembly source into subroutines
	subroutines := segmentSource(assembly)
	tables := segmentConstTables(assembly)

	var result []string
	// Iterate over all subroutines
	for isubroutine, sub := range subroutines {
		golangArgs, golangReturns := parseCompanionFile(goCompanionFile, sub.name)
		stackArgs := argumentsOnStack(sub.body)
		if len(golangArgs) > 6 && len(golangArgs)-6 < stackArgs.Number {
			panic(fmt.Sprintf("Found too few arguments on stack (%d) but needed %d", len(golangArgs)-6, stackArgs.Number))
		}

		// Check for constants table
		if table := getCorrespondingTable(sub.body, tables); table.isPresent() {
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
			panic(fmt.Sprintf("writeGoasmBody: %v", err))
		}
		result = append(result, assembly...)
		if isubroutine < len(subroutines)-1 {
			// Empty lines before next subroutine
			result = append(result, "\n", "\n")
		}
	}
	return result, nil
}
