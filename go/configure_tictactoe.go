// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"

	"tictactoe/models"
	"tictactoe/restapi/operations"
)

//go:generate swagger generate server --target ../../homework-tictactoe --name Tictactoe --spec ../tictactoe.yaml --principal interface{}

func configureFlags(api *operations.TictactoeAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

var games []*models.Game

func remove(s []*models.Game, i int) []*models.Game {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func move(table string, move string) (string, string) {
	var changecounter int
	movearr := []rune(move)
	tablearr := []rune(table)

	if len(movearr) != 9 {
		log.Printf("1")
		return "ERROR", table
	}
	for i := 0; i < len(tablearr); i++ {
		if tablearr[i] != movearr[i] {
			changecounter++
			if movearr[i] != 'X' || (movearr[i] == 'X' && tablearr[i] == 'O') {
				log.Printf("2")
				return "ERROR", table
			}
		}
	}
	if changecounter > 1 || changecounter == 0 {
		log.Printf("3")
		return "ERROR", table
	}
	status := checkend(movearr)
	log.Printf("THE STATUS: " + status)
	if status != "RUNNING" {
		return status, move
	}
	pcstatus, pcmovearr := pcmove(movearr)
	pcmovestring := string([]rune(pcmovearr))
	if pcstatus != "RUNNING" {
		return pcstatus, pcmovestring
	}
	finalstatus := checkend(pcmovearr)
	return finalstatus, pcmovestring
}

func checkend(movearr []rune) string {
	current := movearr[0]
	if current != '-' {
		if current == movearr[1] && current == movearr[2] {
			return string(current) + "_WON"
		}
		if current == movearr[4] && current == movearr[8] {
			return string(current) + "_WON"
		}
		if current == movearr[3] && current == movearr[6] {
			return string(current) + "_WON"
		}
	}
	current = movearr[2]
	if current != '-' {
		if current == movearr[4] && current == movearr[6] {
			return string(current) + "_WON"
		}
		if current == movearr[5] && current == movearr[8] {
			return string(current) + "_WON"
		}
	}
	current = movearr[6]
	if current != '-' {
		if current == movearr[7] && current == movearr[8] {
			return string(current) + "_WON"
		}
	}
	current = movearr[1]
	if current != '-' {
		if current == movearr[4] && current == movearr[7] {
			return string(current) + "_WON"
		}
	}
	current = movearr[3]
	if current != '-' {
		if current == movearr[4] && current == movearr[5] {
			return string(current) + "_WON"
		}
	}

	return "RUNNING"
}

func pcmove(movearr []rune) (string, []rune) {
	var length int
	for i := 0; i < len(movearr); i++ {
		if movearr[i] == '-' {
			length++
		}
	}
	if length < 1 {
		return "DRAW", movearr
	}
	newmove := rand.Intn(length)
	var counter int
	for i := 0; i < len(movearr); i++ {
		if movearr[i] == '-' {
			if counter == newmove {
				movearr[i] = 'O'
			}
			counter++
		}
	}
	return "RUNNING", movearr
}

func configureAPI(api *operations.TictactoeAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	api.Logger = log.Printf
	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()
	api.JSONProducer = runtime.JSONProducer()

	api.Logger("Started API")

	// PUT NEW MOVE
	api.PutAPIV1GamesGameIDHandler = operations.PutAPIV1GamesGameIDHandlerFunc(func(params operations.PutAPIV1GamesGameIDParams) middleware.Responder {
		api.Logger("Put new move")

		var found bool
		var thegame *models.Game
		var index int
		for i := 0; i < len(games); i++ {
			if games[i].ID == params.GameID {
				thegame = games[i]
				index = i
				found = true
			}
		}

		if found {
			api.Logger(*params.Game.Board)
			if thegame.Status != "RUNNING" {
				return operations.NewPutAPIV1GamesGameIDBadRequest()
			}
			newstate, newboard := move(*thegame.Board, *params.Game.Board)
			if newstate == "ERROR" {
				return operations.NewPutAPIV1GamesGameIDBadRequest()
			}
			games = remove(games, index)
			refreshedgame := models.Game{Board: &newboard, Status: newstate, ID: thegame.ID}
			games = append(games, &refreshedgame)
			return operations.NewPutAPIV1GamesGameIDOK().WithPayload(&refreshedgame)
		} else {
			return operations.NewPutAPIV1GamesGameIDNotFound()
		}
	})

	// LIST GAMES
	api.GetAPIV1GamesHandler = operations.GetAPIV1GamesHandlerFunc(func(params operations.GetAPIV1GamesParams) middleware.Responder {
		api.Logger("List games")
		return operations.NewGetAPIV1GamesOK().WithPayload(games)
	})

	// START GAME
	api.PostAPIV1GamesHandler = operations.PostAPIV1GamesHandlerFunc(func(params operations.PostAPIV1GamesParams) middleware.Responder {
		api.Logger("New game created")

		var gamestate string
		rand.Seed(time.Now().UnixNano())
		if rand.Intn(2) == 0 {
			startspot := rand.Intn(9)
			startarr := []string{"-", "-", "-", "-", "-", "-", "-", "-", "-"}
			startarr[startspot] = "O"
			gamestate = strings.Join(startarr, "")
		} else {
			gamestate = "---------"
		}

		gamestatus := "RUNNING"
		id := strfmt.UUID(uuid.New().String())

		game := models.Game{Board: &gamestate, Status: gamestatus, ID: id}
		games = append(games, &game)
		return operations.NewPostAPIV1GamesCreated().WithLocation(game.ID.String())
	})

	// DELETE GAME
	api.DeleteAPIV1GamesGameIDHandler = operations.DeleteAPIV1GamesGameIDHandlerFunc(func(params operations.DeleteAPIV1GamesGameIDParams) middleware.Responder {
		api.Logger("Delete game")

		var found bool
		for i := 0; i < len(games); i++ {
			if games[i].ID == params.GameID {
				games = remove(games, i)
				found = true
				break
			}
		}

		if found {
			return operations.NewDeleteAPIV1GamesGameIDOK()
		} else {
			return operations.NewDeleteAPIV1GamesGameIDNotFound()
		}
	})

	// GET GAME STATE
	api.GetAPIV1GamesGameIDHandler = operations.GetAPIV1GamesGameIDHandlerFunc(func(params operations.GetAPIV1GamesGameIDParams) middleware.Responder {
		api.Logger("Get game state")

		var thegame *models.Game
		var found bool
		for i := 0; i < len(games); i++ {
			if games[i].ID == params.GameID {
				found = true
				thegame = games[i]
			}
		}

		if found {
			return operations.NewGetAPIV1GamesGameIDOK().WithPayload(thegame)
		} else {
			return operations.NewGetAPIV1GamesGameIDNotFound()
		}
	})

	if api.DeleteAPIV1GamesGameIDHandler == nil {
		api.DeleteAPIV1GamesGameIDHandler = operations.DeleteAPIV1GamesGameIDHandlerFunc(func(params operations.DeleteAPIV1GamesGameIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.DeleteAPIV1GamesGameID has not yet been implemented")
		})
	}
	if api.GetAPIV1GamesHandler == nil {
		api.GetAPIV1GamesHandler = operations.GetAPIV1GamesHandlerFunc(func(params operations.GetAPIV1GamesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIV1Games has not yet been implemented")
		})
	}
	if api.GetAPIV1GamesGameIDHandler == nil {
		api.GetAPIV1GamesGameIDHandler = operations.GetAPIV1GamesGameIDHandlerFunc(func(params operations.GetAPIV1GamesGameIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIV1GamesGameID has not yet been implemented")
		})
	}
	if api.PostAPIV1GamesHandler == nil {
		api.PostAPIV1GamesHandler = operations.PostAPIV1GamesHandlerFunc(func(params operations.PostAPIV1GamesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostAPIV1Games has not yet been implemented")
		})
	}
	if api.PutAPIV1GamesGameIDHandler == nil {
		api.PutAPIV1GamesGameIDHandler = operations.PutAPIV1GamesGameIDHandlerFunc(func(params operations.PutAPIV1GamesGameIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PutAPIV1GamesGameID has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
