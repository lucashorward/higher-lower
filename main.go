package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
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
	Min int `json:"min" validate:"required"`
	Max int `json:"max" validate:"required"`
}

type game struct {
	Id      string `json:"id"`
	Result  string `json:"result"`
	Ended   bool   `json:"ended"`
	answer  int
	Message string `json:"message"`
	min     int
	max     int
}

type guess struct {
	GameId string `json:"gameId" validate:"required"`
	Guess  int    `json:"guess" validate:"required"`
}

func (result Result) String() string {
	return [...]string{"TooLow", "TooHigh", "Equal", "None"}[result]
}

func (result Result) EnumIndex() int {
	return int(result)
}

// Saving the games in RAM means we're fast but also super super vulnerable to DOS attacks
// So this should probably not go online without throttling logic

// All games, stored as a map of <gameId, *game>
var games = make(map[string]*game)

// Core creation functionality.
// Creates a game, adds it to the games array, then returns it
//
// @throws if the Max is not higher than the Min value
func CreateGame(inputData *newGame) (*game, error) {
	if inputData.Min >= inputData.Max {
		return &game{}, errors.New("Max must be higher than the Min")
	}

	gameId := uuid.NewString()
	answer := rand.Intn(inputData.Max-inputData.Min) + inputData.Min

	createdGame := &game{
		gameId,
		Result.String(None),
		false,
		answer,
		"Welcome to the game!",
		inputData.Min,
		inputData.Max,
	}
	games[gameId] = createdGame
	return createdGame, nil
}

// CreateGame, wrapped into a HTTP endpoint
func createGameHandler(w http.ResponseWriter, r *http.Request, newGame *newGame) {
	createdGame, gameErr := CreateGame(newGame)
	if gameErr != nil {
		http.Error(w, gameErr.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse, _ := json.Marshal(createdGame)
	w.Write(jsonResponse)
}

/*
- Core guessing functionality.

- @throws if the provided gameId is not in the games map

- @throws if the guess is not between the min and max of the game
*/
func Guess(inputGuess *guess) (*game, error) {
	savedGame, exists := games[inputGuess.GameId]
	if !exists {
		return &game{}, errors.New("The requested game does not exist")
	}
	if savedGame.Ended {
		return savedGame, nil
	}
	if inputGuess.Guess > savedGame.max || inputGuess.Guess < savedGame.min {
		return &game{}, errors.New("Guess is out of bounds, must be between " + strconv.Itoa(savedGame.min) + " and " + strconv.Itoa(savedGame.max))
	}
	var result Result
	ended := false
	var message string
	if inputGuess.Guess == savedGame.answer {
		result = Equal
		ended = true
		message = "Excelsior, you win!!"
	} else if inputGuess.Guess > savedGame.answer {
		result = TooHigh
		message = "Fiddlesticks, that's too high"
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
		savedGame.min,
		savedGame.max,
	}
	games[savedGame.Id] = updatedGame
	return updatedGame, nil
}

func decodeAndValidate[T any](inputFunction func(http.ResponseWriter, *http.Request, T)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		object := new(T)
		err := decoder.Decode(object)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		validate := validator.New(validator.WithRequiredStructEnabled())
		validationError := validate.Struct(*object)
		if validationError != nil {
			http.Error(w, validationError.Error(), http.StatusBadRequest)
			return
		}
		inputFunction(w, r, *object)
	}
}

// Guess endpoint
func guessHandler(w http.ResponseWriter, r *http.Request, guess *guess) {
	game, err := Guess(guess)
	if err != nil {
		var status int
		if err.Error() == "The requested game does not exist" {
			status = http.StatusBadRequest
		} else {
			status = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), status)
		return
	}
	jsonResponse, _ := json.Marshal(game)
	w.Write(jsonResponse)
}

func allGamesHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, _ := json.Marshal(games)
	w.Write(jsonResponse)
}

func main() {
	log.Println("Starting, stand by")
	http.HandleFunc("POST /game", decodeAndValidate[*newGame](createGameHandler))
	http.HandleFunc("POST /guess", decodeAndValidate[*guess](guessHandler))
	http.HandleFunc("GET /games", allGamesHandler)

	log.Println("Server is up! Go play :)")
	http.ListenAndServe(":8080", nil)
}
