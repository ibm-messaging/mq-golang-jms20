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

// QueueBrowser provides the ability for an application to look at messages on
// a queue without removing them.
type QueueBrowser interface {

	// GetEnumeration returns an iterator for browsing the current
	// queue messages in the order they would be received.
	GetEnumeration() (MessageIterator, JMSException)

	// Closes the QueueBrowser in order to free up any resources that were
	// allocated by the provider.
	Close()
}
