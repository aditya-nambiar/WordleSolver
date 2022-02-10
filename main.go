package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"
)

const WORD_FILE = "word_freq.json"
const WORDLE_LENGTH = 5

type Color int

const (
	Grey Color = iota
	Yellow
	Green
)

type WordleSolver struct {
	allWords            []string
	lastAcceptableIndex int
	Words               []string
	WordPopularity      map[string]float64
	greenChars          map[int]byte
	currCharSet         map[byte]int
	doesNotContain      map[byte]bool
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
		// p := false
		// if option[0] == 0 && option[1] == 0 {
		// 	if option[2] == 1 && option[3] == 1 && option[4] == 1 {
		// 		p = true
		// 	}
		// }
		for i := 0; i <= solver.lastAcceptableIndex; i++ {
			word := solver.allWords[i]
			cnt += checkMatch(option, inputWord, word)
			// if p && checkMatch(option, inputWord, word) == 1 {
			// 	fmt.Println(word)
			// }
		}
		if cnt == 0 {
			return 0
		}
		prob := float64(cnt) / float64(solver.lastAcceptableIndex+1)
		e := -1.0 * prob * math.Log2(prob)
		//fmt.Println(option, e, cnt)
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
	option := make([]Color, 5)
	return solver.generateAllPossibleOptions(option, 0, inputWord)
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
		if val > 1 {
			var cnt int
			for _, c := range word {
				if byte(c) == key {
					cnt++
				}
			}
			if cnt < val {
				return false
			}
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

func (solver *WordleSolver) pickWord() string {

	var allWords []wordEntropy

	for i := 0; i <= solver.lastAcceptableIndex; i++ {
		word := solver.allWords[i]
		if !solver.checkFeasibleWord(word) {
			solver.swap(i)
			i--
			continue
		}
	}

	fmt.Println("Number of words left - ", solver.lastAcceptableIndex)

	for i := 0; i <= solver.lastAcceptableIndex; i++ {
		word := solver.allWords[i]
		entropy := solver.calcEntropy(word)
		allWords = append(allWords, wordEntropy{word, entropy})
	}
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].entropy > allWords[j].entropy
	})
	fmt.Println("Picked max entropy ", allWords[0].word, " : ", allWords[0].entropy)
	fmt.Println(allWords[0])
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
			solver.doesNotContain[word[index]] = true
		}
	}
}

func interactiveMode() {

}

func testMode() {

}

func main() {
	solver := &WordleSolver{greenChars: make(map[int]byte), currCharSet: make(map[byte]int), doesNotContain: make(map[byte]bool), WordPopularity: make(map[string]float64)}
	solver.loadAllStrings()
	fmt.Println("Loaded ", len(solver.allWords), " strings")
	var result string
	firstGo := true
	for solver.isSolved() == false {
		var currGuess string
		if firstGo {
			currGuess = "tares"
			firstGo = false
		} else {
			currGuess = solver.pickWord()
		}

		fmt.Println("Guess - ", currGuess)
		fmt.Println("Enter Result")
		fmt.Scanln(&result)
		solver.addToState(currGuess, result)
	}

	fmt.Println("Congrats on solving the puzzle - ", result)
}
