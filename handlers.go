package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"rssgator/internal/database"
	"rssgator/internal/rss"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd cliCommand) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("the `%s` command expects one argument (username), %d found\n", cmd.name, len(cmd.args))
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
		return fmt.Errorf("the `%s` command expects one argument (username), %d found\n", cmd.name, len(cmd.args))
	}

	username := cmd.args[0]
	ctx := context.Background()

	_, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
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
		return fmt.Errorf("the `%s` command expects no arguments, %d found\n", cmd.name, len(cmd.args))
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
		return fmt.Errorf("the `%s` command expects no arguments, %d found\n", cmd.name, len(cmd.args))
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

// Fetch the least recently fetched feed, and print its items.
func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("could not get next feed to fetch: %v\n", err)
		return
	}

	fmt.Printf("fetching feed '%s'\n", feed.Name)

	ctx := context.Background()
	rssFeed, err := rss.FetchFeed(&ctx, feed.Url)
	if err != nil {
		fmt.Printf("ERROR: could not fetch '%s', skipping: %v\n", feed.Url, err)
		return
	}

	// fmt.Printf("found %d RSS items\n", len(rssFeed.Channel.Items))

	err = s.db.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
		ID:        feed.ID,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		fmt.Printf("ERROR: could not update last fetch time: %v\n", err)
		return
	}

	if len(rssFeed.Channel.Items) == 0 {
		fmt.Println("no RSS channel items, skipping")
		return
	}

	for _, item := range rssFeed.Channel.Items {
		// Mon, 12 Oct 2026 20:35:47 +0000
		parsedPubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			fmt.Printf("ERROR: Could not parse publication date %q\n", item.PubDate)
		}
		pubDate := sql.NullTime{
			Time:  parsedPubDate.UTC(),
			Valid: err == nil,
		}
		_, err = s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			Title:       item.Title,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		})
		if err != nil {
			// It's ok if the post cannot be inserted due to duplicate URL, it
			// means it's already in the DB. All other errors should be logged.
			duplicateUrlError := strings.HasSuffix(err.Error(), "constraint \"posts_url_key\" (23505)")
			if !duplicateUrlError {
				fmt.Printf("ERROR: %s\n", err)
			}
			continue
		}

		fmt.Printf("Added post '%s' (published %s)\n", item.Title, pubDate.Time.Format(time.DateTime))
	}
}

// Start a loop that fetches the feeds at regular intervals.
func handlerAggregate(s *state, cmd cliCommand) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("the `%s` command expects 1 argument, %d found\n", cmd.name, len(cmd.args))
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return errors.New("the provided argument is not a valid duration, try '30s'")
	} else if timeBetweenRequests.Seconds() < 5 {
		return errors.New("time between requests must be at least 5 seconds")
	}

	ctx := context.Background()
	feeds, err := s.db.GetAllFeedsWithUsername(ctx)
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Println("no stored feeds")
		return nil
	}

	fmt.Printf("Collecting feeds every %v\n(CTRL+C to stop)\n\n", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd cliCommand, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("the `%s` command expects 2 arguments, %d found\n", cmd.name, len(cmd.args))
	}

	name := strings.TrimSpace(cmd.args[0])
	url := strings.TrimSpace(cmd.args[1])
	if name == "" || url == "" {
		return errors.New("feed name and url cannot be empty or made of whitespace")
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

	_, err = insertFeedFollows(s, user.ID, feed.ID)
	if err != nil {
		return nil
	}

	fmt.Println(feed)
	return nil
}

func handlerListFeeds(s *state, cmd cliCommand) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("the `%s` command expects no arguments, %d found\n", cmd.name, len(cmd.args))
	}

	feeds, err := s.db.GetAllFeedsWithUsername(context.Background())
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Println("no stored feeds")
		return nil
	}

	fmt.Printf("Stored feeds: %d\n", len(feeds))
	for _, feed := range feeds {
		fmt.Printf("- %s (%s) | user: %s\n", feed.Name, feed.Url, feed.Username)
	}
	return nil
}

// Add a new row in the feed_follows table, which means that a user is now
// following a feed.
func insertFeedFollows(s *state, userID, feedID uuid.UUID) (database.CreateFeedFollowsRow, error) {
	feedFollow, err := s.db.CreateFeedFollows(context.Background(), database.CreateFeedFollowsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    userID,
		FeedID:    feedID,
	})
	if err != nil {
		return database.CreateFeedFollowsRow{}, fmt.Errorf("could not follow feed: %w", err)
	}
	return feedFollow, nil
}

func handlerFollow(s *state, cmd cliCommand, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("the `%s` command expects 1 argument, %d found\n", cmd.name, len(cmd.args))
	}

	url := strings.TrimSpace(cmd.args[0])
	if url == "" {
		return errors.New("url cannot be empty or made of whitespace")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not get feed with url '%s': %w", url, err)
	}

	feedFollow, err := insertFeedFollows(s, user.ID, feed.ID)
	if err != nil {
		return nil
	}

	// feedFollow.UserName should be equal to s.conf.CurrentUserName
	fmt.Printf("User '%s' now follows '%s'\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func handlerUnfollow(s *state, cmd cliCommand, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("the `%s` command expects 1 argument, %d found\n", cmd.name, len(cmd.args))
	}

	url := strings.TrimSpace(cmd.args[0])
	if url == "" {
		return errors.New("url cannot be empty or made of whitespace")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not get feed with url '%s': %w", url, err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("could not unfollow '%s': %w", url, err)
	}

	fmt.Printf("User '%s' does not follow '%s' anymore.\n", user.Name, url)
	return nil
}

func handlerListFollowing(s *state, cmd cliCommand, user database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("the `%s` command expects no arguments, %d found\n", cmd.name, len(cmd.args))
	}

	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	if len(followedFeeds) == 0 {
		fmt.Println("no followed feeds")
		return nil
	}

	fmt.Printf("Feeds followed: %d\n", len(followedFeeds))
	for _, followedFeed := range followedFeeds {
		fmt.Printf("- %s | user: %s\n", followedFeed.FeedName, followedFeed.UserName)
	}
	return nil
}

func handlerListPosts(s *state, cmd cliCommand, user database.User) error {
	if len(cmd.args) > 1 {
		return fmt.Errorf("the `%s` command expects 0 or 1 arguments %d found\n", cmd.name, len(cmd.args))
	}

	var limit int32
	if len(cmd.args) == 0 {
		limit = 2
	} else {
		limit, err := strconv.Atoi(cmd.args[0])
		if err != nil || limit < 1 {
			return fmt.Errorf("you must provide a positive, non-zero number")
		}
	}

	ctx := context.Background()
	posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	if len(posts) == 0 {
		fmt.Println("no posts found for current user")
		return nil
	}

	var pubTime string
	for _, post := range posts {
		if post.PublishedAt.Valid {
			pubTime = post.PublishedAt.Time.Format(time.DateOnly)
		} else {
			pubTime = "XXXX-XX-XX"
		}
		fmt.Printf("- %s | %s\n", pubTime, post.Title)
	}
	return nil
}
