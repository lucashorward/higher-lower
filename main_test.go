package main

import (
	"testing"
)

func TestGameCreation(t *testing.T) {
	input := newGame{
		Min: 0,
		Max: 100,
	}
	response, err := CreateGame(&input)
	if err != nil {
		t.Fatalf(`CreateGame = %v, %q, don't want an error`, response, err)
	}
	if response.Result != "None" {
		t.Fatalf(`CreateGame = %v, want "None"`, response)
	}
	if response.Id == "" {
		t.Fatalf(`CreateGame = %v, want a filled in game ID"`, response)
	}
	if response.Ended == true {
		t.Fatalf(`CreateGame = %v, want the ended to be preset to false"`, response)
	}
	if response.answer > 100 || response.answer < 0 {
		t.Fatalf(`CreateGame = %v, want the answer to be in the given min and max"`, response)
	}
	if response.Message != "Welcome to the game!" {
		t.Fatalf(`CreateGame = %v, expect the default message"`, response)
	}
}

func TestGameCreationInvalidBoundaries(t *testing.T) {
	input := newGame{
		Min: 100,
		Max: 0,
	}
	response, err := CreateGame(&input)
	if err == nil {
		t.Fatalf(`CreateGame = %v, %q, want error to be returned`, response, err)
	}
	if err.Error() != "Max must be higher than the Min" {
		t.Fatalf(`CreateGame = %v, %q, want error to be correct`, response, err)
	}
}

func TestGameCreationEqualBoundaries(t *testing.T) {
	input := newGame{
		Min: 100,
		Max: 100,
	}
	response, err := CreateGame(&input)
	if err == nil {
		t.Fatalf(`CreateGame = %v, %q, want error to be returned`, response, err)
	}
	if err.Error() != "Max must be higher than the Min" {
		t.Fatalf(`CreateGame = %v, %q, want error to be correct`, response, err)
	}
}

func TestGuessNonExistingGame(t *testing.T) {
	input := &guess{
		GameId: "Fake",
		Guess:  5000,
	}
	response, err := Guess(input)
	if err == nil {
		t.Fatalf(`Guess = %v, %q, want error to be returned, error`, response, err)
	}
	if err.Error() != "400 The requested game does not exist" {
		t.Fatalf("Error must be about the game not existing, but was %q", err)
	}
}

func TestGuessOutOfBounds(t *testing.T) {
	input := newGame{
		Min: 0,
		Max: 100,
	}
	response, _ := CreateGame(&input)
	guessInput := &guess{
		GameId: response.Id,
		Guess:  5000,
	}
	response, err := Guess(guessInput)
	if err == nil {
		t.Fatalf(`Guess = %v, %q, want error to be returned`, response, err)
	}
	if err.Error() != "400 Guess is out of bounds, must be between 0 and 100" {
		t.Fatalf("Guess Error must be about the guess being Out Of Bounds, but was %q", err)
	}
}
