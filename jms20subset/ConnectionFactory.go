// Derived from the Eclipse Project for JMS, available at;
//     https://github.com/eclipse-ee4j/jms-api
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Interfaces for messaging applications in the style of the Java Message Service (JMS) API.
package jms20subset

// ConnectionFactory defines a Golang interface which provides similar
// functionality as the Java JMS ConnectionFactory - encapsulating a set of
// connection configuration parameters that allows an application to create
// connections to a messaging provider.
type ConnectionFactory interface {

	// CreateContext creates a connection to the messaging provider using the
	// configuration parameters that are encapsulated by this ConnectionFactory.
	CreateContext() (JMSContext, JMSException)
}
