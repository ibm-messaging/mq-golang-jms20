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

// JMSConsumer provides the ability for an application to receive messages
// from a queue or a topic.
//
// Note that Golang doesn't provide Generics in the same way as Java so we have
// to create multiple separate functions to provide capability equivalent to the
// Java JMS receiveBody(Class<T>) method.
type JMSConsumer interface {

	// ReceiveNoWait receives the next message if one is available, or nil if
	// there is no message immediately available
	ReceiveNoWait() (Message, JMSException)

	// ReceiveStringBodyNoWait receives the next message for this JMSConsumer
	// and returns its body as a string. If a message is not immediately
	//available a nil is returned.
	ReceiveStringBodyNoWait() (*string, JMSException)
	GetStringProperty(p string)
	// Closes the JMSConsumer in order to free up any resources that were
	// allocated by the provider on behalf of this consumer.
	Close()
}
