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
const FIRST_GUESS = "tares"
const POPULARITY_WEIGHT = 1.5
const GOAL_RESULT = "GGGGG"

// I for interactive mode or T for test mode
// Else it is single test mode
const MODE = "T"

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
	numTotalWords       int
}

func sigmoid(score float64) float64 {
	return 1.0 / (1 + math.Exp(-score))
}

type Pair struct {
	Key   string
	Value float64
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

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
		solver.WordPopularity[word] = sigmoid(math.Log10(val) + 6.3)
	}

	solver.lastAcceptableIndex = len(solver.allWords) - 1
	solver.numTotalWords = len(solver.allWords)
}

func (solver *WordleSolver) calcEntropy(inputWord string) float64 {
	patternCount := make(map[string]int)
	for i := 0; i <= solver.lastAcceptableIndex; i++ {
		word := solver.allWords[i]
		pattern := getResult(inputWord, word)
		patternCount[pattern] += 1
	}
	totalEntropy := 0.0
	for _, cnt := range patternCount {
		prob := float64(cnt) / float64(solver.lastAcceptableIndex+1)
		totalEntropy += -1.0 * prob * math.Log2(prob)
	}

	return totalEntropy
}

func (solver *WordleSolver) swap(i int) {
	temp := solver.allWords[i]
	solver.allWords[i] = solver.allWords[solver.lastAcceptableIndex]
	solver.allWords[solver.lastAcceptableIndex] = temp
	solver.lastAcceptableIndex--
}

type wordEntropy struct {
	word string
	// Calculated as E[Info] + W * P(word)
	score      float64
	entropy    float64
	popularity float64
}

func (solver *WordleSolver) pickWord(m Mode, prevGuess string, prevResult string, weight float64) string {

	var allWords []wordEntropy
	// print := false
	// if solver.lastAcceptableIndex < 50 {
	// 	print = true
	// }

	for i := 0; i <= solver.lastAcceptableIndex; i++ {
		word := solver.allWords[i]
		// if print {
		// 	fmt.Println(word)
		// }
		if word == prevGuess || getResult(prevGuess, word) != prevResult {
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
		allWords = append(allWords, wordEntropy{word, entropy + (weight * solver.WordPopularity[word]), entropy, solver.WordPopularity[word]})
	}
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].score > allWords[j].score
	})

	// if len(allWords) < 10 {
	// 	fmt.Println(allWords)
	// }
	if m != BatchTest {
		fmt.Println("Picked max score ", allWords[0])
	}
	return allWords[0].word
}

func (solver *WordleSolver) resetState() {
	solver.lastAcceptableIndex = solver.numTotalWords - 1
}

func interactiveMode(solver *WordleSolver) {
	var result, currGuess string
	firstGo := true
	for result != GOAL_RESULT {
		if firstGo {
			currGuess = FIRST_GUESS
			firstGo = false
		} else {
			currGuess = solver.pickWord(Interactive, currGuess, result, POPULARITY_WEIGHT)
		}

		fmt.Println("Guess - ", currGuess)
		fmt.Println("Enter Result ( Format X for Grey, Y for Yellow & G for Green eg 'XYYXG') ")
		fmt.Scanln(&result)
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

func (solver *WordleSolver) solveWordle(m Mode, answer string, weight float64) int {
	solver.resetState()
	var numTries int
	firstGo := true
	if m != BatchTest {
		fmt.Println("Trying to guess word - ", answer)
	}
	var result, currGuess string

	for result != GOAL_RESULT {
		numTries += 1
		if firstGo {
			currGuess = FIRST_GUESS
			firstGo = false
		} else {
			currGuess = solver.pickWord(m, currGuess, result, weight)
		}
		result = getResult(currGuess, answer)
	}
	if m != BatchTest {
		fmt.Println("Tries ", numTries)
	}
	return numTries
}

func testMode(solver *WordleSolver, weight float64) float64 {
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
		// if numTestWords%101 == 0 {
		// 	fmt.Println("Current Avg Score", float64(totalScore)/float64(numTestWords))
		// }
		word := scanner.Text()
		numTries := solver.solveWordle(BatchTest, word, weight)
		totalScore += numTries
	}

	fmt.Println("Avg Score ", float64(totalScore)/float64(numTestWords))
	return float64(totalScore) / float64(numTestWords)
}

func main() {
	solver := &WordleSolver{}
	solver.resetState()
	solver.loadAllStrings()
	fmt.Println("Loaded ", len(solver.allWords), " strings")

	if MODE == "I" {
		interactiveMode(solver)
	} else if MODE == "T" {
		// w := make([]float64, 21)
		// for i := 0; i <= 20; i++ {
		// 	w[i] = float64(i) / 10.0
		// }
		// for _, w := range w {
		// 	fmt.Println("Trying weight w, ", w)
		avg_result := testMode(solver, 1.0)
		fmt.Println(0.0, avg_result)
		// }
	} else { // Single Test Case
		var correctString string
		fmt.Println("Enter Correct String - ")
		fmt.Scanln(&correctString)
		solver.solveWordle(SingleTest, correctString, POPULARITY_WEIGHT)
	}
}
