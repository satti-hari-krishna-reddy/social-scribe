package scheduler

import (
	"container/heap"
	"context"
	"log"
	"social-scribe/backend/internal/models"
	repo "social-scribe/backend/internal/repositories"
	"time"
    "sync"
)

type TaskHeap []models.ScheduledBlogData

func (h TaskHeap) Len() int { return len(h) }

func (h TaskHeap) Less(i, j int) bool {
	return h[i].ScheduledBlog.ScheduledTime.Before(h[j].ScheduledBlog.ScheduledTime)
}

func (h TaskHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *TaskHeap) Push(x interface{}) {
	*h = append(*h, x.(models.ScheduledBlogData))
}

func (h *TaskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type Scheduler struct {
	heap      *TaskHeap
	mu        sync.Mutex        
	ctx       context.Context
	cancel    context.CancelFunc
	newTaskCh chan struct{}      
}


func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		heap:      &TaskHeap{},
		ctx:       ctx,
		cancel:    cancel,
		newTaskCh: make(chan struct{}, 1),
	}
	err := s.loadTasks()
	if err != nil {
		log.Printf("[ERROR] Error loading tasks, Stopping the Scheduler: %v", err)
		cancel()
	}
	go s.runAgent()
	return s
}

func (s *Scheduler) runAgent() {
	var timer *time.Timer 

	for {
		// Check the if heap is empty and if so, wait for new tasks
		s.mu.Lock()
		if s.heap.Len() == 0 {
			log.Println("[INFO] No tasks in the heap, waiting for new tasks")
			s.mu.Unlock()
			select {
			case <-s.newTaskCh: 
				continue
			case <-s.ctx.Done(): 
				return
			}
		}

		nextTask := (*s.heap)[0]
		timeUntil := time.Until(nextTask.ScheduledBlog.ScheduledTime)
		s.mu.Unlock()

		if timer == nil {
			timer = time.NewTimer(timeUntil)
		} else {
			log.Printf("[INFO] Resetting timer from %v to %v", timer, timeUntil)
			timer.Reset(timeUntil)
		}

		select {
		case <-timer.C:
			s.mu.Lock()
			heap.Pop(s.heap) 
			s.mu.Unlock()
			go s.worker(nextTask) 

		case <-s.newTaskCh: 
			if !timer.Stop() {
				<-timer.C 
			}

		case <-s.ctx.Done(): 
			if !timer.Stop() {
				<-timer.C
			}
			return
		}
	}
}

func (s *Scheduler) worker(task models.ScheduledBlogData) {

	// mock function for now
	log.Printf("Executing task: %v", task)
	work := heap.Pop(s.heap).(models.ScheduledBlogData)
	log.Printf(" the time of the scheduled blog is: %v", work.ScheduledBlog.ScheduledTime)
	log.Printf("Task executed: %v", task.ScheduledBlog.ScheduledTime)
}

func (s *Scheduler) loadTasks() error {
	tasks, err := repo.GetScheduledTasks()
	if err != nil {
		log.Printf("[ERROR] Error loading tasks: %v", err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.heap = (*TaskHeap)(&tasks)
	heap.Init(s.heap)
	log.Printf("[INFO] Loaded %d tasks successfully into heap", len(tasks))
	return nil
}

func (s *Scheduler) AddTask(task models.ScheduledBlogData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := repo.StoreScheduledTask(task)
	if err != nil {
		return err
	}
	heap.Push(s.heap, task)

	select {
	// Send a signal to the agent informing about the newly added task
	case s.newTaskCh <- struct{}{}:
	default: 
	}
	return nil
}

// func (s * Scheduler) DeleteTask( task models.ScheduledBlogData) error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
	
// 	err := repo.DeleteScheduledTask(task)
// 	if err != nil {
// 		return err
// 	}
// 	heap.Pop(s.heap)
// 	return nil
// }