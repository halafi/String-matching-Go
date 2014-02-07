package main

import "strings"
import "log"

// Function createNewTrie()initializes new prefix tree. State is the 
// number of first created state, i is the number of first pattern to be
// appended.
func createNewTrie() (trie map[int]map[string]int, finalFor []int, state int, i int) {
	trie = make(map[int]map[string]int)
	finalFor = make([]int, 1)
	return trie, finalFor, 1, 1
}

// AppendPattern creates all the necessary transitions for given pattern
// to output trie.
func appendPattern(tokens map[string]string, pattern string, trie map[int]map[string]int, finalFor []int, state int, i int) (map[int]map[string]int, []int, int, int) {
	patternsNameSplit := strings.Split(pattern, "##") // we will ignore pattern name
	words := strings.Split(patternsNameSplit[1], " ")
	current := 0
	j := 0
	for j < len(words) && getTransition(current, words[j], trie) != -1 {
		current = getTransition(current, words[j], trie)
		j++
	}
	for j < len(words) {
		finalFor = intArraySizeUp(finalFor, 1)
		finalFor[state] = 0
		if len(getTransitionWords(current, trie)) > 0 && words[j][0] == '<' && words[j][len(words[j])-1] == '>' { // conflict check when adding regex transition
			transitionWords := getTransitionWords(current, trie)
			for w := range transitionWords {
				tokenWithoutBrackets := cutWord(1, len(words[j])-2, words[j])
				tokenWithoutBracketsSplit := strings.Split(tokenWithoutBrackets, ":")
				switch len(tokenWithoutBracketsSplit) {
				case 2:
					{
						if matchToken(tokens, tokenWithoutBracketsSplit[0], transitionWords[w]) {
							log.Fatal("pattern conflict: token \"" + words[j] + "\" matches word \"" + transitionWords[w] + "\"")
						}
					}
				case 1:
					{
						if matchToken(tokens, tokenWithoutBrackets, transitionWords[w]) {
							log.Fatal("pattern conflict: token \"" + words[j] + "\" matches word \"" + transitionWords[w] + "\"")
						}
					}
				default:
					log.Fatal("invalid token definition: \"<" + tokenWithoutBrackets + ">\"")
				}
			}
		} else if len(getTransitionTokens(current, trie)) > 0 && words[j][0] != '<' && words[j][len(words[j])-1] != '>' { //conflict check when adding word transition
			transitionTokens := getTransitionTokens(current, trie)
			for t := range transitionTokens {
				tokenWithoutBrackets := cutWord(1, len(transitionTokens[t])-2, transitionTokens[t])
				tokenWithoutBracketsSplit := strings.Split(tokenWithoutBrackets, ":")
				switch len(tokenWithoutBracketsSplit) {
				case 2:
					{
						if matchToken(tokens, tokenWithoutBracketsSplit[0], words[j]) {
							log.Fatal("pattern conflict: token \"" + transitionTokens[t] + "\" matches word \"" + words[j] + "\"")
						}
					}
				case 1:
					{
						if matchToken(tokens, tokenWithoutBrackets, words[j]) {
							log.Fatal("pattern conflict: token \"" + transitionTokens[t] + "\" matches word \"" + words[j] + "\"")
						}
					}
				default:
					log.Fatal("invalid token definition: \"<" + tokenWithoutBrackets + ">\"")
				}
			}
		}
		createTransition(current, words[j], state, trie)
		current = state
		j++
		state++
	}
	if finalFor[current] != 0 {
		log.Fatal("duplicate pattern detected: \"", pattern, "\"")
	} else {
		finalFor[current] = i // mark current state as terminal for pattern number i
	}
	i++ // increment pattern number
	return trie, finalFor, state, i
}

// Returns all transitioning tokens (without words) for a given 'state'
// in an automaton 'at' (map with stored states and their transitions)
// as an array of strings.
func getTransitionTokens(state int, at map[int]map[string]int) []string {
	transitionTokens := make([]string, 0)
	for s, _ := range at[state] {
		if s[0] == '<' && s[len(s)-1] == '>' {
			transitionTokens = append(transitionTokens, s)
		}
	}
	return transitionTokens
}

// Returns all transitioning words (without tokens) for a given 'state'
// in an automaton 'at' (map with stored states and their transitions)
// as an array of strings.
func getTransitionWords(state int, at map[int]map[string]int) []string {
	transitionWords := make([]string, 0)
	for s, _ := range at[state] {
		if s[0] != '<' && s[len(s)-1] != '>' {
			transitionWords = append(transitionWords, s)
		}
	}
	return transitionWords
}

// Returns an ending state for transition 'σ(fromState,overString)' in
// an automaton 'at' (map with stored states and their transitions).
// Returns '-1' if there is no transition.
func getTransition(fromState int, overString string, at map[int]map[string]int) int {
	if !stateExists(fromState, at) {
		return -1
	}
	toState, ok := at[fromState][overString]
	if ok == false {
		return -1
	}
	return toState
}

// If there is no state 'fromState', this function creates it, after
// that transitionion 'σ(fromState,overString) = toState' is created in
// an automaton 'at' (map with stored states and their transitions).
func createTransition(fromState int, overString string, toState int, at map[int]map[string]int) {
	if stateExists(fromState, at) {
		at[fromState][overString] = toState
	} else {
		at[fromState] = make(map[string]int)
		at[fromState][overString] = toState
	}
}

// Checks if a state 'state' exists in an automaton (map with stored
// states and their transitions) 'at'. Returns 'true' if it does,
// 'false' otherwise.
func stateExists(state int, at map[int]map[string]int) bool {
	_, ok := at[state]
	if !ok || state == -1 || at[state] == nil {
		return false
	}
	return true
}
