package main

import (
	"fmt"
	"sync"
)

// Spawn 启动一个 actor：内部 goroutine 是唯一消费者，返回值是它的 mailbox
func Spawn(handle func(msg any)) chan<- any {
	mailbox := make(chan any, 64)
	go func() {
		for msg := range mailbox { // 串行：一次只处理一条
			handle(msg)
		}
	}()
	return mailbox
}

func main() {
	count := 0 // 私有状态：只有 actor 的 goroutine 会碰它

	addr := Spawn(func(msg any) {
		switch m := msg.(type) {
		case int:
			count += m // 无锁累加，跑一千年也不会少一次
		case chan int:
			m <- count // 查询：附上回信地址，结果以消息的形式寄回
		}
	})

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr <- 1 // 任意 goroutine 都能投递：多生产者
		}()
	}
	wg.Wait()

	reply := make(chan int, 1)
	addr <- reply
	fmt.Println(<-reply) // 永远输出 1000
}
