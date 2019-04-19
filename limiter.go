package main

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-redis/redis_rate"
	consul "github.com/hashicorp/consul/api"
)

// FrequencyState .
type FrequencyState int

// ReqFrequency Values
const (
	FrequencyStateNormal  FrequencyState = 0
	FrequencyStateTooHigh FrequencyState = 1
)

const KeyIPRateThreshold = "IP_RATE_THREDSHOLD"

type RedisLimiter interface {
	AllowMinute(name string, maxn int64) (count int64, delay time.Duration, allow bool)
}

type SettingGetter interface {
	Get(key string) ([]byte, error)
}

type ConsulKV struct {
	kv *consul.KV
}

func (ck *ConsulKV) Get(key string) (value []byte, err error) {
	pair, _, err := ck.kv.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return pair.Value, nil
}

type IPLimit struct {
	Threshold     int64
	SettingGetter SettingGetter
	RedisLimiter  RedisLimiter
}

func NewIPLimit(consulClient *consul.Client, redisClient redis.Cmdable) *IPLimit {
	return &IPLimit{
		SettingGetter: &ConsulKV{kv: consulClient.KV()},
		RedisLimiter:  redis_rate.NewLimiter(redisClient),
	}
}

func (il *IPLimit) UpdateThreshold() error {
	value, err := il.SettingGetter.Get(KeyIPRateThreshold)
	if err != nil {
		return err
	}

	threshold, err := strconv.Atoi(string(value))
	if err != nil {
		return err
	}

	il.Threshold = int64(threshold)
	return nil
}

func (il *IPLimit) IncrFrequency(id string) (FrequencyState, error) {
	if il.Threshold == 0 {
		return FrequencyStateNormal, nil
	}

	_, _, allow := il.RedisLimiter.AllowMinute(id, il.Threshold)
	if allow {
		return FrequencyStateNormal, nil
	}

	return FrequencyStateTooHigh, nil
}
