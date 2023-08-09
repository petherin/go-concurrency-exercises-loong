//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"fmt"
	"time"
)

func producer(stream Stream) chan *Tweet {
	tweets := make(chan *Tweet)

	go func() {
		for {
			tweet, err := stream.Next() // 320ms delay x 5 = 1600ms
			if err == ErrEOF {
				close(tweets)
				return
			}

			tweets <- tweet
		}
	}()

	return tweets
}

func consumer(tweets chan *Tweet) {
	for tweet := range tweets {
		if tweet.IsTalkingAboutGo() { //330 ms delay x 5 = 1650ms
			fmt.Println(tweet.Username, "\ttweets about golang")
		} else {
			fmt.Println(tweet.Username, "\tdoes not tweet about golang")
		}
	}
}

func main() {
	start := time.Now()
	stream := GetMockStream()

	// Producer
	tweets := producer(stream)

	// Consumer
	consumer(tweets)

	fmt.Printf("Process took %s\n", time.Since(start)) // Without concurrency, total time at least 1600 + 1650 = 3250ms
}
