package service

import (
	"task-cli/datamodel"
	"task-cli/repository"
	"time"
)

type TaskService interface {
	GetAll() ([]datamodel.Task, error)
	GetByStatus(status string) ([]datamodel.Task, error)
	AddTask(description string) (int, error)
	UpdateTaskDescription(taskID int, description string) (datamodel.Task, error)
	UpdateTaskStatus(taskID int, status string) (datamodel.Task, error)
	DeleteBy(taskID int) (bool, error)
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) GetAll() ([]datamodel.Task, error) {
	return s.repo.SelectMany(func(_ datamodel.Task) bool {
		return true
	}, -1)
}

func (s *taskService) GetByStatus(status string) ([]datamodel.Task, error) {
	return s.repo.SelectMany(func(t datamodel.Task) bool {
		return t.Status == status
	}, -1)
}

func (s *taskService) AddTask(description string) (int, error) {
	task, err := s.repo.InsertOrUpdate(0, func(task datamodel.Task) datamodel.Task {
		task.Description = description
		task.Status = "todo"
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
		return task
	})
	return task.ID, err
}

func (s *taskService) UpdateTaskDescription(taskID int, description string) (datamodel.Task, error) {
	return s.repo.InsertOrUpdate(taskID, func(task datamodel.Task) datamodel.Task {
		task.Description = description
		task.UpdatedAt = time.Now()
		return task
	})
}

func (s *taskService) UpdateTaskStatus(taskID int, status string) (datamodel.Task, error) {
	return s.repo.InsertOrUpdate(taskID, func(task datamodel.Task) datamodel.Task {
		task.Status = status
		task.UpdatedAt = time.Now()
		return task
	})
}

func (s *taskService) DeleteBy(taskID int) (bool, error) {
	return s.repo.Delete(taskID)
}
