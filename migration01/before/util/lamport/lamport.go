/*

	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this file,
	You can obtain one at http://mozilla.org/MPL/2.0/.

	Copyright (c) 2013, Armon Dadgar armon.dadgar@gmail.com
	Copyright (c) 2013, Mitchell Hashimoto mitchell.hashimoto@gmail.com

	Alternatively, the contents of this file may be used under the terms
	of the GNU General Public License Version 3 or later, as described below:

	This file is free software: you may copy, redistribute and/or modify
	it under the terms of the GNU General Public License as published by the
	Free Software Foundation, either version 3 of the License, or (at your
	option) any later version.

	This file is distributed in the hope that it will be useful, but
	WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
	Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program. If not, see http://www.gnu.org/licenses/.

*/

package lamport

import (
	"sync/atomic"
)

// Clock is a thread safe implementation of a lamport clock. It
// uses efficient atomic operations for all of its functions, falling back
// to a heavy lock only if there are enough CAS failures.
type Clock struct {
	counter uint64
}

// Time is the value of a Clock.
type Time uint64

// NewClock create a new clock with the value 1.
// Value 0 is considered as invalid.
func NewClock() Clock {
	return Clock{
		counter: 1,
	}
}

// NewClockWithTime create a new clock with a value.
func NewClockWithTime(time uint64) Clock {
	return Clock{
		counter: time,
	}
}

// Time is used to return the current value of the lamport clock
func (l *Clock) Time() Time {
	return Time(atomic.LoadUint64(&l.counter))
}

// Increment is used to return the value of the lamport clock and increment it afterwards
func (l *Clock) Increment() Time {
	return Time(atomic.AddUint64(&l.counter, 1) - 1)
}

// Witness is called to update our local clock if necessary after
// witnessing a clock value received from another process
func (l *Clock) Witness(v Time) {
WITNESS:
	// If the other value is old, we do not need to do anything
	cur := atomic.LoadUint64(&l.counter)
	other := uint64(v)
	if other < cur {
		return
	}

	// Ensure that our local clock is at least one ahead.
	if !atomic.CompareAndSwapUint64(&l.counter, cur, other+1) {
		// CAS: CompareAndSwap
		// The CAS failed, so we just retry. Eventually our CAS should
		// succeed or a future witness will pass us by and our witness
		// will end.
		goto WITNESS
	}
}
