package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"task-cli/repository"
	"task-cli/service"
	"task-cli/util"
	"task-cli/view"
)

func main() {
	// Get the command arguments without the command line name
	argumentsWithoutProg := os.Args[1:]

	// Instanciating the Task Service to handle the tasks actions
	taskService := service.NewTaskService(
		// Instanciating the Task Repository to read/write the "task.json" file
		repository.NewTaskRepository("./db/tasks.json"),
	)

	// Validate if the command was called with arguments, otherwise throw an error and exit
	if len(argumentsWithoutProg) == 0 {
		util.LogError(errors.New("you must provide an option"))
	}

	// Get the first argument and check if is valid
	switch argument := argumentsWithoutProg[0]; argument {
	case "add":
		if len(argumentsWithoutProg) == 1 {
			util.LogError(errors.New("description not provided"))
		}
		newTaskID, err := taskService.AddTask(argumentsWithoutProg[1])
		util.LogError(err)

		fmt.Printf("Task added successfully (ID: %d)\n", newTaskID)
	case "update":
		if len(argumentsWithoutProg) == 1 {
			util.LogError(errors.New("task id not provided"))
		}
		taskID, err := strconv.Atoi(argumentsWithoutProg[1])
		util.LogError(err)

		description := argumentsWithoutProg[2]
		if len(description) == 0 {
			util.LogError(errors.New("description not provided"))
		}
		_, err = taskService.UpdateTaskDescription(taskID, description)
		util.LogError(err)

	case "delete":
		if len(argumentsWithoutProg) == 1 {
			util.LogError(errors.New("task id not provided"))
		}
		taskID, err := strconv.Atoi(argumentsWithoutProg[1])
		util.LogError(err)

		taskService.DeleteBy(taskID)
	case "mark-in-progress":
		if len(argumentsWithoutProg) == 1 {
			util.LogError(errors.New("task id not provided"))
		}
		taskID, err := strconv.Atoi(argumentsWithoutProg[1])
		util.LogError(err)

		_, err = taskService.UpdateTaskStatus(taskID, "in-progress")
		util.LogError(err)

	case "mark-done":
		if len(argumentsWithoutProg) == 1 {
			util.LogError(errors.New("task id not provided"))
		}
		taskID, err := strconv.Atoi(argumentsWithoutProg[1])
		util.LogError(err)

		_, err = taskService.UpdateTaskStatus(taskID, "done")
		util.LogError(err)

	case "list":
		if len(argumentsWithoutProg) > 1 {
			tasks, err := taskService.GetByStatus(argumentsWithoutProg[1])
			util.LogError(err)

			view.PromptTableTasks(tasks)
			return
		}
		tasks, err := taskService.GetAll()
		util.LogError(err)

		view.PromptTableTasks(tasks)
		return
	default:
		util.LogError(errors.New("option provided not valid"))
	}
}
