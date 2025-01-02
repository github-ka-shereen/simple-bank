package utils

import (
	"math/rand"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var seededRand *rand.Rand

func init() {
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Example function using the local seeded random generator
func RandomInt(min, max int64) int64 {
	return seededRand.Int63n(max-min+1) + min
}

// RandomCamelCaseString generates a random camel-cased string with the specified number of words.
// Each word will have a length between minWordLength and maxWordLength.
func RandomCamelCaseString(numWords, minWordLength, maxWordLength int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"

	generateWord := func(length int) string {
		word := make([]byte, length)
		for i := range word {
			word[i] = charset[seededRand.Intn(len(charset))]
		}
		return string(word)
	}

	caser := cases.Title(language.English)
	var words []string
	for i := 0; i < numWords; i++ {
		wordLength := minWordLength + seededRand.Intn(maxWordLength-minWordLength+1)
		word := generateWord(wordLength)
		// Capitalize the first letter of each word for camel casing
		word = caser.String(word)
		words = append(words, word)
	}

	// Join words together to form the camel case string
	return strings.Join(words, "")
}

func RandomOwner() string {
	return RandomCamelCaseString(1, 6, 6)
}

func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

func RandomCurrency() string {
	currencies := []string{"USD","ZWL","EURO"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

