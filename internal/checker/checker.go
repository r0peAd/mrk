package checker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/conncheck/internal/config"
)

type Kind string

const (
	KindMySQL      Kind = "mysql"
	KindPostgreSQL Kind = "postgresql"
	KindRedis      Kind = "redis"
	KindMongoDB    Kind = "mongodb"
	KindKafka      Kind = "kafka"
)

type Result struct {
	Kind      string        `json:"kind"`
	Name      string        `json:"name"`
	Target    string        `json:"target"`
	Success   bool          `json:"success"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	StartedAt time.Time     `json:"started_at"`
}

func safeDisplayUser(username string) string {
	if username == "" {
		return "anonymous"
	}
	return username
}

func Run(ctx context.Context, cfg config.Config, concurrency int) []Result {
	tasks := make([]struct {
		kind   Kind
		target config.Target
	}, 0)

	appendTasks := func(kind Kind, targets []config.Target) {
		for _, target := range targets {
			tasks = append(tasks, struct {
				kind   Kind
				target config.Target
			}{kind, target})
		}
	}
	appendTasks(KindMySQL, cfg.MySQL)
	appendTasks(KindPostgreSQL, cfg.PostgreSQL)
	appendTasks(KindRedis, cfg.Redis)
	appendTasks(KindMongoDB, cfg.MongoDB)
	appendTasks(KindKafka, cfg.Kafka)

	if concurrency <= 0 {
		concurrency = len(tasks)
	}
	if concurrency > len(tasks) {
		concurrency = len(tasks)
	}

	results := make([]Result, len(tasks))
	if len(tasks) == 0 {
		return results
	}

	queue := make(chan int)
	var workers sync.WaitGroup
	var writer sync.Mutex

	for worker := 0; worker < concurrency; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range queue {
				result := Check(ctx, tasks[index].kind, tasks[index].target)
				writer.Lock()
				results[index] = result
				writer.Unlock()
			}
		}()
	}

	for index := range tasks {
		select {
		case queue <- index:
		case <-ctx.Done():
		}
	}
	close(queue)
	workers.Wait()
	return results
}

func Check(ctx context.Context, kind Kind, target config.Target) Result {
	started := time.Now()
	result := Result{
		Kind:      string(kind),
		Name:      target.Name,
		StartedAt: started,
	}

	checkCtx, cancel := context.WithTimeout(ctx, target.Timeout)
	defer cancel()

	var err error
	switch kind {
	case KindMySQL:
		result.Target = mysqlTarget(target)
		err = checkMySQL(checkCtx, target)
	case KindPostgreSQL:
		result.Target = postgresTarget(target)
		err = checkPostgreSQL(checkCtx, target)
	case KindRedis:
		result.Target = redisTarget(target)
		err = checkRedis(checkCtx, target)
	case KindMongoDB:
		result.Target = mongoTarget(target)
		err = checkMongoDB(checkCtx, target)
	case KindKafka:
		result.Target = kafkaTarget(target)
		err = checkKafka(checkCtx, target)
	default:
		err = fmt.Errorf("unsupported check kind: %s", kind)
	}

	result.Latency = time.Since(started)
	result.Success = err == nil
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
