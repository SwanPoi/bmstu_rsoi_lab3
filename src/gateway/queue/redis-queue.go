package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RetryRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

var RedisCtx = context.Background()
var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",            
		DB:       0,      
	})

	_, err := RedisClient.Ping(RedisCtx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connected")
}

func EnqueueRetry(req RetryRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return RedisClient.RPush(RedisCtx, "retry_queue", data).Err()
}

func DoRequest(method, url string, headers map[string]string, body []byte) (int, []byte, error) {
    req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
    if err != nil {
        return 0, nil, err
    }

    for k, v := range headers {
        req.Header.Set(k, v)
    }

    client := &http.Client{
        Timeout: 5 * time.Second,
    }

    resp, err := client.Do(req)
    if err != nil {
        return 0, nil, err
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    return resp.StatusCode, respBody, err
}


func StartRetryWorker() {
	if RedisClient == nil {
		return
	}
	
	go func() {
		for {
			res, err := RedisClient.BLPop(RedisCtx, 5*time.Second, "retry_queue").Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				log.Printf("Redis BLPop error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			var req RetryRequest
			if err := json.Unmarshal([]byte(res[1]), &req); err != nil {
				log.Printf("Failed to unmarshal request: %v", err)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			reqHTTP, _ := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
			for k, v := range req.Headers {
				reqHTTP.Header.Set(k, v)
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(reqHTTP)
			cancel()

			if err != nil || (resp != nil && resp.StatusCode >= 400) {
				log.Printf("Retry failed for %s, re-enqueueing", req.URL)
				EnqueueRetry(req)
			} else {
				log.Printf("Retry succeeded for %s", req.URL)
			}

			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

