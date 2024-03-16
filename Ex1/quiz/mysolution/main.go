package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Question struct {
	prompt string
	answer string
	result Result
}

type Result struct {
	recordedAnswer string
}

func (q *Question) showQuestion(resChannel chan string, timer *time.Timer, onTimerExpires func()) {
	go func() {
		fmt.Printf("Question: %s\nAnswer: ", q.prompt)
		var userAnswer string
		fmt.Scanln(&userAnswer)
		resChannel <- userAnswer
	}()

	select {
	case x := <-resChannel:
		{
			q.result = Result{x}
		}
	case <-timer.C:
		{
			onTimerExpires()
		}
	}
}

func (q *Question) compareAnswers() bool {
	return strings.EqualFold(strings.TrimSpace(q.answer), strings.TrimSpace(q.result.recordedAnswer))
}

type QuestionSet struct {
	questions []Question
	result    SetResult
}

type SetResult struct {
	questionCount int
	answeredCount int
	correctCount  int
}

func (set *QuestionSet) shuffle(nRound int) {
	now := time.Now().Unix()
	random := rand.New(rand.NewSource(now))

	n := len(set.questions)
	for i := 0; i < nRound; i++ {
		if x, y := random.Intn(n), random.Intn(n); x != y {
			set.questions[x], set.questions[y] = set.questions[y], set.questions[x]
		}
	}

}

func (set *QuestionSet) startQuiz(timeLimit int) {
	n := len(set.questions)

	timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
	response := make(chan string, 1)
	defer close(response)

	timesUp := false
	for i := 0; !timesUp && i < n; i++ {
		set.questions[i].showQuestion(response, timer, func() { timesUp = true })
	}

	if timesUp {
		fmt.Printf("\n\n--------------- TIME'S UP!!! -------------------\n\n")
	}

	set.judge()
	set.printResult()
}

func (set *QuestionSet) judge() {
	questionCount := len(set.questions)
	var correctCount = 0
	var answeredCount = 0
	for i := 0; i < questionCount; i++ {
		if len(set.questions[i].result.recordedAnswer) > 0 {
			answeredCount++
			if set.questions[i].compareAnswers() {
				correctCount++
			}
		}
	}
	set.result = SetResult{
		questionCount: questionCount,
		answeredCount: answeredCount,
		correctCount:  correctCount,
	}
}

func (set *QuestionSet) printResult() {
	fmt.Println("--------------------------------------")
	fmt.Printf("%-10s: %-5d\n", "Total", set.result.questionCount)
	fmt.Printf("%-10s: %-5d\n", "Answered", set.result.answeredCount)
	fmt.Printf("%-10s: %-5d\n", "Correct", set.result.correctCount)
	fmt.Println("--------------------------------------")
}

type QuestionsReader struct {
	filePath string
}

func (reader *QuestionsReader) Read() (QuestionSet, error) {
	file, err := os.OpenFile(reader.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return QuestionSet{}, err
	}
	defer file.Close()

	var questions []Question
	csvReader := csv.NewReader(file)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return QuestionSet{}, err
		}
		if len(record) != 2 {
			return QuestionSet{}, errors.New("invalid csv format, expected 2 columns")
		}
		questions = append(questions, Question{
			prompt: record[0],
			answer: record[1],
			result: Result{""},
		})
	}
	return QuestionSet{
		questions: questions,
	}, nil
}

func main() {
	outFilePtr := flag.String("questionFile", "../problems.csv", "The path to csv file containing questions")
	timeLimitPtr := flag.Int("timeLimit", 30, "Maximum time allowed in seconds")
	shufflePtr := flag.Bool("shuffle", false, "Whether to shuffle the questions set or not")

	flag.Parse()

	questionReader := QuestionsReader{
		filePath: *outFilePtr,
	}

	questionSet, err := questionReader.Read()
	if err != nil {
		log.Fatal(err)
	}

	if *shufflePtr {
		questionSet.shuffle(20)
	}

	fmt.Println("Press enter to start")
	fmt.Scanln()

	questionSet.startQuiz(*timeLimitPtr)

}
