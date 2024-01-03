package send_message_service

import (
	"message-nest/pkg/constant"
	"message-nest/pkg/logging"
	"sync"
)

var maxBufferSize = 10
var Buffer = make(chan SendMessageService, maxBufferSize)

func DoSendTask(task SendMessageService, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logging.Logger.Error("DoSendTask: Recovered from panic:", r)
		}
	}()

	constant.MaxSemaphore <- struct{}{}
	defer func() {
		<-constant.MaxSemaphore
	}()

	go task.Send()
}

func MessageConsumer(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		task, ok := <-Buffer
		if !ok {
			logging.Logger.Error("MessageConsumer: Channel closed. Exiting.")
			return
		}

		wg.Add(1)
		go DoSendTask(task, wg)
	}

}
