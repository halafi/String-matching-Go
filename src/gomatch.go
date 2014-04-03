package main

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	// Command-line flags.
	inputFilePath         = flag.String("i", "/dev/stdin", "Data input.")
	inputSocketFilePath   = flag.String("s", "none", "Reading from Socket.")
	ampqConfigFilePath    = flag.String("a", "none", "Filepath for AMQP config file.")
	patternsFilePath      = flag.String("p", "Patterns", "Patterns input.")
	tokensFilePath        = flag.String("t", "Tokens", "Tokens input.")
	outputFilePath        = flag.String("o", "/dev/stdout", "Matched data output.")
	noMatchOutputFilePath = flag.String("n", "no_match.log", "Unmatched data output.")
	// Shared variables between all goroutines.
	trie          map[int]map[Token]int
	finalFor      []int
	state         int
	patternNumber int
	patterns      []Pattern
	regexes       map[string]Regex
)

// Starts when the program is executed.
// Performs parsing of flags, reading of both Tokens and Patterns file,
// prefix tree construction and output files init.
// Runs separate goroutine for watching file with patterns.
// Uses AMQP if flag -a is set. Otherwise reads input from either socket
// or input file/pipe.
// For each input line performs matching and writing to output.
func main() {
	flag.Parse()

	if *ampqConfigFilePath != "none" && *inputSocketFilePath != "none" {
		log.Fatal("cannot use both socket and amqp at the same time")
	}

	trie, finalFor, state, patternNumber = initTrie()

	regexes, patterns = readPatterns(*patternsFilePath, *tokensFilePath)
	for p := range patterns {
		finalFor, state, patternNumber = appendPattern(patterns[p], trie, finalFor, state, patternNumber, regexes)
	}

	outputFile := createFile(*outputFilePath)
	noMatchOutputFile := createFile(*noMatchOutputFilePath)

	go watchPatterns()

	if *ampqConfigFilePath != "none" { // amqp
		// init configuration parameters
		parseAmqpConfigFile(*ampqConfigFilePath)

		// set up connections and channels, ensure that they are closed
		cSend := openConnection(amqpMatchedSendUri)
		chSend := openChannel(cSend)
		defer cSend.Close()
		defer chSend.Close()

		cReceive := openConnection(amqpReceiveUri)
		chReceive := openChannel(cReceive)
		defer cReceive.Close()
		defer chReceive.Close()

		// declare queues
		qReceive := declareQueue(amqpReceiveQueueName, chReceive)
		qSend := declareQueue(amqpMatchedSendQueueName, chSend)

		// bind the receive exchange with the receive queue
		bindReceiveQueue(chSend, qReceive)

		// start consuimng until terminated
		msgs, err := chReceive.Consume(qReceive.Name, "", true, false, false, false, nil)
		if err != nil {
			log.Fatal(err)
		}
		switch amqpReceiveFormat {
		case "plain", "PLAIN": // incoming logs
			for delivery := range msgs {
				match := getMatch(string(delivery.Body), patterns, trie, finalFor, regexes)
				if match.Type != "" {
					send([]byte(marshalJson(match)), chSend, qSend)
				} else {
					writeFile(noMatchOutputFile, string(delivery.Body)+"\r\n")
				}
			}
		case "json", "JSON": // incoming json
			for delivery := range msgs {
				m := unmarshalJson(delivery.Body)
				if attExists("@gomatch", m) { // att @gomatch is present
					if str, ok := m["@gomatch"].(string); ok {
						match := getMatch(str, patterns, trie, finalFor, regexes)
						if match.Type != "" {
							m["@type"] = match.Type
							m["@p"] = match.Body
							delete(m, "@gomatch")
							send([]byte(marshalJson(m)), chSend, qSend)
						}
					} else {
						log.Println("@gomatch is not a string (skipping)")
					}
				} else { // we return the former json msg
					send(delivery.Body, chSend, qSend)
				}
			}
		default:
			log.Fatal("Unknown RabbitMQ input format, use either plain or json.")
		}

	} else if *inputSocketFilePath != "none" { // socket
		conn := openSocket(*inputSocketFilePath)

		for {
			lines, eof := readFully(conn)
			for i := range lines {
				match := getMatch(lines[i], patterns, trie, finalFor, regexes)
				if match.Type != "" {
					writeFile(outputFile, marshalMatch(match)+"\r\n")
				} else {
					writeFile(noMatchOutputFile, lines[i]+"\r\n")
				}
			}
			if eof {
				break
			}
		}
		defer conn.Close()
	} else { // file, pipeline
		inputReader := openFile(*inputFilePath)

		for {
			line, eof := readLine(inputReader)
			logLine := string(line)
			match := getMatch(logLine, patterns, trie, finalFor, regexes)
			if match.Type != "" {
				writeFile(outputFile, marshalMatch(match)+"\r\n")
			} else {
				writeFile(noMatchOutputFile, logLine+"\r\n")
			}
			if eof {
				break
			}
		}
	}
	closeFile(outputFile)
	return
}

// watchPatterns performs re-reading of the first line in Patterns file
// (if it was recently modified).
// Then tries to add that line as a new pattern to trie.
func watchPatterns() {
	patternsFileInfo, err := os.Stat(*patternsFilePath)
	if err != nil {
		log.Fatal("watchPatterns(): ", err)
	}
	patternsLastModTime := patternsFileInfo.ModTime()
	for {
		time.Sleep(1 * time.Second)

		patternsFileInfo, err := os.Stat(*patternsFilePath)
		if err != nil {
			log.Println("watchPatterns(): ", err)
			break
		}

		if patternsLastModTime != patternsFileInfo.ModTime() {
			patternReader := openFile(*patternsFilePath)
			line, eof := readLine(patternReader)
			if !eof {
				oldLen := len(patterns)
				regexes, patterns = addPattern(string(line), patterns, regexes)

				if len(patterns) > oldLen {
					finalFor, state, patternNumber = appendPattern(patterns[len(patterns)-1], trie, finalFor, state, patternNumber, regexes)
					log.Println("new event: ", patterns[len(patterns)-1].Name)
				}
				patternsLastModTime = patternsFileInfo.ModTime()
			}
		}
	}
}
