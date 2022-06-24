package internal

// import (
// 	"context"
// 	"sync"
// )

// type Task func(context)

// type workerPool struct {
// 	workers int
// 	tasksQueue chan Task
// }

// func InitWorkerPool(workers int, tasks []Task) workerPool {
// 	queue := make(chan Task, len(tasks))
// 	for i, task := range tasks {
// 		queue <- task
// 	}
// 	return workerPool{
// 		workers: workers,
// 		tasksQueue: queue,
// 	}
// }

// func (w *workerPool) Run(ctx context) error {
// 	var wg *sync.WaitGroup
// 	wg.Add(w.workers)

// 	// each worker will only exit (mark wg.Done()) when it would see that the tasksQueue is empty
// 	for i:=0; i<w.workers; i++ {
// 		go func(c context) {
// 			defer wg.Done()
// 			for {
// 				select {
// 				case task := <-w.tasksQueue:
// 					task()
// 				default:
// 					break
// 				}
// 			}
// 		}()
// 	}
// 	wg.Wait()
// }
