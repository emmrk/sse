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
	ev := NewEventLog(1*time.Second)
	testEvent := &Event{Data: []byte("test")}

	ev.Add(testEvent)
	ev.Clear()

	assert.Equal(t, 0, ev.events.Len())

	ev.Add(testEvent)
	ev.Add(testEvent)

	assert.Equal(t, 2, ev.events.Len())
}
