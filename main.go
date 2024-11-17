package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Flashcard represents a vocabulary word, its definition, its box and the date last reviewed.
type Flashcard struct {
	Word       string
	Definition string
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

		layout := "2006-01-02T15:04:05"

		if s, err := strconv.Atoi(line[2]); err == nil {

			if parsedTime, err := time.Parse(layout, line[3]); err == nil {

				card := Flashcard{
					Word:       line[0],
					Definition: line[1],
					Box:        s,
					LastReview: parsedTime,
				}
				flashcards = append(flashcards, card)
			}
		}

	}
	return flashcards, nil
}

// Saves flashcards back to the CSV file with updated progress
func SaveFlashcards(filename string, flashcards []Flashcard) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, card := range flashcards {
		record := []string{
			card.Word,
			card.Definition,
			strconv.Itoa(card.Box),
			card.LastReview.Format("2006-01-02T15:04:05"), // Format LastReview as a string
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

// Determines if a card should be reviewed today based on its last review date and box interval
func ShouldReview(card Flashcard) bool {
	daysSinceLastReview := int(time.Since(card.LastReview).Hours() / 24)
	nextReviewInterval := GetNextReviewInterval(card.Box)
	return daysSinceLastReview >= nextReviewInterval
}

// Presents a card's definition and prompts the user for the word.
// It returns whether they got it correct or not.
func ReviewCard(card Flashcard) bool {
	fmt.Printf("Definition: %s\n", card.Definition)
	fmt.Print("Your answer: ")
	var answer string
	fmt.Scanln(&answer)

	if strings.TrimSpace(strings.ToLower(answer)) == strings.ToLower(card.Word) {
		fmt.Println("Correct!")
		return true
	} else {
		fmt.Printf("Incorrect! The correct word was: %s\n", card.Word)
		return false
	}
}

// Updates the flashcard's box based on whether it was answered correctly or not.
func UpdateCard(card *Flashcard, correct bool) {
	if correct && card.Box < 5 {
		card.Box++ // Move to next box if correct and not already in Box 5
	} else if !correct {
		card.Box = 1 // Move back to Box 1 if incorrect
	}
	card.LastReview = time.Now() // Update last review time to now
}

func main() {
	const filename = "New_flashcards.csv"

	fmt.Println("Loading flashcards...")
	cards, err := LoadFlashcards(filename)
	if err != nil {
		log.Fatalf("Error loading flashcards: %v", err)
	}

	for {
		fmt.Println("\n--- Reviewing today's cards ---")

		var cardsToReview []Flashcard

		// Collect cards that need to be reviewed today based on their box and last review date.
		for i := range cards {
			if ShouldReview(cards[i]) {
				cardsToReview = append(cardsToReview, cards[i])
			}
		}

		if len(cardsToReview) == 0 {
			fmt.Println("No cards to review today.")
			break
		}

		for i := range cardsToReview {
			correct := ReviewCard(cardsToReview[i])
			UpdateCard(&cardsToReview[i], correct)
		}

		fmt.Println("Saving progress...")
		if err := SaveFlashcards(filename, cardsToReview); err != nil {
			log.Fatalf("Error saving progress: %v", err)
		}

		fmt.Println("All done for today! See you tomorrow.")

	}
}
