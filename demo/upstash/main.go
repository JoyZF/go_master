package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"sync"
)

var ctx = context.Background()

type Post struct {
	Slug    string
	Content string
}

func GetLastNPostsFromDatabases(n int) []Post {
	var posts []Post
	for i := 0; i < n; i++ {
		posts = append(posts, Post{
			Slug:    fmt.Sprintf("post-%d-slug", i),
			Content: fmt.Sprintf("Some random content for the post #%d", i),
		})
	}
	return posts
}

func FillCacheWithPostsOneByOne(ctx context.Context, rdb *redis.Client, posts []Post) error {
	for _, post := range posts {
		// save each post one by one
		if err := rdb.Set(ctx, fmt.Sprintf("post:%s", post.Slug), post.Content, 0).Err(); err != nil {
			return err
		}
	}
	return nil
}

func FIllCacheWithPostsInBatches(ctx context.Context, rdb *redis.Client, posts []Post) error {
	_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, post := range posts {
			if err := pipe.Set(ctx, fmt.Sprintf("post:%s", post.Slug), post.Content, 0).Err(); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func MakeNewPage(ctx context.Context, rdb *redis.Client, slug string, viewLimit int) error {
	if err := rdb.Set(ctx, fmt.Sprintf("page:%s:views", slug), 0, 0).Err(); err != nil {
		return fmt.Errorf("error while saving page default view: %v", err)
	}
	if err := rdb.Set(ctx, fmt.Sprintf("page:%s:viewLimit", slug), viewLimit, 0).Err(); err != nil {
		fmt.Errorf("error while setting page view limit: %v", err)
	}
	return nil
}

func CheckIfCanVisitPageWithoutTransaction(ctx context.Context, rdb *redis.Client, slug string) (bool, error) {
	limit, err := rdb.Get(ctx, fmt.Sprintf("page:%s:viewLimit", slug)).Int()
	if err != nil {
		return false, fmt.Errorf("error while getting page view limit: %v", err)
	}
	currentViews, err := rdb.Get(ctx, fmt.Sprintf("page:%s:views", slug)).Int()
	if err != nil {
		return false, fmt.Errorf("error while getting page's current views: %v", err)
	}

	// the page has reached its view limit
	if currentViews >= limit {
		return false, nil
	}

	// adding the new view
	if err = rdb.Set(ctx, fmt.Sprintf("page:%s:views", slug), currentViews+1, 0).Err(); err != nil {
		// if an error happens the view has not been added, and we don't show the user the page
		return false, fmt.Errorf("error while saving page default view: %v", err)
	}
	return true, nil
}

func CheckIfCanVisitPageWithTransaction(ctx context.Context, rdb *redis.Client, slug string) (bool, error) {
	viewLimitKey := fmt.Sprintf("page:%s:viewLimit", slug)
	viewsKey := fmt.Sprintf("page:%s:views", slug)

	canView := false
	return canView, rdb.Watch(ctx, func(tx *redis.Tx) error {
		// using tx instead of rdb ensures if those values has changed, the transaction will fail
		limit, err := tx.Get(ctx, viewLimitKey).Int()
		if err != nil {
			return fmt.Errorf("error while getting page view limit: %v", err)
		}

		currentViews, err := tx.Get(ctx, viewsKey).Int()
		if err != nil {
			return fmt.Errorf("error while getting page's current views: %v", err)
		}

		// the page has reached its view limit
		if currentViews >= limit {
			return nil
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			// adding the new view
			if err := pipe.Set(ctx, viewsKey, currentViews+1, 0).Err(); err != nil {
				// if an error happens the view has not been added, and we don't show the user the page
				return fmt.Errorf("error while saving page default view: %v", err)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("error while executing the pipeline: %v", err)
		}
		canView = true
		return nil

	}, viewsKey)
}

func main() {
	opt, _ := redis.ParseURL("redis://default:812bac92b4fc4c7687423b4e8ad1565e@apn1-subtle-parrot-34093.upstash.io:34093")
	rdb := redis.NewClient(opt)

	//startTime := time.Now()
	//posts := GetLastNPostsFromDatabases(100)
	//if err := FIllCacheWithPostsInBatches(ctx, client, posts); err != nil {
	//	panic(err)
	//}
	//fmt.Println("took ", time.Since(startTime))
	if err := MakeNewPage(context.Background(), rdb, "test-1", 10); err != nil {
		panic(err)
	}

	//var wg sync.WaitGroup
	//wg.Add(2)
	//go func() {
	//	can, err := CheckIfCanVisitPageWithoutTransaction(context.Background(), rdb, "test-1")
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println("Can #1", can)
	//	wg.Done()
	//}()
	//go func() {
	//	<-time.After(time.Millisecond * 500) // to ensure the first one will change the second one's value
	//	can, err := CheckIfCanVisitPageWithoutTransaction(context.Background(), rdb, "test-1")
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println("Can #2", can)
	//	wg.Done()
	//}()
	//wg.Wait()

	scriptContent, _ := os.ReadFile("./script.lua")
	script := redis.NewScript(string(scriptContent))

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			can, err := script.Run(context.Background(), rdb, []string{"test-1"}).Result()
			if err != nil {
				panic(err)
			}
			fmt.Printf("can #%d = %v\n", i, can)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
