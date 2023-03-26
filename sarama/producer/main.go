package main

import (
	"github.com/Shopify/sarama"
	"log"
	"time"
)

func main() {
	Producer("auto_create_test", 1000)
}
func Producer(topic string, limit int) {
	config := sarama.NewConfig()
	// 异步生产者不建议把 Errors 和 Successes 都开启，一般开启 Errors 就行
	// 同步生产者就必须都开启，因为会同步返回发送成功或者失败
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Version = sarama.V2_8_0_0
	asyncProducer, err := sarama.NewAsyncProducer([]string{"101.43.31.63:9092"}, config)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case suc := <-asyncProducer.Successes():
				log.Printf("success %+v", suc)
			case err := <-asyncProducer.Errors():
				log.Printf("error %+v", err)
			}
		}
	}()
	for i := 0; i < limit; i++ {
		asyncProducer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(time.Now().Format(time.RFC3339Nano)),
			Value: sarama.StringEncoder(time.Now().Format(time.RFC3339Nano)),
		}
	}
	time.Sleep(10 * time.Second)
}
