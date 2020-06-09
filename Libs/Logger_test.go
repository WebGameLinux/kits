package Libs

import (
		"fmt"
		"github.com/hashicorp/go-uuid"
		"testing"
		"time"
)

func TestNewLoggerBird(t *testing.T) {
		var log = NewLoggerBird()
		var channels = []chan interface{}{
				make(chan interface{}, 2),
				make(chan interface{}, 2),
		}
		var ch = map[string][]chan interface{}{
				LogTypeError: channels,
		}
		log.AppendChannel(ch)

		for {
				select {
				case <-time.NewTicker(3 * time.Second).C:
						log.Error(time.Now().String() + ":::log")
				case <-time.NewTicker(10 * time.Second).C:
						go func(arr []chan interface{}) {
								uid, _ := uuid.GenerateUUID()
								for _, ch := range arr {
										fmt.Println("worker:", uid)
										if len(ch) > 0 {
												fmt.Printf("%T,%v\n", ch, <-ch)
										}
								}
						}(channels)
				}
		}

}
