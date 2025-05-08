package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pliliya111/go_final_sprint/internal/agent"
	"github.com/pliliya111/go_final_sprint/internal/calculator"
)

func getComputingPower() int {
	powerStr := os.Getenv("COMPUTING_POWER")
	if powerStr == "" {
		return 1
	}

	power, err := strconv.Atoi(powerStr)
	if err != nil {
		log.Printf("Invalid COMPUTING_POWER value: %v. Using default value: 1", err)
		return 1
	}

	if power < 1 {
		log.Printf("COMPUTING_POWER must be at least 1. Using default value: 1")
		return 1
	}

	return power
}

func worker(id int) {
	for {
		task, err := agent.FetchTask()
		if err != nil {
			log.Printf("Worker %d: Error fetching task: %v", id, err)
			time.Sleep(1 * time.Second)
			continue
		}

		if task == nil {
			log.Printf("Worker %d: No tasks available, waiting...", id)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("Worker %d: Processing task %s: %v %s %v", id, task.ID, task.Arg1, task.Operation, task.Arg2)

		result := calculator.PerformOperation(task)
		log.Printf("Worker %d: Task %s result: %v", id, task.ID, result)

		if err := agent.SubmitTaskResult(task.ID, result); err != nil {
			log.Printf("Worker %d: Error submitting result for task %s: %v", id, task.ID, err)
		}

		time.Sleep(3 * time.Second)
	}
}

func main() {
	computingPower := getComputingPower()
	log.Printf("Starting agent with %d workers", computingPower)

	for i := 0; i < computingPower; i++ {
		go worker(i)
	}

	select {}
}
