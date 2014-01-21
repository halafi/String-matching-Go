// Package input provides input funcionality - read of STDIN or command
// line argument.
package input

import "io/ioutil"
import "os"
import "code.google.com/p/go.crypto/ssh/terminal"
import "log"
import "strings"

// ReadLog attempts to read Log data from STDIN if it's possible, if not
// it tries reading from a FilePath given in a single command line
// argument.
func ReadLog() (logLines []string) {
	if !terminal.IsTerminal(0) {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		logLines = lineSplit(string(bytes))
	} else {
		if len(os.Args) == 2 {
			logLines = lineSplit(fileToString(os.Args[1]))
		} else {
			log.Fatal("No standard input or FilePath argument given.")
		}
	}
	return logLines
}

// ReadPatterns reads a single file of patterns located at
// 'filePath' argument location.
func ReadPatterns(filePath string) (output []string) {
	patterns := lineSplit(fileToString(filePath))
	output = make([]string, 0)
	for i := range patterns {
		if patterns[i] == "" || patterns[i][0] == '#' {
			// skip empty lines and comments
		} else {
			patternsNameSplit := strings.Split(patterns[i], "##") //separate pattern name from its definition
			if len(patternsNameSplit) != 2 {
				log.Fatal("Error with pattern number ", i+1, " name, use [NAME##<token> word ...].")
			}
			if len(patternsNameSplit[0]) == 0 {
				log.Fatal("Error with pattern number ", i+1, ": name cannot be empty.")
			}
			if len(patternsNameSplit[1]) == 0 {
				log.Fatal("Error with pattern number ", i+1, ": pattern cannot be empty.")
			}
			newOutput := make([]string, cap(output)+1)
			copy(newOutput, output)
			output = newOutput
			output[len(output)-1] = patterns[i]
		}
	}
	return output
}

// ReadTokens reads a single file of tokens (regex definitions) located
// at 'filePath' argument location into map of key=token, value=regex.
func ReadTokens(filePath string) (output map[string]string) {
	tokens := lineSplit(fileToString(filePath))
	output = make(map[string]string)
	for t := range tokens {
		if tokens[t] == "" || tokens[t][0] == '#' {
			// skip empty lines and comments
		} else {
			currentTokenLine := strings.Split(tokens[t], " ")
			if len(currentTokenLine) == 2 {
				output[currentTokenLine[0]] = currentTokenLine[1]
			} else {
				log.Fatal("Problem in tokens definition, error reading: " + tokens[t])
			}
		}
	}
	return output
}

// Simple file reader that returns a string content of a given
// 'filePath' file location.
func fileToString(filePath string) string {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return string(file)
}

// Function that parses a mutli-line string into single lines (array of
// strings).
func lineSplit(input string) []string {
	inputSplit := make([]string, 1)
	inputSplit[0] = input                // default single pattern, no line break
	if strings.Contains(input, "\r\n") { //CR+LF
		inputSplit = strings.Split(input, "\r\n")
	} else if strings.Contains(input, "\n") { //LF
		inputSplit = strings.Split(input, "\n")
	} else if strings.Contains(input, "\r") { //CR
		inputSplit = strings.Split(input, "\r")
	}
	return inputSplit
}
