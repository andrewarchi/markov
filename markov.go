package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const usage = `./markov words file [file]...`
const prevLen = 3

func main() {
	if len(os.Args) < 3 {
		fmt.Println(usage)
		return
	}
	l, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}
	rand.Seed(time.Now().UnixNano())

	textLen := int(l)
	files := os.Args[2:]
	wordMap := make(map[[prevLen]string][]string)

	for _, fname := range files {
		f, err := os.Open(fname)
		if err != nil {
			fmt.Println(err)
			continue
		}
		chainWords(readWords(f), wordMap)
	}

	fmt.Println(generateText(textLen, wordMap))
}

func readWords(r io.Reader) []string {
	var words []string
	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		lineWords := strings.Fields(scan.Text())
		if len(lineWords) > 0 {
			lineWords[len(lineWords)-1] += "\n"
		}
		words = append(words, lineWords...)
	}
	return words
}

func chainWords(words []string, wordMap map[[prevLen]string][]string) {
	var prev [prevLen]string
	for _, word := range words {
		if succ, exists := wordMap[prev]; exists {
			wordMap[prev] = append(succ, word)
		} else {
			wordMap[prev] = []string{word}
		}
		for i := 1; i < prevLen; i++ {
			prev[i-1] = prev[i]
		}
		prev[prevLen-1] = sanitizeWord(word)
	}
}

func generateText(textLen int, wordMap map[[prevLen]string][]string) string {
	var b strings.Builder
	var prev [prevLen]string
	for i := 0; ; i++ {
		var word string
		if next := wordMap[prev]; len(next) > 0 {
			word = next[rand.Intn(len(next))]
		}
		b.WriteString(word)
		if b.Len() > 0 && word != "" && word[len(word)-1] != '\n' {
			b.WriteByte(' ')
		}
		sanitized := sanitizeWord(word)
		if i > textLen && strings.HasSuffix(sanitized, ".") {
			break
		}
		for i := 1; i < prevLen; i++ {
			prev[i-1] = prev[i]
		}
		prev[prevLen-1] = sanitized
	}
	return strings.TrimSuffix(b.String(), " ")
}

func sanitizeWord(word string) string {
	if word == "" {
		return ""
	}
	if _, err := url.ParseRequestURI(word); err == nil {
		return "<URL>"
	}
	t := transform.Chain(norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		runes.Remove(runes.In(unicode.Punct)),
		norm.NFKC,
		runes.Map(unicode.ToLower))
	result, _, err := transform.String(t, word)
	if err != nil || result == "" {
		result = word
	}
	if r := []rune(word); len(r) > 0 {
		last := r[len(r)-1]
		if last == '\n' || unicode.Is(unicode.Sentence_Terminal, last) {
			return result + "."
		}
	}
	return result
}
