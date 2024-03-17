package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	fmt "fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"

	pb "example.com/higher-lower-game/higherlower"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type Result int

const (
	TooLow = iota
	TooHigh
	Equal
	None
)

type HandlerError struct {
	error
	StatusCode int
}

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

/*
- Core creation functionality.

- Creates a game, adds it to the games array, then returns it

- @throws `HandlerError` if the Max is not higher than the Min value
*/
func CreateGame(inputData *newGame) (*game, error) {
	if inputData.Min >= inputData.Max {
		return &game{}, HandlerError{
			errors.New("max must be higher than the Min"),
			http.StatusBadRequest,
		}
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

/*
- Core guessing functionality.

- @throws `HandlerError` if the provided gameId is not in the games map

- @throws `HandlerError` if the guess is not between the min and max of the game
*/
func Guess(inputGuess *guess) (*game, error) {
	savedGame, exists := games[inputGuess.GameId]
	if !exists {
		return &game{}, HandlerError{
			errors.New("the requested game does not exist"),
			http.StatusNotFound,
		}
	}
	if savedGame.Ended {
		return savedGame, nil
	}
	if inputGuess.Guess > savedGame.max || inputGuess.Guess < savedGame.min {
		return &game{}, HandlerError{
			errors.New("guess is out of bounds, must be between " + strconv.Itoa(savedGame.min) + " and " + strconv.Itoa(savedGame.max)),
			http.StatusBadRequest,
		}
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

/*
* This wraps http calls to prevent duplicate code. This function does:

* JSON decoding

* Validation based on `go-playground/validator/v10`

  - Error handling:
    returns the provided status code if the Error is a HandlerError (which has a statusCode)
    returns InternalServerError if an error is returned without statusCode
*/
func httpWrapper[T any, K any](inputFunction func(T) (K, error)) func(http.ResponseWriter, *http.Request) {
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
		output, functionError := inputFunction(*object)
		if functionError != nil {
			if handlerError, ok := functionError.(HandlerError); ok {
				http.Error(w, handlerError.Error(), handlerError.StatusCode)
			} else {
				http.Error(w, functionError.Error(), http.StatusInternalServerError)
			}
			return
		}
		jsonResponse, _ := json.Marshal(output)
		w.Write(jsonResponse)
	}
}

// Simply returns all games in a JSON response (no wrapper needed)
func allGamesHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse, _ := json.Marshal(games)
	w.Write(jsonResponse)
}

type server struct {
	pb.UnimplementedHigherLowerGameServer
}

func (s *server) GetGameState(_ context.Context, _ *pb.GameState) (*pb.GameState, error) {
	protoGames := make([]*pb.Game, len(games))
	for _, game := range games {
		protoGames = append(protoGames, &pb.Game{
			GameId: game.Id,
			Result: game.Result,
			Ended:  game.Ended,
			Answer: int32(game.answer),
			Min:    int32(game.min),
			Max:    int32(game.max),
		})
	}
	return &pb.GameState{Game: protoGames}, nil
}

func createMultipleGamesOnStartup() {
	for i := 0; i < 10; i++ {
		newGame := &newGame{
			Min: 0,
			Max: 100,
		}
		CreateGame(newGame)
	}
}

func main() {
	createMultipleGamesOnStartup()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterHigherLowerGameServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Println("Starting, stand by")
	http.HandleFunc("POST /game", httpWrapper[*newGame, *game](CreateGame))
	http.HandleFunc("POST /guess", httpWrapper[*guess, *game](Guess))
	http.HandleFunc("GET /games", allGamesHandler)

	log.Println("Server is up! Go play :)")
	http.ListenAndServe(":8080", nil)
}
