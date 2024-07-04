package queue

import (
	"auto-update/internal/sshclient"
	"fmt"
)

type UpdateQueue struct {
	updateChannel  chan *sshclient.UpdateOptions
	workingChannel chan bool
}

var sshClientService = sshclient.NewSshClientService()

// NewupdateQueue is a function to create new email queue
func NewUpdateQueue() *UpdateQueue {
	updateChannel := make(chan *sshclient.UpdateOptions, 100)
	workingChannel := make(chan bool, 100)
	return &UpdateQueue{
		updateChannel:  updateChannel,
		workingChannel: workingChannel,
	}
}

// Logical flow from the queue
func (e *UpdateQueue) Work() {

	fmt.Println("Starting queue worker")
	for {
		select {
		case id := <-e.updateChannel:
			fmt.Println("Starting queue worker updating repository")
			// Enqueue message to workingChannel to avoid miscalculation in queue size.
			e.workingChannel <- true

			fmt.Println("testeeeeeee", id)
			sshClientService.UpdateRepository(id)
			fmt.Println("Finish queue worker updating repository")

			<-e.workingChannel
		}
	}
}

// Size is a function to get the size of email queue
func (e *UpdateQueue) Size() int {
	return len(e.updateChannel) + len(e.workingChannel)
}

// Enqueue is a function to enqueue email string into email channel
func (e *UpdateQueue) Enqueue(options *sshclient.UpdateOptions) {
	fmt.Println("Enqueue:", options)
	e.updateChannel <- options
}
