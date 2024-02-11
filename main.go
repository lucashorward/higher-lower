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
	Id     string `json:"id"`
	Result string `json:"result"`
	Ended  bool   `json:"ended"`
	answer int
}

type guess struct {
	GameId string `json:"gameId"`
	Guess  int    `json:"guess"`
}

// String - Creating common behavior - give the type a String function
func (d Result) String() string {
	return [...]string{"TooLow", "TooHigh", "Equal", "None"}[d]
}

// EnumIndex - Creating common behavior - give the type a EnumIndex functio
func (d Result) EnumIndex() int {
	return int(d)
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
	if guess.Guess == savedGame.answer {
		result = Equal
		ended = true
	} else if guess.Guess > savedGame.answer {
		result = TooHigh
	} else {
		result = TooLow
	}
	updatedGame := &game{
		savedGame.Id,
		Result.String(result),
		ended,
		savedGame.answer,
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
