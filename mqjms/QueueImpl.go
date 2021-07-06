// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package mqjms provides the implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqjms

import (
	"fmt"
	"strconv"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
)

// QueueImpl encapsulates the provider-specific attributes necessary to
// communicate with an IBM MQ queue.
type QueueImpl struct {
	queueName       string
	putAsyncAllowed int
}

// GetQueueName returns the provider-specific name of the queue that is
// represented by this object.
func (queue QueueImpl) GetQueueName() string {

	return queue.queueName

}

// GetDestinationName returns the name of the destination represented by this
// object.
func (queue QueueImpl) GetDestinationName() string {

	return queue.queueName

}

// SetPutAsyncAllowed allows the async allowed setting to be updated.
func (queue QueueImpl) SetPutAsyncAllowed(paa int) jms20subset.Queue {

	// Check that the specified paa parameter is one of the values that we permit,
	// and if so store that value inside queue.
	if paa == jms20subset.Destination_PUT_ASYNC_ALLOWED_ENABLED ||
		paa == jms20subset.Destination_PUT_ASYNC_ALLOWED_DISABLED ||
		paa == jms20subset.Destination_PUT_ASYNC_ALLOWED_AS_DEST {

		queue.putAsyncAllowed = paa

	} else {
		// Normally we would throw an error here to indicate that an invalid value
		// was specified, however we have decided that it is more useful to support
		// method chaining, which prevents us from returning an error object.
		// Instead we settle for printing an error message to the console.
		fmt.Println("Invalid PutAsyncAllowed value specified: " + strconv.Itoa(paa))
	}

	return queue
}

// GetPutAsyncAllowed returns the current setting for async put.
func (queue QueueImpl) GetPutAsyncAllowed() int {
	return queue.putAsyncAllowed
}
