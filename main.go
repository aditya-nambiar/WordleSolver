package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strings"
)

const WORD_FILE = "data/word_freq.json"
const TEST_FILE = "data/test_words.txt"
const WORDLE_LENGTH = 5
const FIRST_GUESS = "corms"

// I for interactive mode or T for test mode
// Else it is single test mode
const MODE = "ST"

type Color int

const (
	Grey Color = iota
	Yellow
	Green
)

type Mode int

const (
	Interactive Mode = iota
	BatchTest
	SingleTest
)

type WordleSolver struct {
	allWords            []string
	lastAcceptableIndex int
	Words               []string
	WordPopularity      map[string]float64
	greenChars          map[int]byte
	currCharSet         map[byte]int
	knownCount          map[byte]int
	doesNotContain      map[byte]bool
	numTotalWords       int
}

func sigmoid(score float64) float64 {
	return 1.0 / (1 + math.Exp(-score))
}

func (solver *WordleSolver) loadAllStrings() {
	jsonFile, err := os.Open(WORD_FILE)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(byteValue, &solver.WordPopularity)
	for word, val := range solver.WordPopularity {
		solver.allWords = append(solver.allWords, word)
		solver.WordPopularity[word] = sigmoid(val)
		// if rand.Float64() < 0.01 {
		// 	fmt.Println(word, val, solver.WordPopularity[word])
		// }
	}

	solver.lastAcceptableIndex = len(solver.allWords) - 1
	solver.numTotalWords = len(solver.allWords)
}

// Function that tells if we have solved the puzzle
func (solver *WordleSolver) isSolved() bool {
	return len(solver.greenChars) == WORDLE_LENGTH
}

func checkMatch(option []Color, inputWord, currWord string) int {
	for index, val := range option {
		switch val {
		case Green:
			if currWord[index] != inputWord[index] {
				return 0
			}
		case Yellow:
			if !strings.Contains(currWord, string(inputWord[index])) {
				return 0
			}
			if currWord[index] == inputWord[index] {
				return 0
			}

		case Grey:
			if strings.Contains(currWord, string(inputWord[index])) {
				return 0
			}
		}
	}
	return 1
}

func (solver *WordleSolver) generateAllPossibleOptions(option []Color, index int, inputWord string) float64 {
	if index == WORDLE_LENGTH {
		cnt := 0

		for i := 0; i <= solver.lastAcceptableIndex; i++ {
			word := solver.allWords[i]
			cnt += checkMatch(option, inputWord, word)
		}
		if cnt == 0 {
			return 0
		}
		prob := float64(cnt) / float64(solver.lastAcceptableIndex+1)
		e := -1.0 * prob * math.Log2(prob)
		return e
	}

	var totalEntropy float64
	for i := 0; i < 3; i++ {
		option[index] = Color(i)
		totalEntropy += solver.generateAllPossibleOptions(option, index+1, inputWord)
	}
	return totalEntropy
}

func (solver *WordleSolver) calcEntropy(inputWord string) float64 {
	patternCount := make(map[string]int)
	for i := 0; i <= solver.lastAcceptableIndex; i++ {
		word := solver.allWords[i]
		pattern := getResult(word, inputWord)
		patternCount[pattern] += 1
		if pattern == "XXYXY" && inputWord == "saner" {
			fmt.Println(word)
		}
	}
	totalEntropy := 0.0
	for pattern, cnt := range patternCount {

		if pattern == "XXYXY" && inputWord == "saner" {
			fmt.Println(cnt)
		}
		prob := float64(cnt) / float64(solver.lastAcceptableIndex+1)
		totalEntropy += -1.0 * prob * math.Log2(prob)
	}

	return totalEntropy
}

// Check if this given word is even possible to exist given current state.
func (solver *WordleSolver) checkFeasibleWord(word string) bool {
	for key, val := range solver.greenChars {
		if word[key] != val {
			return false
		}
	}

	for key, _ := range solver.doesNotContain {
		if strings.Contains(word, string(key)) {
			return false
		}
	}

	for key, val := range solver.currCharSet {
		if !strings.Contains(word, string(key)) {
			return false
		}
		var cnt int
		for _, c := range word {
			if byte(c) == key {
				cnt++
			}
		}

		_, knowCnt := solver.knownCount[key]
		if cnt < val {
			return false
		} else if cnt > val && knowCnt && cnt > solver.knownCount[key] {
			return false
		}

	}

	return true
}

func (solver *WordleSolver) swap(i int) {
	temp := solver.allWords[i]
	solver.allWords[i] = solver.allWords[solver.lastAcceptableIndex]
	solver.allWords[solver.lastAcceptableIndex] = temp
	solver.lastAcceptableIndex--
}

type wordEntropy struct {
	word    string
	entropy float64
}

func (solver *WordleSolver) pickWord(m Mode, prevGuess string) string {

	var allWords []wordEntropy
	print := false
	if solver.lastAcceptableIndex < 50 {
		print = true
	}

	for i := 0; i <= solver.lastAcceptableIndex; i++ {
		word := solver.allWords[i]
		if print {
			fmt.Println(word)
		}
		if word == prevGuess || !solver.checkFeasibleWord(word) {
			solver.swap(i)
			i--
			continue
		}
	}
	if m != BatchTest {
		fmt.Println("Number of words left - ", solver.lastAcceptableIndex)
	}

	for i := 0; i <= solver.lastAcceptableIndex; i++ {
		word := solver.allWords[i]
		entropy := solver.calcEntropy(word)
		allWords = append(allWords, wordEntropy{word, entropy})
	}
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].entropy > allWords[j].entropy
	})
	if m != BatchTest {
		fmt.Println("Picked max entropy ", allWords[0].word, " : ", allWords[0].entropy)
	}
	return allWords[0].word
}

func (solver *WordleSolver) addToState(word, result string) {
	solver.Words = append(solver.Words, word)

	for index, c := range result {
		if c == 'Y' || c == 'G' {
			solver.currCharSet[word[index]] = 0
		}
	}

	for index, c := range result {
		switch c {
		case 'G':
			solver.greenChars[index] = word[index]
			solver.currCharSet[word[index]] += 1
		case 'Y':
			solver.currCharSet[word[index]] += 1
		case 'X':
			//  The X could be due to repeat character and hence
			// we need to first check the currCharSet
			if _, ok := solver.currCharSet[word[index]]; !ok {
				solver.doesNotContain[word[index]] = true
			} else {
				solver.knownCount[word[index]] = solver.currCharSet[word[index]]
			}
		}
	}
}

func (solver *WordleSolver) resetState() {
	solver.currCharSet = make(map[byte]int)
	solver.doesNotContain = make(map[byte]bool)
	solver.greenChars = make(map[int]byte)
	solver.knownCount = make(map[byte]int)
	solver.lastAcceptableIndex = solver.numTotalWords - 1
}

func interactiveMode(solver *WordleSolver) {
	var result, currGuess string
	firstGo := true
	for solver.isSolved() == false {
		if firstGo {
			currGuess = FIRST_GUESS
			firstGo = false
		} else {
			currGuess = solver.pickWord(Interactive, currGuess)
		}

		fmt.Println("Guess - ", currGuess)
		fmt.Println("Enter Result ( Format X for Grey, Y for Yellow & G for Green eg 'XYYXG') ")
		fmt.Scanln(&result)
		solver.addToState(currGuess, result)
	}

	fmt.Println("Congrats on solving the puzzle - ", currGuess)
}

func getResult(currGuess, answer string) string {
	var result string
	new_str := []byte(answer)

	for i := 0; i < WORDLE_LENGTH; i++ {
		if currGuess[i] == new_str[i] {
			result += "G"
			new_str[strings.Index(string(new_str), string(currGuess[i]))] = '#'
		} else if !strings.Contains(string(new_str), string(currGuess[i])) {
			result += "X"
		} else if strings.Contains(string(new_str), string(currGuess[i])) {
			result += "Y"
			new_str[strings.Index(string(new_str), string(currGuess[i]))] = '#'
		}
	}
	return result
}

func (solver *WordleSolver) solveWordle(m Mode, answer string) int {
	solver.resetState()
	var numTries int
	firstGo := false
	fmt.Println("Trying to guess word - ", answer)
	var result, currGuess string

	for solver.isSolved() == false {
		numTries += 1
		if firstGo {
			currGuess = FIRST_GUESS
			firstGo = false
		} else {
			currGuess = solver.pickWord(m, currGuess)
		}
		result = getResult(currGuess, answer)
		// fmt.Println("Guess - ", currGuess, result)
		solver.addToState(currGuess, result)
		break
	}
	fmt.Println("Tries ", numTries)
	return numTries
}

func testMode(solver *WordleSolver) {
	file, err := os.Open(TEST_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var totalScore, numTestWords int
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		numTestWords += 1
		if numTestWords%101 == 0 {
			fmt.Println("Current Avg Score", float64(totalScore)/float64(numTestWords))
		}
		word := scanner.Text()
		numTries := solver.solveWordle(BatchTest, word)
		totalScore += numTries
	}

	fmt.Println("Avg Score ", float64(totalScore)/float64(numTestWords))
}

func main() {
	solver := &WordleSolver{}
	solver.resetState()
	solver.loadAllStrings()
	fmt.Println("Loaded ", len(solver.allWords), " strings")

	if MODE == "I" {
		interactiveMode(solver)
	} else if MODE == "T" {
		testMode(solver)
	} else { // Single Test Case
		var correctString string
		fmt.Println("Enter Correct String - ")
		fmt.Scanln(&correctString)
		solver.solveWordle(SingleTest, correctString)
	}
}
