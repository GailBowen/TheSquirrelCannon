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
// Updates flashcards in a CSV file without clearing other records
func SaveFlashcards(filename string, flashcards []Flashcard) error {
	// Step 1: Load existing records from the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	existingRecords, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// Step 2: Create a map of flashcards for quick lookup by word
	flashcardMap := make(map[string]Flashcard)
	for _, card := range flashcards {
		flashcardMap[card.Word] = card
	}

	// Step 3: Update existing records with new data from flashcards
	for i, record := range existingRecords {
		if len(record) < 4 {
			continue // Skip malformed lines
		}

		word := record[0]
		if updatedCard, exists := flashcardMap[word]; exists {
			// Update Box and LastReview fields if this card was reviewed
			existingRecords[i][2] = strconv.Itoa(updatedCard.Box)
			existingRecords[i][3] = updatedCard.LastReview.Format("2006-01-02T15:04:05")
			delete(flashcardMap, word) // Remove from map once updated
		}
	}

	// Step 4: Append any new flashcards that were not found in the existing records
	for _, card := range flashcards {
		if _, exists := flashcardMap[card.Word]; exists {
			newRecord := []string{
				card.Word,
				card.Definition,
				strconv.Itoa(card.Box),
				card.LastReview.Format("2006-01-02T15:04:05"),
			}
			existingRecords = append(existingRecords, newRecord)
		}
	}

	// Step 5: Write all records (updated and unchanged) back to the CSV file
	file, err = os.Create(filename) // This truncates (empties) the file before writing
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range existingRecords {
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

	for {

		fmt.Println("Loading flashcards...")
		cards, err := LoadFlashcards(filename)
		if err != nil {
			log.Fatalf("Error loading flashcards: %v", err)
		}

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
