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

// MessageIterator provides the ability for the application to consume
// a sequence of Messages.
type MessageIterator interface {

	// GetNext returns the next Message that is available
	// or else nil if no messages are available.
	GetNext() (Message, JMSException)
}
