// Derived from the Eclipse Project for JMS, available at;
//     https://github.com/eclipse-ee4j/jms-api
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

//
package jms20subset

// Queue encapsulates a provider-specific queue name through which an
// application can carry out point to point messaging. It is the way a client
// specifies the identity of a queue to the JMS API functions.
type Queue interface {

	// GetQueueName returns the provider-specific name of the queue that is
	// represented by this object.
	GetQueueName() string

	// GetDestinationName returns the provider-specific name of the queue that is
	// represented by this object.
	//
	// This method is implemented to allow us to consider the Queue interface
	// as a specialization of the Destination interface.
	GetDestinationName() string
}
