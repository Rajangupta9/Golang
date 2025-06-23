package repository

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"task-cli/datamodel"
)

// Query represents the visitor and action queries.
type Query func(datamodel.Task) bool

// Assignment represents the write operations.
type Assignment func(datamodel.Task) datamodel.Task

type TaskRepository interface {
	Exec(query Query, action Query, limit int, mode int) (bool, error)

	Select(query Query) (datamodel.Task, bool, error)
	SelectMany(query Query, limit int) ([]datamodel.Task, error)

	InsertOrUpdate(id int, action Assignment) (datamodel.Task, error)
	Delete(id int) (bool, error)
}

func NewTaskRepository(sourceFile string) TaskRepository {
	return &taskFileRepository{sourceFile: sourceFile}
}

type taskFileRepository struct {
	sourceFile string
	mu         sync.RWMutex
}

const (
	// ReadOnlyMode will RLock(read) the data .
	ReadOnlyMode = iota
	// ReadWriteMode will Lock(read/write) the data.
	ReadWriteMode
)

func (r *taskFileRepository) validateSourceFile() (bool, error) {
	if _, err := os.Stat("db"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("db", os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return false, err
		}

		databaseFile, err := os.Create(r.sourceFile)
		if err != nil {
			return false, err
		}
		databaseFile.Close()

		return true, nil
	}

	if _, err := os.Stat(r.sourceFile); errors.Is(err, os.ErrNotExist) {
		databaseFile, err := os.Create(r.sourceFile)
		if err != nil {
			return false, err
		}
		databaseFile.Close()
	}
	return true, nil
}

func (r *taskFileRepository) readFile() (datamodel.Wrapper[datamodel.Task], error) {
	r.validateSourceFile()

	content, err := os.ReadFile(r.sourceFile)
	if err != nil {
		return datamodel.Wrapper[datamodel.Task]{}, err
	}
	if len(content) == 0 {
		return datamodel.Wrapper[datamodel.Task]{
			CurrentIncrement: 0,
			Records:          []datamodel.Task{},
		}, nil
	}
	wrapper := datamodel.Wrapper[datamodel.Task]{}
	err = json.Unmarshal(content, &wrapper)
	if err != nil {
		return datamodel.Wrapper[datamodel.Task]{}, err
	}
	return wrapper, nil
}

func (r *taskFileRepository) writeFile(wrapper datamodel.Wrapper[datamodel.Task]) error {
	_, err := r.validateSourceFile()
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	err = os.WriteFile(r.sourceFile, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (r *taskFileRepository) Exec(query Query, action Query, actionLimit int, mode int) (bool, error) {
	loops := 0

	if mode == ReadOnlyMode {
		r.mu.RLock()
		defer r.mu.RUnlock()
	} else {
		r.mu.Lock()
		defer r.mu.Unlock()
	}

	wrapper, err := r.readFile()
	if err != nil {
		return false, err
	}
	ok := false
	for _, task := range wrapper.Records {
		ok = query(task)
		if ok {
			if action(task) {
				loops++
				if actionLimit >= loops {
					break // break
				}
			}
		}
	}

	return ok, nil
}

func (r *taskFileRepository) Select(query Query) (datamodel.Task, bool, error) {
	task := datamodel.Task{}
	found, err := r.Exec(query, func(t datamodel.Task) bool {
		task = t
		return true
	}, 1, ReadOnlyMode)

	if err != nil {
		return datamodel.Task{}, found, err
	}

	// set an empty datamodel.Task if not found at all.
	if !found {
		task = datamodel.Task{}
	}

	return task, found, nil
}

func (r *taskFileRepository) SelectMany(query Query, limit int) ([]datamodel.Task, error) {
	results := []datamodel.Task{}
	_, err := r.Exec(query, func(m datamodel.Task) bool {
		results = append(results, m)
		return true
	}, limit, ReadOnlyMode)
	if err != nil {
		return []datamodel.Task{}, err
	}

	return results, nil
}

func (r *taskFileRepository) InsertOrUpdate(id int, action Assignment) (datamodel.Task, error) {
	task := datamodel.Task{}

	if id == 0 { // Create new action
		var currentIncrement int
		// find the biggest ID in order to not have duplications
		// in productions apps you can use a third-party
		// library to generate a UUID as string.
		r.mu.RLock()
		wrapper, err := r.readFile()
		if err != nil {
			return datamodel.Task{}, err
		}

		currentIncrement = wrapper.CurrentIncrement
		r.mu.RUnlock()

		id = currentIncrement + 1
		task.ID = id
		task = action(task)

		// map-specific thing
		r.mu.Lock()

		wrapper.CurrentIncrement = task.ID
		wrapper.Records = append(wrapper.Records, task)

		err = r.writeFile(wrapper)
		if err != nil {
			return datamodel.Task{}, err
		}

		r.mu.Unlock()

		return task, nil
	}

	current, exists, err := r.Select(func(t datamodel.Task) bool {
		return t.ID == id
	})
	if err != nil {
		return datamodel.Task{}, err
	}

	if !exists { // ID is not a real one, return an error.
		return datamodel.Task{}, errors.New("failed to update a nonexistent task")
	}

	// map-specific thing
	r.mu.RLock()
	wrapper, err := r.readFile()
	if err != nil {
		return datamodel.Task{}, err
	}
	r.mu.RUnlock()

	r.mu.Lock()
	for i := 0; i < len(wrapper.Records); i++ {
		if wrapper.Records[i].ID == id {
			wrapper.Records[i] = action(current)
		}
	}

	err = r.writeFile(wrapper)
	if err != nil {
		return datamodel.Task{}, err
	}
	r.mu.Unlock()

	return task, nil
}

func (r *taskFileRepository) Delete(id int) (bool, error) {
	deleted := false

	r.mu.RLock()
	wrapper, err := r.readFile()
	if err != nil {
		return false, err
	}
	r.mu.RUnlock()

	r.mu.Lock()
	for i := 0; i < len(wrapper.Records); i++ {
		if wrapper.Records[i].ID == id {
			wrapper.Records = append(wrapper.Records[:i], wrapper.Records[i+1:]...)
			deleted = true
			break
		}
	}

	err = r.writeFile(wrapper)
	if err != nil {
		return false, err
	}
	r.mu.Unlock()

	return deleted, nil
}
