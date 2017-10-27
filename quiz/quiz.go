package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	flagFilePath  string
	flagTimeLimit int
	wg            sync.WaitGroup
)

func init() {
	flag.StringVar(&flagFilePath, "file", "questions.csv", "path/to/file.csv")
	flag.IntVar(&flagTimeLimit, "limit", 30, "Quiz time limit")
	flag.Parse()
}

func handleErrors(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

func getAbsPath(p string) (absPath string) {
	absPath, err := filepath.Abs(p)
	handleErrors(err)

	return
}

func getFile(p string) (f *os.File) {
	f, err := os.Open(p)
	handleErrors(err)

	return
}

func readFile(f *os.File) (c [][]string) {
	fc := csv.NewReader(f)
	c, err := fc.ReadAll()
	handleErrors(err)

	return
}

func askQuestion(w io.Writer, r io.Reader, question string, replyTo chan string) {
	reader := bufio.NewReader(r)
	fmt.Fprintln(w, "Question: "+question)
	fmt.Fprint(w, "Answer: ")
	ans, err := reader.ReadString('\n')
	if err != nil {
		close(replyTo)
		if err == io.EOF {
			return
		}
		log.Fatalln(err)
	}

	replyTo <- ans
}

func checkAnswer(ans string, expected string) (c bool) {
	c = strings.EqualFold(strings.TrimSpace(ans), strings.TrimSpace(expected))

	return
}

func summary(correct int, totalQuestions int) {
	fmt.Fprintf(os.Stdout, "You answered %d questions correctly: (%d / %d)", correct, correct, totalQuestions)
}

func main() {
	csvPath := getAbsPath(flagFilePath)
	csvFile := getFile(csvPath)

	defer func() {
		err := csvFile.Close()
		handleErrors(err)
	}()

	csvData := readFile(csvFile)
	totalQuestions := len(csvData)
	questions := make(map[int]string, totalQuestions)
	answers := make(map[int]string, totalQuestions)
	responses := make(map[int]string, totalQuestions)

	for i, data := range csvData {
		questions[i] = data[0]
		answers[i] = data[1]
	}

	respondTo := make(chan string)

	fmt.Println("Press [Enter] to start.")
	bufio.NewScanner(os.Stdout).Scan()

	wg.Add(1)
	timer := time.NewTimer(time.Duration(flagTimeLimit) * time.Second)

	go func() {
	label:
		for i := 0; i < totalQuestions; i++ {
			go askQuestion(os.Stdout, os.Stdin, questions[i], respondTo)
			select {
			case <-timer.C:
				fmt.Fprintln(os.Stderr, "\n Time's up!!!!")
				break label
			case ans, ok := <-respondTo:
				if ok {
					responses[i] = ans
				} else {
					break label
				}
			}
		}
		wg.Done()
	}()
	wg.Wait()

	correct := 0
	for i := 0; i < totalQuestions; i++ {
		if checkAnswer(responses[i], answers[i]) {
			correct++
		}
	}

	summary(correct, totalQuestions)
}
