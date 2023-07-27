/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package sse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventLog(t *testing.T) {
	ev := NewEventLog(1*time.Second, 1024)
	testEvent := &Event{Data: []byte("test")}

	ev.Add(testEvent)
	ev.Clear()

	assert.Equal(t, 0, ev.events.Len())

	ev.Add(testEvent)
	ev.Add(testEvent)

	assert.Equal(t, 2, ev.events.Len())
}

func TestEventLogMaxCapacity(t *testing.T) {
	ev := NewEventLog(10*time.Second, 3)
	testEvent := &Event{Data: []byte("test")}

	ev.Add(testEvent)
	ev.Add(testEvent)
	ev.Add(testEvent)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 3, ev.events.Len())

	ev.Add(testEvent)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 3, ev.events.Len())

	ev.Add(testEvent)
	ev.Add(testEvent)
	ev.Add(testEvent)
	ev.Add(testEvent)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 3, ev.events.Len())
}

func BenchmarkEventLogMaxCapacity(b *testing.B) {
	ev := NewEventLog(100*time.Second, 100)
	testEvent := &Event{Data: []byte("test")}

	b.RunParallel(func (pb *testing.PB) {
		for pb.Next() {
			ev.Add(testEvent)
		}
	})
}
