package main

import (
	"fmt"
	"time"
)

const (
	date = "2006-01-02"
	datetime = "2006-01-02 15:04:05"
)

func main()  {
	now := time.Now()
	format := now.Format(date)
	fmt.Println(format)
	s := now.Format(datetime)
	fmt.Println(s)
	s2 := now.Format(time.UnixDate)
	fmt.Println(s2)

	h, _ := time.ParseDuration("10h")
	add := now.Add(h)
	fmt.Println(add.Format(datetime))
	sub := now.Sub(time.Now())
	fmt.Println(sub)

	year2000 := time.Date(2023, 12, 01, 0, 0, 0, 0, time.UTC)
	fmt.Println(now.After(year2000))

	time.After(time.Duration(10) * time.Second)
	fmt.Println(123)
	t := time.Now()
	time.Sleep(1 * time.Second)
	fmt.Println(time.Since(t).String())
	t2 := time.Now()
	fmt.Println(t.Equal(t2))

	ticker := time.NewTimer(5 * time.Second)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case t := <-ticker.C:
			fmt.Println("Current time: ", t)
		}
	}
}
