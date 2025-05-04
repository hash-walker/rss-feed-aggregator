package main

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/hash-walker/rss-feed-aggregator/internal/database"
	"log"
	"strings"
	"sync"
	"time"
)

func startScraper(
	db *database.Queries,
	concurrency int,
	timeBetweenRequests time.Duration,
) {
	log.Printf("Scraping on %v goroutines every %s duration\n", concurrency, timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)

		if err != nil {
			log.Println("err fetching feeds: ", err)
			continue
		}

		wg := &sync.WaitGroup{}

		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(db, wg, feed)
		}

		wg.Wait()

	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)

	if err != nil {
		log.Println("err marking feed : ", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)

	if err != nil {
		log.Println("err fetching feed : ", err)
		return
	}

	for _, item := range rssFeed.Channel.Item {

		description := sql.NullString{}

		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)

		if err != nil {
			log.Printf("couldn't parse date %s with err %v", item.PubDate, err)
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: description,
			PublishedAt: pubDate,
			Url:         item.Link,
			FeedID:      feed.ID,
		})

		if err != nil {

			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}

			log.Printf("couldn't add the post")
		}
	}

	log.Printf("Feed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))

}
