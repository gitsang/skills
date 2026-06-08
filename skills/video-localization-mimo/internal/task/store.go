package task

import (
	"fmt"
	"sync"
	"time"
)

type Store interface {
	Create(task *Task) error
	Get(id string) (*Task, error)
	Update(task *Task) error
	List() ([]*Task, error)
	Delete(id string) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks: make(map[string]*Task),
	}
}

func (s *MemoryStore) Create(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task %s already exists", task.ID)
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	task.Status = StatusPending
	task.Progress = 0

	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStore) Get(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return task, nil
}

func (s *MemoryStore) Update(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; !exists {
		return fmt.Errorf("task %s not found", task.ID)
	}

	task.UpdatedAt = time.Now()
	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStore) List() ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return fmt.Errorf("task %s not found", id)
	}

	delete(s.tasks, id)
	return nil
}
