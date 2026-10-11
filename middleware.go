package main

import (
	"context"
	"fmt"
	"rssgator/internal/database"
)

// Ensure the current user (as per config) is logged in and pass its object to
// the provided handler.
//
// This is an Higher order function that takes a "login" handler, wraps it
// within a "normal" handler that is then returned.
func middlewareLoggedIn(
	handler func(s *state, cmd cliCommand, user database.User) error,
) cliCommandHandler {
	return func(s *state, cmd cliCommand) error {
		user, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
		if err != nil {
			// TODO do this only if the error is "no rows found"
			return fmt.Errorf("current user is '%s' but it does not exist in database\n", s.conf.CurrentUserName)
		}
		return handler(s, cmd, user)
	}
}
