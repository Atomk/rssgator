package main

import (
	"context"
	"errors"
	"fmt"
	"rssgator/internal/database"
	"strings"
	"time"

	"github.com/google/uuid"
)

func handlerAddFeed(s *state, cmd cliCommand) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("the `addFeed` command expects 2 arguments, %d found\n", len(cmd.args))
	}

	name := strings.TrimSpace(cmd.args[0])
	url := strings.TrimSpace(cmd.args[1])
	if name == "" || url == "" {
		return errors.New("feed name and url cannot be empty or made of whitespace")
	}

	user, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
	if err != nil {
		// TODO do this only if the error is "no rows found"
		return fmt.Errorf("current user is '%s' but it does not exist in database\n", s.conf.CurrentUserName)
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("could not create feed: %w", err)
	}

	fmt.Println(feed)
	return nil
}
