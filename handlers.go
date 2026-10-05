package main

import (
	"context"
	"errors"
	"fmt"
	"rssgator/internal/database"
	"rssgator/internal/rss"
	"strings"
	"time"

	"github.com/google/uuid"
)

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

func handlerAggregate(s *state, cmd cliCommand) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("the `agg` command expects zero arguments, %d found\n", len(cmd.args))
	}

	const url = "https://www.wagslane.dev/index.xml"
	ctx := context.Background()
	rssFeed, err := rss.FetchFeed(&ctx, url)
	if err != nil {
		return err
	}

	if len(rssFeed.Channel.Items) == 0 {
		fmt.Println("no RSS channel items")
		return nil
	}

	fmt.Printf("found %d RSS items\n", len(rssFeed.Channel.Items))

	fmt.Println(rssFeed)
	return nil
}

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
