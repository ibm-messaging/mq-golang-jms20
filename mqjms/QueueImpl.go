// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

//
package mqjms

// QueueImpl encapsulates the provider-specific attributes necessary to
// communicate with an IBM MQ queue.
type QueueImpl struct {
	queueName string
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
