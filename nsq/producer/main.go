package main

import (
	"github.com/nsqio/go-nsq"
	"log"
)

func main() {

	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("101.43.17.86:4150", config)
	if err != nil {
		log.Fatal(err)
	}
	messageBody := []byte("hello")
	topicName := "topic"

	// Synchronously publish a single message to the specified topic.
	// Messages can also be sent asynchronously and/or in batches.
	for i := 0; i < 100000; i++ {
		go func() {
			err = producer.Publish(topicName, messageBody)
			if err != nil {

			}
		}()
	}
	for {

	}
	// Gracefully stop the producer when appropriate (e.g. before shutting down the service)
	producer.Stop()
}
