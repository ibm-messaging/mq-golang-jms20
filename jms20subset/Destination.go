// Derived from the Eclipse Project for JMS, available at;
//     https://github.com/eclipse-ee4j/jms-api
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package jms20subset provides interfaces for messaging applications in the style of the Java Message Service (JMS) API.
package jms20subset

// Destination encapsulates a provider-specific address, which is typically
// specialized as either a Queue or a Topic.
type Destination interface {

	// GetDestinationName returns the name of the destination represented by this
	// object.
	//
	// In Java JMS this interface contains no methods and is only a parent for the
	// Queue and Topic interfaces, however in Golang an interface with no methods
	// is automatically implemented by every object, so we need to define at least
	// one method here in order to make it meet the JMS style semantics.
	GetDestinationName() string

	// SetPutAsyncAllowed controls whether asynchronous put is allowed for this
	// destination.
	//
	// Permitted values are:
	//  * Destination_PUT_ASYNC_ALLOWED_ENABLED - enables async put
	//  * Destination_PUT_ASYNC_ALLOWED_DISABLED - disables async put
	//  * Destination_PUT_ASYNC_ALLOWED_AS_DEST - delegate to queue configuration (default)
	SetPutAsyncAllowed(paa int) Queue

	// GetPutAsyncAllowed returns whether asynchronous put is configured for this
	// destination.
	//
	// Returned value is one of:
	//  * Destination_PUT_ASYNC_ALLOWED_ENABLED - async put is enabled
	//  * Destination_PUT_ASYNC_ALLOWED_DISABLED - async put is disabled
	//  * Destination_PUT_ASYNC_ALLOWED_AS_DEST - delegated to queue configuration (default)
	GetPutAsyncAllowed() int
}

// Destination_PUT_ASYNC_ALLOWED_ENABLED is used to enable messages being sent asynchronously.
const Destination_PUT_ASYNC_ALLOWED_ENABLED int = 1

// Destination_PUT_ASYNC_ALLOWED_DISABLED is used to disable messages being sent asynchronously.
const Destination_PUT_ASYNC_ALLOWED_DISABLED int = 0

// Destination_PUT_ASYNC_ALLOWED_AS_DEST allows the async message behaviour to be controlled by
// the queue on the queue manager.
const Destination_PUT_ASYNC_ALLOWED_AS_DEST int = -1
