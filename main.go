package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"rssgator/internal/config"
	"rssgator/internal/database"
	"time"

	// Postgres driver for the database/sql package
	"github.com/google/uuid"
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
		nameToHandler: make(map[string]cliCommandHandler),
	}
	err := commandsList.register("login", handlerLogin)
	if err != nil {
		panic(err)
	}
	err = commandsList.register("register", handlerRegister)
	if err != nil {
		panic(err)
	}
	err = commandsList.register("reset", handlerReset)
	if err != nil {
		panic(err)
	}
	err = commandsList.register("users", handlerListUsers)
	if err != nil {
		panic(err)
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

// --------------------
//      Handlers
// --------------------

func handlerLogin(s *state, cmd cliCommand) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("the `login` command expects one argument (username)")
	}

	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		// TODO do this only if the error is "no rows found"
		return fmt.Errorf("user '%s' is not registered\n", username)
	}

	if s.conf.CurrentUserName == username {
		fmt.Printf("user '%s' is already the current user\n", username)
		return nil
	}

	err = s.conf.SetUser(username)
	if err != nil {
		return fmt.Errorf("cannot set user '%s': %w\n", username, err)
	}
	return nil
}

func handlerRegister(s *state, cmd cliCommand) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("the `register` command expects one argument (username)")
	}

	username := cmd.args[0]
	ctx := context.Background()

	_, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return err
	}

	fmt.Printf("user '%s' registered successfully\n", username)
	err = s.conf.SetUser(username)
	if err != nil {
		return fmt.Errorf("cannot set user '%s': %w\n", username, err)
	}
	return nil
}

func handlerReset(s *state, cmd cliCommand) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("the `reset` command expects zero arguments, %d found\n", len(cmd.args))
	}

	ctx := context.Background()
	err := s.db.DeleteAllUsers(ctx)
	if err != nil {
		return err
	}

	fmt.Println("users table cleared successfully")
	return nil
}

func handlerListUsers(s *state, cmd cliCommand) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("the `users` command expects zero arguments, %d found\n", len(cmd.args))
	}

	ctx := context.Background()
	users, err := s.db.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("no registered users")
		return nil
	}

	fmt.Printf("found %d registered users\n", len(users))
	for _, user := range users {
		if user.Name == s.conf.CurrentUserName {
			fmt.Printf("- %s (current)\n", user.Name)
		} else {
			fmt.Printf("- %s\n", user.Name)
		}
	}
	return nil
}
