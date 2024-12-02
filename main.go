package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Flashcard represents a vocabulary question, its answer, its box and the date last reviewed.
type Flashcard struct {
	Question   string
	Answer     string
	Box        int       // Leitner box number (1 to 5)
	LastReview time.Time // Last time this card was reviewed
}

// Loads flashcards from a CSV file
func LoadFlashcards(filename string) ([]Flashcard, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var flashcards []Flashcard
	for _, line := range lines {
		if len(line) < 2 {
			continue // skip malformed lines
		}

		layout := "2006-01-02"

		if s, err := strconv.Atoi(line[2]); err == nil {

			if parsedTime, err := time.Parse(layout, line[3]); err == nil {

				card := Flashcard{
					Question:   line[0],
					Answer:     line[1],
					Box:        s,
					LastReview: parsedTime,
				}
				flashcards = append(flashcards, card)
			}
		}

	}
	return flashcards, nil
}

// Saves all flashcards back to the CSV file by clearing the file and writing all records.
func SaveFlashcards(filename string, flashcards []Flashcard) error {
	file, err := os.Create(filename) // This truncates (empties) the file before writing
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, card := range flashcards {
		record := []string{
			card.Question,
			card.Answer,
			strconv.Itoa(card.Box),
			card.LastReview.Format("2006-01-02"),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// Returns the number of days until the next review based on the box number
func GetNextReviewInterval(box int) int {
	switch box {
	case 1:
		return 1 // Every day
	case 2:
		return 2 // Every other day
	case 3:
		return 4 // Every 4 days
	case 4:
		return 7 // Every week
	case 5:
		return 14 // Every two weeks
	default:
		return 1 // Default to daily if something goes wrong
	}
}

// Determines if a card should be reviewed today based on its last review date and box interval.
// Takes an additional 'dateToUse' parameter for testing purposes.
func ShouldReview(card Flashcard, dateToUse time.Time) bool {
	daysSinceLastReview := int(dateToUse.Sub(card.LastReview).Hours() / 24)
	nextReviewInterval := GetNextReviewInterval(card.Box)
	return daysSinceLastReview >= nextReviewInterval
}

// Presents a card's question and prompts the user for the answer.
// First bool return: Is it correct?
// Second bool return: The user typed STOP
func ReviewCard(card Flashcard) (bool, bool) {
	fmt.Printf("Question: %s\n", card.Question)
	fmt.Print("Your answer (type STOP to exit): ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = answer[:len(answer)-1]

	if strings.TrimSpace(strings.ToUpper(answer)) == "STOP" {
		return false, true // false for incorrect, true for STOP!
	}

	if strings.TrimSpace(strings.ToLower(answer)) == strings.ToLower(card.Answer) {
		fmt.Println("Correct!")
		return true, false
	} else {
		fmt.Printf("Incorrect! The correct answer was: %s\n", card.Answer)
		return false, false
	}
}

// Updates the flashcard's box based on whether it was answered correctly or not.
// Takes an additional 'dateToUse' parameter for testing purposes.
func UpdateCard(card *Flashcard, correct bool, dateToUse time.Time) {
	if correct && card.Box < 5 {
		card.Box++ // Move to next box if correct and not already in Box 5
	} else if !correct {
		card.Box = 1 // Move back to Box 1 if incorrect
	}
	card.LastReview = dateToUse // Update last review time to 'dateToUse'
}

func main() {
	//const filename = "New_flashcards.csv"
	//const filename = "C:\\Users\\gailb\\Documents\\Code\\TheSquirrelCannon\\Cards\\Languages\\Human\\Latin\\MrGwynneTmr.csv"
	const filename = "C:\\Users\\gailb\\Documents\\Code\\TheSquirrelCannon\\Cards\\Languages\\Human\\Italian\\Italian.csv"

	mode := os.Getenv("APP_MODE")
	var dateToUse time.Time

	if mode == "test" {
		fmt.Println("\nhello tester!")

		fmt.Println("\nEnter the date you want it to be (format: YYYY-MM-DD):")

		reader := bufio.NewReader(os.Stdin)

		testDateStr, _ := reader.ReadString('\n')

		testDateStr = strings.TrimSpace(testDateStr)

		layout := "2006-01-02"

		var err error

		dateToUse, err = time.Parse(layout, testDateStr)

		if err != nil {
			fmt.Println("Invalid input. Using today's date.")
			dateToUse = time.Now()
		} else {
			fmt.Println("\nUsing test date:", testDateStr)
		}
	} else {
		// Use today's date in normal mode.
		dateToUse = time.Now()
	}

	for {

		fmt.Println("Loading flashcards...")

		cards, err := LoadFlashcards(filename)

		if err != nil {
			log.Fatalf("Error loading flashcards: %v", err)
			break
		}

		fmt.Println("\n--- Reviewing today's cards ---")

		hasCardsToReview := false

		// Review each card that needs to be reviewed today.
		for i := range cards {
			if ShouldReview(cards[i], dateToUse) {
				hasCardsToReview = true

				correct, stopNow := ReviewCard(cards[i])

				if stopNow {
					break
				}

				UpdateCard(&cards[i], correct, dateToUse)
			}
		}

		if !hasCardsToReview {
			fmt.Println("No cards to review today.")
			break
		}

		fmt.Println("Saving progress...")

		if err := SaveFlashcards(filename, cards); err != nil {
			log.Fatalf("Error saving progress: %v", err)
		}

		fmt.Println("All done for today! See you tomorrow.")
	}
}
