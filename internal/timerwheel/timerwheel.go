/*
A hierarchical timer wheel to add, remove, and fire node active in amortized O(1) time.
The node active are deferred until the timer is advanced, which is performed yourself.
*/
package timerwheel

import (
	"math"
	"math/bits"
	"time"
)

var (
	// buckets number of each wheel
	buckets = []int{64, 64, 32, 4, 1}
	// spans of a bucket in wheel
	spans = []int64{
		ceilingPowerOfTwo64(int64(time.Second)),                      // 1.07s
		ceilingPowerOfTwo64(int64(time.Minute)),                      // 1.14m
		ceilingPowerOfTwo64(int64(time.Hour)),                        // 1.22h
		ceilingPowerOfTwo64(int64(time.Hour * 24)),                   // 1.63d
		int64(buckets[3]) * ceilingPowerOfTwo64(int64(time.Hour*24)), // 6.5d
		int64(buckets[3]) * ceilingPowerOfTwo64(int64(time.Hour*24)), // 6.5d
	}
	// shift of each wheel to get index
	shift = []int{
		bits.TrailingZeros(uint(spans[0])),
		bits.TrailingZeros(uint(spans[1])),
		bits.TrailingZeros(uint(spans[2])),
		bits.TrailingZeros(uint(spans[3])),
		bits.TrailingZeros(uint(spans[4])),
	}
)

// ceilingPowerOfTwo64 returns smallest power of two greater than or equal x
func ceilingPowerOfTwo64(x int64) int64 {
	return 1 << (64 - bits.LeadingZeros64(uint64(x-1)))
}

// TimerWheel
// A timer wheel [1] stores timer events in buckets on a circular buffer. A bucket represents a
// coarse time span, e.g. one minute, and holds a doubly-linked list of events. The wheels are
// structured in a hierarchy (seconds, minutes, hours, days) so that events scheduled in the
// distant future are cascaded to lower buckets when the wheels rotate. This allows for events
// to be added, removed, and expired in O(1) time, where expiration occurs for the entire bucket,
// and penalty of cascading is amortized by the rotations.
//
// [1] Hashed and Hierarchical Timing Wheels
// http://www.cs.columbia.edu/~nahum/w6998/papers/ton97-timing-wheels.pdf
// You need to maintain nodes you put into wheel, you may want to remove it or change active
// time.
// It accept active time is less then nanos but we don't guarantee when it will be executed.
// It means you need to maintain nanos by method Advance relationship with nodes`s active time
// yourself.
type TimerWheel struct {
	wheel [][]Node
	nanos int64
}

func NewTimerWheel() *TimerWheel {
	W := new(TimerWheel)
	wheel := make([][]Node, len(buckets))
	for i := range wheel {
		wheel[i] = make([]Node, buckets[i])
		for j := range wheel[i] {
			wheel[i][j] = newSentinel()
		}
	}
	W.wheel = wheel
	W.nanos = time.Now().UnixNano()
	return W
}

//Schedules a timer event for the node.
func (w *TimerWheel) Schedule(node Node) {
	sentinel := w.findBucket(node.GetVariableTime())
	link(sentinel, node)
}

//Reschedule an active timer event for the node.
func (w *TimerWheel) ReSchedule(n Node) {
	if n.GetPreviousInVariableOrder() != nil {
		unlink(n)
		w.Schedule(n)
	}
}

//DeSchedule a timer event for this entry if present.
func (w *TimerWheel) DeSchedule(n Node) {
	unlink(n)
	n.SetNextInVariableOrder(nil)
	n.SetPreviousInVariableOrder(nil)
}

//Determines the bucket that the timer event should be added to.
func (w *TimerWheel) findBucket(t int64) Node {
	duration := t - w.nanos
	for i := range w.wheel {
		if duration < spans[i+1] {
			ticks := t >> shift[i]
			index := ticks & int64(buckets[i]-1)
			return w.wheel[i][index]
		}
	}
	return w.wheel[len(w.wheel)-1][0]
}

//Adds the entry at the tail of the bucket's list.
func link(sentinel, n Node) {
	n.SetPreviousInVariableOrder(sentinel.GetPreviousInVariableOrder())
	n.SetNextInVariableOrder(sentinel)

	sentinel.GetPreviousInVariableOrder().SetNextInVariableOrder(n)
	sentinel.SetPreviousInVariableOrder(n)
}

//Removes the entry from its bucket, if scheduled.
func unlink(n Node) {
	next := n.GetNextInVariableOrder()
	if next != nil {
		prev := n.GetPreviousInVariableOrder()
		next.SetPreviousInVariableOrder(prev)
		prev.SetNextInVariableOrder(next)
	}
}

//expire entries or reschedules into the proper bucket if still active.
func (w *TimerWheel) expire(index int, previousTicks, currentTicks int64) {
	timerWheel := w.wheel[index]
	var start, end int
	if (currentTicks - previousTicks) >= int64(len(timerWheel)) {
		end = len(timerWheel)
		start = 0
	} else {
		mask := spans[index] - 1
		start = int(previousTicks & mask)
		end = 1 + int(currentTicks&mask)
	}

	mask := len(timerWheel) - 1
	for i := start; i < end; i++ {
		sentinel := timerWheel[i&mask]
		//prev := sentinel.GetPreviousInVariableOrder()
		/*
			todo 只执行一部分的node需要恢复链 如果是要执行全部的node就没问题了
				原java代码在执行部分会有一个try-catch,catch部分会恢复链,再并推出expire
		*/
		node := sentinel.GetNextInVariableOrder()
		sentinel.SetPreviousInVariableOrder(sentinel)
		sentinel.SetNextInVariableOrder(sentinel)

		for node != sentinel {
			next := node.GetNextInVariableOrder()
			node.SetPreviousInVariableOrder(nil)
			node.SetNextInVariableOrder(nil)
			if node.GetVariableTime() > w.nanos {
				//Time doesn't reach then
				//Put it back or put it into smaller span wheel.
				w.Schedule(node)
			} else {
				node.Active()
			}
			node = next
		}
	}
}

//Advance the timer and evicts entries that have expired.
func (w *TimerWheel) Advance(currentTimeNanos int64) {
	previousTimeNanos := w.nanos
	w.nanos = currentTimeNanos
	for i := range shift {
		previousTicks := previousTimeNanos >> shift[i]
		currentTicks := currentTimeNanos >> shift[i]
		if currentTicks-previousTicks <= 0 {
			break
		}
		w.expire(i, previousTimeNanos, currentTimeNanos)
	}

}

//GetExpirationDelay Returns the duration until the next bucket expires, or MaxInt64 if none.
func (w *TimerWheel) GetExpirationDelay() int64 {
	for i := range w.wheel {
		timerWheel := w.wheel[i]
		ticks := w.nanos >> shift[i]

		spanMask := spans[i] - 1
		start := int(ticks & spanMask)
		end := start + len(timerWheel)
		mask := len(timerWheel) - 1
		for j := start; j < end; j++ {
			sentinel := timerWheel[i&mask]
			next := sentinel.GetNextInVariableOrder()
			if next == sentinel {
				continue
			}

			buckets := j - start
			delay := (int64(buckets) << shift[i]) - (w.nanos & spanMask)
			if delay <= 0 {
				delay = spans[i]
			}

			for k := i + 1; k < len(w.wheel); k++ {
				nextDelay := w.peekAhead(k)
				if delay < nextDelay {
					delay = nextDelay
				}
			}
			return delay
		}
	}
	return math.MaxInt64
}

//peekAhead Returns the duration when the wheel's next bucket expires, or MaxInt64 if empty.
func (w *TimerWheel) peekAhead(i int) int64 {
	ticks := w.nanos >> shift[i]
	timerWheel := w.wheel[i]

	spanMask := spans[i] - 1
	mask := len(timerWheel) - 1
	probe := (ticks + 1) & int64(mask)
	sentinel := timerWheel[probe]
	next := sentinel.GetNextInVariableOrder()
	if next == sentinel {
		return math.MaxInt64
	}
	return spans[i] - (w.nanos & spanMask)
}

//Snapshot Returns an unmodifiable snapshot map roughly ordered by the expiration time.
//The wheels are evaluated in order, but the timers that fall within the bucket's range are not sorted.
//Beware that obtaining the mappings is NOT a constant-time operation.
func (w *TimerWheel) Snapshot(ascending bool, limit int) map[interface{}]interface{} {
	if limit <= 0 {
		return nil
	}
	r := make(map[interface{}]interface{}, limit)
	startLevel := boolAB(ascending, 0, len(w.wheel)-1).(int)
	for i, timerWheel := range w.wheel {
		indexOffset := boolAB(ascending, i, -i).(int)
		index := startLevel + indexOffset

		ticks := w.nanos >> shift[index]
		bucketMask := len(timerWheel) - 1
		startBucket := int(ticks&int64(bucketMask)) + boolAB(ascending, 1, 0).(int)
		for j := range timerWheel {
			bucketOffset := boolAB(ascending, j, -j).(int)
			sentinel := timerWheel[(startBucket+bucketOffset)&bucketMask]
			for node := boolAB(ascending, sentinel.GetNextInVariableOrder(), sentinel.GetPreviousInVariableOrder()).(Node);
				node != sentinel;
			node = boolAB(ascending, node.GetNextInVariableOrder(), node.GetPreviousInVariableOrder()).(Node) {
				if len(r) >= limit {
					break
				}
				key := node.GetKey()
				value := node.GetValue()
				if key != nil && value != nil {
					r[key] = value
				}
			}
		}
	}
	return r
}

func boolAB(bool2 bool, a, b interface{}) interface{} {
	if bool2 {
		return a
	}
	return b
}
