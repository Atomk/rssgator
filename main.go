package main

import (
	"database/sql"
	"fmt"
	"os"
	"rssgator/internal/config"
	"rssgator/internal/database"

	// Postgres driver for the database/sql package
	_ "github.com/lib/pq"
)

type state struct {
	conf *config.Config
	db   *database.Queries
}

type cliCommand struct {
	name string
	args []string
}

type cliCommandHandler func(*state, cliCommand) error

type commands struct {
	nameToHandler map[string]cliCommandHandler
}

func main() {
	// Initialize commands list
	commandsList := commands{
		nameToHandler: map[string]cliCommandHandler{
			"login":     handlerLogin,
			"register":  handlerRegister,
			"reset":     handlerReset,
			"users":     handlerListUsers,
			"agg":       handlerAggregate,
			"addfeed":   middlewareLoggedIn(handlerAddFeed),
			"feeds":     handlerListFeeds,
			"follow":    middlewareLoggedIn(handlerFollow),
			"following": middlewareLoggedIn(handlerListFollowing),
			"unfollow":  middlewareLoggedIn(handlerUnfollow),
		},
	}

	configOptions, err := config.Read()
	if os.IsNotExist(err) {
		// Create a basic config so that user is not forced to create it manually
		fmt.Println("config file not found, creating a new one")
		configOptions = &config.Config{
			DbUrl: "postgres://example",
		}
		configOptions.Write()
	}

	db, err := sql.Open("postgres", configOptions.DbUrl)
	if err != nil {
		fmt.Printf("Error while opening DB connection: %s\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	programState := state{
		conf: configOptions,
		db:   dbQueries,
	}

	// Command execution
	if len(os.Args) < 2 {
		fmt.Println("this program needs at least one argument")
		os.Exit(1)
	}
	cmd := cliCommand{
		name: os.Args[1],
		args: os.Args[2:],
	}
	err = commandsList.run(&programState, cmd)
	if err != nil {
		fmt.Println("error while executing command:", err)
		os.Exit(1)
	}
}

// --------------------
//    Struct methods
// --------------------

// Run a given command.
func (c *commands) run(s *state, cmd cliCommand) error {
	handler, exists := c.nameToHandler[cmd.name]
	if !exists {
		return fmt.Errorf("command `%s` not found", cmd.name)
	}
	return handler(s, cmd)
}

// Register a handler function for a given command name.
func (c *commands) register(name string, f cliCommandHandler) error {
	_, exists := c.nameToHandler[name]
	if exists {
		return fmt.Errorf("command name `%s` is already registered", name)
	}
	c.nameToHandler[name] = f
	return nil
}
