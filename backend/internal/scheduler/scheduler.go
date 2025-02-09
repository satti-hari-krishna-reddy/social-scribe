package scheduler

import (
	"container/heap"
	"context"
	"log"
	"social-scribe/backend/internal/models"
	repo "social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/services"
	"sync"
	"time"
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
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] Agent panicked: %v", r)
			s.Stop()
		}
	}()

	log.Println("[INFO] Scheduler agent started")

	var timer *time.Timer

	for {
		s.mu.Lock()
		if s.heap.Len() == 0 {
			log.Println("[INFO] No tasks in the heap, waiting for new tasks")
			s.mu.Unlock()

			select {
			case <-s.newTaskCh:
				log.Println("[INFO] New task added, rechecking heap")
				continue
			case <-s.ctx.Done():
				log.Println("[INFO] Scheduler stopped")
				return
			}
		}

		// Peek at the next task
		nextTask := (*s.heap)[0]
		scheduledTimeUTC := nextTask.ScheduledBlog.ScheduledTime.UTC()
		timeUntil := time.Until(scheduledTimeUTC)

		// If the task is overdue or due immediately, force execution
		if timeUntil <= 0 {
			timeUntil = 1 * time.Millisecond
		}

		s.mu.Unlock()

		if timeUntil == 1*time.Millisecond {
			s.mu.Lock()
			if s.heap.Len() > 0 {
				task := heap.Pop(s.heap).(models.ScheduledBlogData)
				s.mu.Unlock()
				go s.worker(task)
			} else {
				s.mu.Unlock()
			}
			// continue to check for more tasks that are due immediately
			continue
		}

		// Otherwise, set up or reset the timer for future tasks.
		if timer == nil {
			timer = time.NewTimer(timeUntil)
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			log.Printf("[INFO] Resetting timer to %v", timeUntil)
			timer.Reset(timeUntil)
		}

		select {
		case <-timer.C:
			s.mu.Lock()
			if s.heap.Len() > 0 {
				task := heap.Pop(s.heap).(models.ScheduledBlogData)
				s.mu.Unlock()
				go s.worker(task)
			} else {
				s.mu.Unlock()
			}

		case <-s.newTaskCh:
			// If a new task is added, recheck the heap
			continue

		case <-s.ctx.Done():
			if timer != nil {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
			return
		}
	}
}

func (s *Scheduler) worker(task models.ScheduledBlogData) {

	log.Printf("[INFO] Worker executing task for user %v with blog %v, for platforms %v", 
	task.UserID, task.ScheduledBlog.Blog.Id, task.ScheduledBlog.Platforms)

	user, err := repo.GetUserById(task.UserID)
	if err != nil || user == nil {
		log.Printf("[ERROR] Error getting user or user not found: %v", task.UserID)
		if delErr := repo.DeleteScheduledTask(task); delErr != nil {
			log.Printf("[ERROR] Error deleting scheduled task: %v", delErr)
		}
		return
	}

	blogId := task.ScheduledBlog.Blog.Id
	platforms := task.ScheduledBlog.Platforms

	processErr := services.ProcessSharedBlog(user, blogId, platforms)
	if processErr != nil {
		log.Printf("[ERROR] Error processing shared blog for blog id %s and user id %s: %v", blogId, task.UserID, processErr)
	}

	delErr := repo.DeleteScheduledTask(task)
	if delErr != nil {
		log.Printf("[ERROR] Error deleting scheduled task: %v", delErr)
	}

	removed := false
	for i, blog := range user.ScheduledBlogs {
		if blog.Id == blogId {
			user.ScheduledBlogs = append(user.ScheduledBlogs[:i], user.ScheduledBlogs[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		log.Printf("[WARN] Blog with id %s not found in user's scheduled blogs", blogId)
	}

	updErr := repo.UpdateUser(task.UserID, user)
	if updErr != nil {
		log.Printf("[ERROR] Error updating user: %v", updErr)
	}

	if processErr != nil {
		log.Printf("[INFO] Task executed with errors for blog with ID %s and user ID %s, error: %v", blogId, task.UserID, processErr)
	} else {
		log.Printf("[INFO] Task executed successfully for blog with ID %s and user ID %s at %v", blogId, task.UserID, task.ScheduledBlog.ScheduledTime)
	}
}

func (s *Scheduler) loadTasks() error {
	tasks, err := repo.GetScheduledTasks()
	if err != nil {
		log.Printf("[ERROR] Error loading tasks: %v", err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.heap = &TaskHeap{}
	*s.heap = append(*s.heap, tasks...)
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

func (s *Scheduler) Stop() {
	s.cancel()
}
