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
		time.AfterFunc(10*time.Second,func() {
				uid, _ := uuid.GenerateUUID()
				for _, ch := range channels {
						fmt.Println("worker:", uid)
						if len(ch) > 0 {
								fmt.Printf("%T,%v\n", ch, <-ch)
						}
				}
		})
		for {
				select {
				case <-time.NewTicker(2 * time.Second).C:
						log.Error("日志测试" + ":::log")
				}
		}

}
