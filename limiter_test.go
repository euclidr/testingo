package main

import (
	"fmt"
	"testing"

	"github.com/go-redis/redis_rate"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

type MockSettingGetter struct {
	mock.Mock
}

func (m *MockSettingGetter) Get(key string) (value []byte, err error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func TestUpdateThreshold(t *testing.T) {
	tests := []struct {
		v      string
		err    error
		rs     int64
		hasErr bool
	}{
		{v: "1000", err: nil, rs: 1000, hasErr: false},
		{v: "a", err: nil, rs: 0, hasErr: true},
		{v: "", err: fmt.Errorf("consul is down"), rs: 0, hasErr: true},
	}

	for idx, test := range tests {
		mockSettingGetter := new(MockSettingGetter)
		mockSettingGetter.On("Get", mock.Anything).Return([]byte(test.v), test.err)

		limiter := &IPLimit{SettingGetter: mockSettingGetter}
		err := limiter.UpdateThreshold()
		if test.hasErr {
			assert.Error(t, err, "row %d", idx)
		} else {
			assert.NoError(t, err, "row %d", idx)
		}
		assert.Equal(t, test.rs, limiter.Threshold, "thredshold should equal, row %d", idx)
	}
}

func TestIncrFrequency(t *testing.T) {
	tests := []struct {
		id    string
		thres int64
		times int
		state FrequencyState
	}{
		{id: "1.2.3.4", thres: 10, times: 1, state: FrequencyStateNormal},
		// repeat 7 times to make sure it's ok when crossing minutes
		{id: "1.2.3.4", thres: 3, times: 7, state: FrequencyStateTooHigh},
		{id: "1.2.3.4", thres: 0, times: 2, state: FrequencyStateNormal},
	}
	for idx, test := range tests {
		testRdsSrv.FlushAll()
		rClient := redis.NewClient(&redis.Options{
			Addr: testRdsSrv.Addr(),
		})
		limiter := &IPLimit{
			Threshold:    test.thres,
			RedisLimiter: redis_rate.NewLimiter(rClient),
		}
		for i := 0; i < test.times-1; i++ {
			limiter.IncrFrequency(test.id)
		}

		state, _ := limiter.IncrFrequency(test.id)
		assert.Equal(t, test.state, state, "row %d", idx)
	}
	testRdsSrv.FlushAll()
}
