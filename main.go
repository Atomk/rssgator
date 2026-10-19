package main

import (
	"database/sql"
	"fmt"
	"os"
	"rssgator/internal/config"
	"rssgator/internal/database"
	"strings"

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

type cliCommandInfo struct {
	description string
	argsNames   []string
	handler     cliCommandHandler
}

type commands struct {
	nameToCmdInfo map[string]cliCommandInfo
}

func main() {
	// Initialize commands list
	commandsList := commands{
		nameToCmdInfo: map[string]cliCommandInfo{
			"login": {
				description: "Login and set as the current user. Use the `register` command to register first.",
				argsNames:   []string{"username"},
				handler:     handlerLogin,
			},
			"register": {
				description: "Register a new user, and set it as the current user.",
				argsNames:   []string{"username"},
				handler:     handlerRegister,
			},
			"reset": {
				description: "Resets the user table, deleting all registered users.",
				handler:     handlerReset,
			},
			"users": {
				description: "Show a list of all registered users.",
				handler:     handlerListUsers,
			},
			"agg": {
				description: "Start a loop that periodically aggregates feeds, storing all new posts. This is not asynchronous, meaning it does not run in the background.",
				argsNames:   []string{"timeBetweenRequests (e.g. 30s, 2m)"},
				handler:     handlerAggregate,
			},
			"addfeed": {
				description: "Add a new feed, and start following it.",
				argsNames:   []string{"name", "url"},
				handler:     middlewareLoggedIn(handlerAddFeed),
			},
			"feeds": {
				description: "Show a list of each stored feed, along with the user that added it.",
				handler:     handlerListFeeds,
			},
			"follow": {
				description: "Start following the feed with the given URL.",
				argsNames:   []string{"url"},
				handler:     middlewareLoggedIn(handlerFollow),
			},
			"following": {
				description: "List all feeds followed by the current user.",
				handler:     middlewareLoggedIn(handlerListFollowing),
			},
			"unfollow": {
				description: "Unfollow the feed with the given URL.",
				argsNames:   []string{"url"},
				handler:     middlewareLoggedIn(handlerUnfollow),
			},
			"browse": {
				description: "List all posts found in feeds followed by the current user.",
				argsNames:   []string{"limit (optional, positive integer)"},
				handler:     middlewareLoggedIn(handlerListPosts),
			},
		},
	}

	configOptions, err := config.Read()
	if os.IsNotExist(err) {
		// Create a basic config so that user is not forced to create it manually
		fmt.Println("config file not found, creating a new one")
		configOptions = &config.Config{
			DbUrl: "postgres://example",
		}
		err = configOptions.Write()
		if err != nil {
			fmt.Printf("could not write the config file: %v\n", err)
			os.Exit(1)
		}
		// File was written successfully, but the config file still contains
		// example data. Notify the user and stop here for now.
		fmt.Println("see the README for information on how to set up the config file")
		os.Exit(0)
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
		fmt.Println("This program needs at least one argument.")
		commandsList.printHelp()
		os.Exit(1)
	}
	if os.Args[1] == "help" || os.Args[1] == "-h" {
		commandsList.printHelp()
		os.Exit(0)
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
	cmdInfo, exists := c.nameToCmdInfo[cmd.name]
	if !exists {
		return fmt.Errorf("command `%s` not found", cmd.name)
	}
	return cmdInfo.handler(s, cmd)
}

// Register a handler function for a given command name.
func (c *commands) register(name string, f cliCommandInfo) error {
	_, exists := c.nameToCmdInfo[name]
	if exists {
		return fmt.Errorf("command name `%s` is already registered", name)
	}
	c.nameToCmdInfo[name] = f
	return nil
}

func (c *commands) printHelp() {
	fmt.Println("Available commands:")
	for name, cmdInfo := range c.nameToCmdInfo {
		fmt.Printf("- %s: %s\n", name, cmdInfo.description)
		if len(cmdInfo.argsNames) > 0 {
			fmt.Printf("  Arguments: %v\n", strings.Join(cmdInfo.argsNames, ", "))
		}
	}
}
