package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

type Result int

const (
	TooLow = iota
	TooHigh
	Equal
	None
)

type newGame struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type game struct {
	Id      string `json:"id"`
	Result  string `json:"result"`
	Ended   bool   `json:"ended"`
	answer  int
	Message string `json:"message"`
}

type guess struct {
	GameId string `json:"gameId"`
	Guess  int    `json:"guess"`
}

func (result Result) String() string {
	return [...]string{"TooLow", "TooHigh", "Equal", "None"}[result]
}

func (result Result) EnumIndex() int {
	return int(result)
}

var games = make(map[string]*game)

func createGameHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	newGame := &newGame{}
	err := decoder.Decode(newGame)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Println(newGame)
	gameId := uuid.NewString()
	answer := rand.Intn(newGame.Max-newGame.Min) + newGame.Min

	createdGame := &game{
		gameId,
		Result.String(None),
		false,
		answer,
		"Welcome to the game!",
	}
	games[gameId] = createdGame
	jsonResponse, _ := json.Marshal(createdGame)
	w.Write(jsonResponse)
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	guess := &guess{}
	err := decoder.Decode(guess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Println(guess)
	savedGame := games[guess.GameId]
	fmt.Println(savedGame)
	if savedGame.Ended {
		jsonResponse, _ := json.Marshal(savedGame)
		w.Write(jsonResponse)
		return
	}
	var result Result
	ended := false
	var message string
	if guess.Guess == savedGame.answer {
		result = Equal
		ended = true
		message = "Excelsior, you win!!"
	} else if guess.Guess > savedGame.answer {
		result = TooHigh
		message = "Fiddlesticks, that's a too high"
	} else {
		result = TooLow
		message = "Dingleberries, that's too low!"
	}
	updatedGame := &game{
		savedGame.Id,
		Result.String(result),
		ended,
		savedGame.answer,
		message,
	}
	games[savedGame.Id] = updatedGame
	jsonResponse, _ := json.Marshal(updatedGame)
	w.Write(jsonResponse)
}

func main() {
	http.HandleFunc("/game", createGameHandler)
	http.HandleFunc("/guess", guessHandler)

	log.Println("Go!")
	http.ListenAndServe(":8080", nil)
}
