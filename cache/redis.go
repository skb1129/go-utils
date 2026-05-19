package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
	"github.com/skb1129/go-utils/request"
	"go.uber.org/zap"
)

var logger *zap.Logger

type Cache struct {
	r *redis.Client
}

func NewCache() *Cache {
	logger = logs.GetLogger()

	client := redis.NewClient(&redis.Options{
		Addr:     config.GetString("redis.address"),
		Password: config.GetString("redis.password"),
		DB:       config.GetInt("redis.db"),
		PoolSize: config.GetInt("redis.poolSize"),
	})

	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	logger.Info("Connected to Redis", zap.String("PING", pong))

	return &Cache{r: client}
}

func (cache *Cache) Close() error {
	return cache.r.Close()
}

func (cache *Cache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	logger.Debug("SETTING IN REDIS", zap.String("key", key), zap.Any("value", value))
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	result := cache.r.Set(ctx, key, bytes, expiration)
	return result.Err()
}

func (cache *Cache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	logger.Debug("SETTING IN REDIS", zap.String("key", key), zap.String("value", value))
	result := cache.r.Set(ctx, key, value, expiration)
	return result.Err()
}

func (cache *Cache) Incr(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	logger.Debug("INCREMENTING IN REDIS", zap.String("key", key))
	count, err := cache.r.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 && expiration != 0 {
		cache.r.Expire(ctx, key, expiration)
	}
	return count, nil
}

func (cache *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return cache.r.TTL(ctx, key).Result()
}

func (cache *Cache) Get(ctx context.Context, key string) (string, error) {
	logger.Debug("GETTING FROM REDIS", zap.String("key", key))
	result, err := cache.r.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) || result == "" {
		return "", fmt.Errorf(string(request.KeyNotFoundError))
	} else if err != nil {
		return "", err
	}
	return result, nil
}

func (cache *Cache) GetJSON(ctx context.Context, key string, value interface{}) error {
	logger.Debug("GETTING FROM REDIS", zap.String("key", key))
	result := cache.r.Get(ctx, key)
	storedBytes, err := result.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(storedBytes, &value)
}

func (cache *Cache) Delete(ctx context.Context, key string) error {
	logger.Debug("DELETING FROM REDIS", zap.String("key", key))
	return cache.r.Del(ctx, key).Err()
}

func (cache *Cache) DeleteWithPattern(ctx context.Context, pattern string) error {
	logger.Debug("DELETING KEYS MATCHING PATTERN", zap.String("pattern", pattern))
	result, err := cache.r.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	return cache.r.Del(ctx, result...).Err()
}

func (cache *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	logger.Debug("GETTING HASH FROM REDIS", zap.String("key", key), zap.String("field", field))
	result, err := cache.r.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) || result == "" {
		return "", fmt.Errorf(string(request.KeyNotFoundError))
	} else if err != nil {
		return "", err
	}
	return result, nil
}

func (cache *Cache) HSet(ctx context.Context, key, field string, value interface{}) error {
	logger.Debug("SETTING HASH IN REDIS", zap.String("key", key), zap.String("field", field), zap.Any("value", value))
	result := cache.r.HSet(ctx, key, field, value)
	return result.Err()
}

func (cache *Cache) HDel(ctx context.Context, key, field string) error {
	logger.Debug("DELETING HASH FROM REDIS", zap.String("key", key), zap.String("field", field))
	return cache.r.HDel(ctx, key, field).Err()
}

func (cache *Cache) HGetAll(ctx context.Context, key string) (*map[string]string, error) {
	logger.Debug("GETTING HASH FROM REDIS", zap.String("key", key))
	result, err := cache.r.HGetAll(ctx, key).Result()
	if errors.Is(err, redis.Nil) || len(result) == 0 {
		return nil, fmt.Errorf(string(request.KeyNotFoundError))
	} else if err != nil {
		return nil, err
	}
	return &result, nil
}

func (cache *Cache) HKeys(ctx context.Context, key string) (*[]string, error) {
	logger.Debug("GETTING HASH KEYS FROM REDIS", zap.String("key", key))
	result, err := cache.r.HKeys(ctx, key).Result()
	if errors.Is(err, redis.Nil) || len(result) == 0 {
		return nil, fmt.Errorf(string(request.KeyNotFoundError))
	} else if err != nil {
		return nil, err
	}
	return &result, nil
}

func (cache *Cache) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	logger.Debug("RPUSH IN REDIS", zap.String("key", key), zap.Any("values", values))
	return cache.r.RPush(ctx, key, values...).Result()
}

func (cache *Cache) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	logger.Debug("LPUSH IN REDIS", zap.String("key", key), zap.Any("values", values))
	return cache.r.LPush(ctx, key, values...).Result()
}

func (cache *Cache) RPop(ctx context.Context, key string) (string, error) {
	logger.Debug("RPOP FROM REDIS", zap.String("key", key))
	result, err := cache.r.RPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf(string(request.KeyNotFoundError))
	}
	return result, err
}

func (cache *Cache) LPop(ctx context.Context, key string) (string, error) {
	logger.Debug("LPOP FROM REDIS", zap.String("key", key))
	result, err := cache.r.LPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf(string(request.KeyNotFoundError))
	}
	return result, err
}

func (cache *Cache) LRange(ctx context.Context, key string, start, stop int64) (*[]string, error) {
	logger.Debug("LRANGE FROM REDIS", zap.String("key", key), zap.Int64("start", start), zap.Int64("stop", stop))
	result, err := cache.r.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf(string(request.KeyNotFoundError))
	}
	return &result, nil
}

func (cache *Cache) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	logger.Debug("SADD IN REDIS", zap.String("key", key), zap.Any("members", members))
	return cache.r.SAdd(ctx, key, members...).Result()
}

func (cache *Cache) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	logger.Debug("SREM IN REDIS", zap.String("key", key), zap.Any("members", members))
	return cache.r.SRem(ctx, key, members...).Result()
}

func (cache *Cache) SMembers(ctx context.Context, key string) (*[]string, error) {
	logger.Debug("SMEMBERS FROM REDIS", zap.String("key", key))
	result, err := cache.r.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf(string(request.KeyNotFoundError))
	}
	return &result, nil
}

func (cache *Cache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	logger.Debug("SISMEMBER IN REDIS", zap.String("key", key), zap.Any("member", member))
	return cache.r.SIsMember(ctx, key, member).Result()
}

func (cache *Cache) SCard(ctx context.Context, key string) (int64, error) {
	logger.Debug("SCARD FROM REDIS", zap.String("key", key))
	return cache.r.SCard(ctx, key).Result()
}

func (cache *Cache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	logger.Debug("EXPIRE IN REDIS", zap.String("key", key), zap.Duration("expiration", expiration))
	return cache.r.Expire(ctx, key, expiration).Result()
}

func (cache *Cache) Exists(ctx context.Context, key string) (bool, error) {
	logger.Debug("EXISTS IN REDIS", zap.String("key", key))
	result, err := cache.r.Exists(ctx, key).Result()
	return result > 0, err
}

func (cache *Cache) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	logger.Debug("LOCK IN REDIS", zap.String("key", key), zap.Duration("expiration", expiration))
	return cache.r.SetNX(ctx, key, "locked", expiration).Result()
}
