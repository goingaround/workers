package exec

import "sync"

const (
	execDO = iota
	execWAIT
	execABORT
	execDONE
)

type state struct {
	code int
	err  error
}

func (s state) Code() int {
	return s.code
}

func (s state) Err() error {
	return s.err
}

type State struct {
	m sync.RWMutex
	v interface {
		Code() int
		Err() error
	}
}

func NewState(initial interface {
	Code() int
	Err() error
}) State {
	return State{v: initial}
}

func (s *State) Set(v interface {
	Code() int
	Err() error
}) {
	s.m.Lock()
	defer s.m.Unlock()
	s.v = v
}

func (s *State) Code() int {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.v.Code()
}

func (s *State) Err() error {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.v.Err()
}