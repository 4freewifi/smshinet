package main

import (
	"errors"
	"github.com/golang/glog"
	"time"
)

var NotInitialized error = errors.New("Not initialized")

type ResourcePool struct {
	pool   []interface{}
	tokens chan int
	size   int
}

func (t *ResourcePool) Initialize(r []interface{}) error {
	if t.tokens != nil {
		return errors.New("Cannot re-initialize")
	}
	t.pool = r
	size := len(r)
	t.tokens = make(chan int, size)
	for i := 0; i < size; i++ {
		t.tokens <- i
	}
	t.size = size
	glog.V(1).Infof("ResourcePool initialized, size %d", size)
	return nil
}

func (t *ResourcePool) Get() (int, interface{}) {
	id, ok := <-t.tokens
	if !ok {
		panic(NotInitialized)
	}
	t.size--
	glog.V(1).Infof("ResourcePool get token %d, left %d", id, t.size)
	return id, t.pool[id]
}

func (t *ResourcePool) GetWithTimeout(timeout time.Duration) (
	int, interface{}, error) {
	select {
	case <-time.After(timeout):
		glog.V(1).Info("ResourcePool get timeout")
		return 0, nil, errors.New("Timeout")
	case id, ok := <-t.tokens:
		if !ok {
			panic(NotInitialized)
		}
		t.size--
		glog.V(1).Infof("ResourcePool get token %d, left %d", id, t.size)
		return id, t.pool[id], nil
	}
}

func (t *ResourcePool) Put(id int, r interface{}) {
	t.pool[id] = r
	t.tokens <- id
	t.size++
	glog.V(1).Infof("ResourcePool put token %d, left %d", id, t.size)
	return
}
