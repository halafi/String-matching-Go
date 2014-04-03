package main

import "log"

// Match represents a single log event matched.
type Match struct {
	Type string            // event name
	Body map[string]string // key=matched_token, value=matched_value
}

// getMatch returns a match for a log line. It goes through all of the
// log line words, one by one, changing state using the given trie.
// Word transitions are prioritized over regex transitions.
// If a final state is reached for some pattern after matching the last
// log word, then a match with matched data is returned.
// Otherwise an empty match is returned.
func getMatch(logLine string, patterns []Pattern, trie map[int]map[Token]int, finalFor []int, regexMap map[string]Regex) Match {
	match := Match{}
	matchBody := make(map[string]string)

	current := 0
	logWords := logLineSplit(logLine)

	for i := range logWords {
		transitionTokens := getTransitionRegexes(current, trie)
		validTokens := 0

		if getTransition(current, Token{false, logWords[i], ""}, trie) != -1 {
			current = getTransition(current, Token{false, logWords[i], ""}, trie)
		} else if len(transitionTokens) > 0 {
			for t := range transitionTokens {
				if (regexMap[transitionTokens[t].Value].Compiled).MatchString(logWords[i]) {
					validTokens++
					current = getTransition(current, transitionTokens[0], trie)
					matchBody[transitionTokens[0].OutputName] = logWords[i]
				}
			}
			if validTokens > 1 {
				log.Fatal("multiple acceptable tokens for one word: \"" + logWords[i] + "\"")
			}
		} else {
			break
		}

		if finalFor[current] != 0 && i == len(logWords)-1 {
			if len(matchBody) > 0 {
				match = Match{patterns[finalFor[current]-1].Name, matchBody}
			} else {
				match = Match{patterns[finalFor[current]-1].Name, nil}
			}
		}
	}
	return match
}

// logLineSplit splits a single log line into words.
// Words can be separated by a single space or any amount of spaces.
func logLineSplit(line string) []string {
	words := make([]string, 0)
	if line == "" {
		return words
	}
	words = append(words, "")
	wordIndex := 0
	chars := []uint8(line)
	for c := range chars {
		if chars[c] == ' ' && c < len(chars)-1 {
			if words[wordIndex] != "" {
				words = append(words, "")
				wordIndex++
			}
		} else if chars[c] != ' ' {
			words[wordIndex] = words[wordIndex] + string(chars[c])
		}
	}
	return words
}
