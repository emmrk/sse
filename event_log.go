/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package sse

import (
	"container/list"
	"strconv"
	"sync"
	"time"
)

// EventLog holds unexpired previous events
type EventLog struct {
	sync.RWMutex
	events   *list.List
	ticker   time.Ticker
	eventTTL time.Duration
	maxCap   int
}

// NewEventLog creates a new Event Log.
//
// EventTTL determines for how long the event is considered valid. Valid events
// will be replayed for newly joined clients. Expired events are periodically
// removed from the log to save space if EventTTL != 0. To preserve backwards
// compatibility, with EventTTL == 0 all events ever published on a given
// stream are forever retained and replayed, bevare of the balooning memory as
// a result.
//
// MaxCapacity is a soft limit on the number of entries in the Event Log.
// Filling the Event Log up to MaxCapacity will trigger an unscheduled removal
// of expired entries; if that is not enough, the oldest entry in the Event Log
// will be deleted to free up the space. MaxCapacity == 0 means unlimited
// capacity.
func NewEventLog(eventTTL time.Duration, maxCapacity int) *EventLog {
	e := &EventLog{
		events:   list.New(),
		eventTTL: eventTTL,
		maxCap:   maxCapacity,
	}

	if eventTTL > 0 {
		ticker := time.NewTicker(3 * eventTTL)

		go func() {
			<-ticker.C
			e.CleanUp()
		}()
	}

	return e
}

// Add event to EventLog
func (e *EventLog) Add(ev *Event) {
	if !ev.hasContent() {
		return
	}

	ev.timestamp = time.Now()

	e.Lock()
	defer e.Unlock()

	ev.ID = []byte(e.currentindex())
	e.events.PushBack(ev)

	if e.maxCap != 0 && e.events.Len() >= e.maxCap {
		go func() {
			e.Lock()

			// clean up expired events
			e.cleanUpUnlocked()
			// this is still not enough, remove the oldest element
			if e.events.Len() > e.maxCap {
				e.events.Remove(e.events.Front())
			}

			e.Unlock()
		}()
	}
}

// Clear removes all events from the Event Log
func (e *EventLog) Clear() {
	e.Lock()
	e.events = list.New()
	e.Unlock()
}

// CleanUp removes expired events immediately
func (e *EventLog) CleanUp() {
	e.Lock()
	e.cleanUpUnlocked()
	e.Unlock()
}

func (e *EventLog) cleanUpUnlocked() {
	for element := e.events.Front(); element != nil; element = element.Next() {
		event, ok := element.Value.(*Event)
		if !ok {
			continue
		}

		if time.Now().Sub(event.timestamp) > e.eventTTL {
			e.events.Remove(element)
		} else {
			break
		}
	}
}

// Replay plays unexpired previous events to a subscriber
func (e *EventLog) Replay(s *Subscriber) {
	e.RLock()
	defer e.RUnlock()

	for element := e.events.Front(); element != nil; element = element.Next() {
		event, ok := element.Value.(*Event)
		if !ok {
			continue
		}

		id, _ := strconv.Atoi(string(event.ID))
		if id >= s.eventid {
			if e.eventTTL == 0 {
				s.connection <- event
			} else if time.Now().Sub(event.timestamp) <= e.eventTTL {
				s.connection <- event
			}
		}
	}
}

func (e *EventLog) currentindex() string {
	return strconv.Itoa(e.events.Len())
}
