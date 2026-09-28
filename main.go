package main

import (
	"fmt"
	"os"
	"rssgator/internal/config"
)

type state struct {
	conf *config.Config
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

	configOptions := config.Read()
	programState := state{conf: &configOptions}

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
	err := s.conf.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Println("username set successfully")
	return nil
}
