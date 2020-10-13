package timerwheel

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type showExecuteNode struct {
	t         int64
	p, n      Node
	activated bool
}

func newsnode(t int64) *showExecuteNode {
	s := new(showExecuteNode)
	s.t = t
	s.p = s
	s.n = s
	return s
}

func (s *showExecuteNode) GetVariableTime() int64 {
	return s.t
}

func (s *showExecuteNode) SetPreviousInVariableOrder(node Node) {
	s.p = node
}

func (s *showExecuteNode) SetNextInVariableOrder(node Node) {
	s.n = node
}

func (s *showExecuteNode) GetPreviousInVariableOrder() Node {
	return s.p
}

func (s *showExecuteNode) GetNextInVariableOrder() Node {
	return s.n
}

func (s *showExecuteNode) Active() {
	//fmt.Println("Execute: ", time.Now(), time.Unix(0, s.t))
	s.activated = true
}

func (s *showExecuteNode) GetKey() interface{} {
	return s.t
}

func (s *showExecuteNode) GetValue() interface{} {
	return s.activated
}

func TestTimerWheel_Schedule(t *testing.T) {
	NOW := time.Now()
	w := NewTimerWheel()
	tests := []struct {
		name   string
		args   Node
		wanted bool
	}{
		{"three second", newsnode(NOW.Add(time.Second * 3).UnixNano()), true},
		{"ten seconds", newsnode(NOW.Add(time.Second * 10).UnixNano()), true},
		{"three minutes", newsnode(NOW.Add(time.Minute * 3).UnixNano()), true},
		{"ten minutes", newsnode(NOW.Add(time.Minute * 10).UnixNano()), true},
		{"ten minutes", newsnode(NOW.Add(time.Minute).UnixNano()), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w.Schedule(tt.args)
			w.Advance(tt.args.GetVariableTime())
			if tt.wanted != tt.args.GetValue().(bool) {
				t.Fatal("not execute: ", tt.name)
			}
		})
	}
}

type FuzzyTestSet struct {
	name string
	args struct {
		clock        int64
		advanceClock int64
		times        []Node
	}
}

func FuzzyTestSetProvider() []FuzzyTestSet {
	sets := make([]FuzzyTestSet, 10)
	for i := range sets {
		clock := rand.Int63()
		bound := 1 + spans[len(spans)-1]
		nodes := make([]Node, 1000)
		adClock := clock + 1 + rand.Int63n(bound)
		for j := range nodes {
			nodes[j] = newsnode(rand.Int63n(bound) + 1 + clock)
		}
		sets[i] = FuzzyTestSet{
			name: strconv.FormatInt(clock, 10),
			args: struct {
				clock        int64
				advanceClock int64
				times        []Node
			}{
				clock:        clock,
				advanceClock: adClock,
				times:        nodes,
			},
		}
	}
	return sets
}

func TestTimerWheel_FuzzySchedule(t *testing.T) {
	for _, tt := range FuzzyTestSetProvider() {
		t.Run(tt.name, func(t *testing.T) {
			w := NewTimerWheel()
			w.nanos = tt.args.clock
			for _, node := range tt.args.times {
				w.Schedule(node)
			}
			w.Advance(tt.args.advanceClock)
			for _, node := range tt.args.times {
				if node.GetVariableTime() <= tt.args.advanceClock && !node.GetValue().(bool) {
					t.Log(node.GetVariableTime(), tt.args.clock, tt.args.advanceClock, node.GetValue())
				}
			}
		})
	}
}

func TestTimerWheel_ReSchedule(t *testing.T) {
	w := NewTimerWheel()
	NOW := time.Now().UnixNano()
	w.nanos = NOW
	node := newsnode(NOW + int64(time.Minute*15))
	w.Schedule(node)
	sentinel1 := node.GetNextInVariableOrder()
	node.t = NOW + int64(time.Hour*2)
	w.ReSchedule(node)
	if sentinel1 == node.GetNextInVariableOrder() {
		t.Fatal("bucket should change")
	}
}

func TestTimerWheel_DeSchedule(t *testing.T) {
	w := NewTimerWheel()
	NOW := time.Now().UnixNano()
	w.nanos = NOW
	node := newsnode(NOW + int64(time.Minute*15))
	w.Schedule(node)
	sentinel1 := node.GetPreviousInVariableOrder()
	w.DeSchedule(node)
	if node.GetNextInVariableOrder() != nil || node.GetPreviousInVariableOrder() != nil ||
		sentinel1.GetPreviousInVariableOrder() != sentinel1 || sentinel1.GetNextInVariableOrder() != sentinel1 {
		t.Fatal("node should be single")
	}
}

func TestTimerWheel_DeScheduleNotScheduled(t *testing.T) {
	w := NewTimerWheel()
	w.DeSchedule(newsnode(1))
}
