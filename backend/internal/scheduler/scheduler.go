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

type TaskHeap struct {
	tasks    []models.ScheduledBlogData
	indexMap map[string]int
}

func (h TaskHeap) Len() int { return len(h.tasks) }

func (h TaskHeap) Less(i, j int) bool {
	return h.tasks[i].ScheduledBlog.ScheduledTime.Before(h.tasks[j].ScheduledBlog.ScheduledTime)
}

func (h TaskHeap) Swap(i, j int) {
	h.tasks[i], h.tasks[j] = h.tasks[j], h.tasks[i]
	h.indexMap[h.tasks[i].ScheduledBlog.Blog.Id] = i
	h.indexMap[h.tasks[j].ScheduledBlog.Blog.Id] = j
}

func (h *TaskHeap) Push(x interface{}) {
	task := x.(models.ScheduledBlogData)
	h.tasks = append(h.tasks, task)
	h.indexMap[task.ScheduledBlog.Blog.Id] = len(h.tasks) - 1
}

func (h *TaskHeap) Pop() interface{} {
	n := len(h.tasks)
	task := h.tasks[n-1]
	h.tasks = h.tasks[0 : n-1]
	delete(h.indexMap, task.ScheduledBlog.Blog.Id)
	return task
}

func (h *TaskHeap) RemoveAt(index int) models.ScheduledBlogData {
	n := len(h.tasks)
	h.Swap(index, n-1)
	removed := h.tasks[n-1]
	h.tasks = h.tasks[:n-1]
	delete(h.indexMap, removed.ScheduledBlog.Blog.Id)
	if index < len(h.tasks) {
		heap.Fix(h, index)
	}
	return removed
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
		ctx:       ctx,
		cancel:    cancel,
		newTaskCh: make(chan struct{}, 1),
		heap: &TaskHeap{
			tasks:    []models.ScheduledBlogData{},
			indexMap: make(map[string]int),
		},
	}
	if err := s.loadTasks(); err != nil {
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

		nextTask := s.heap.tasks[0]
		scheduledTimeUTC := nextTask.ScheduledBlog.ScheduledTime.UTC()
		timeUntil := time.Until(scheduledTimeUTC)
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
			continue
		}

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

	if task.EmailId != "" {
		if err := services.SendEmail(task.EmailId, task.Message); err != nil {
			log.Printf("[ERROR] Sending OTP email to %s: %v", task.EmailId, err)
		}
		delErr := repo.DeleteScheduledTask(task)
		if delErr != nil {
			log.Printf("[ERROR] Error deleting scheduled task: %v", delErr)
		}

	} else {
		log.Printf("[INFO] Worker executing task for user %v with blog %v, for platforms %v", task.UserID, task.ScheduledBlog.Blog.Id, task.ScheduledBlog.Platforms)

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
}

func (s *Scheduler) loadTasks() error {
	tasks, err := repo.GetScheduledTasks()
	if err != nil {
		log.Printf("[ERROR] Error loading tasks: %v", err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.heap = &TaskHeap{
		tasks:    tasks,
		indexMap: make(map[string]int),
	}
	for i, task := range s.heap.tasks {
		s.heap.indexMap[task.ScheduledBlog.Blog.Id] = i
	}
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
	case s.newTaskCh <- struct{}{}:
	default:
	}
	return nil
}

func (s *Scheduler) RemoveTask(blogId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, ok := s.heap.indexMap[blogId]
	if !ok {
		return nil
	}
	task := s.heap.RemoveAt(index)
	err := repo.DeleteScheduledTask(task)
	if err != nil {
		log.Printf("[ERROR] Error deleting task: %v", err)
		return err
	}
	select {
	case s.newTaskCh <- struct{}{}: // so we have to notify the agent to recheck the heap
	default:
	}
	return nil
}

func (s *Scheduler) Stop() {
	s.cancel()
}
