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

// JMSContextAUTOACKNOWLEDGE is used to specify a sessionMode that automatically acknowledge message transmission.
const JMSContextAUTOACKNOWLEDGE int = 1

// JMSContextSESSIONTRANSACTED is used to specify a sessionMode that requires manual commit/rollback of transactions.
const JMSContextSESSIONTRANSACTED int = 0

// JMSContext represents a connection to the messaging provider, and
// provides the capability for applications to create Producer and Consumer
// objects so that it can send and receive messages.
type JMSContext interface {

	// CreateProducer creates a new producer object that can be used to configure
	// and send messages.
	//
	// Note that the Destination object is always supplied when making the
	// individual producer.Send calls, and not as part of creating the producer
	// itself.
	CreateProducer() JMSProducer

	// CreateConsumer creates a consumer for the specified Destination so that
	// an application can receive messages from that Destination.
	CreateConsumer(dest Destination) (JMSConsumer, JMSException)

	// CreateConsumer creates a consumer for the specified Destination using a
	// message selector, so that an application can receive messages from a
	// Destination that match the selector criteria.
	//
	// Note that since Golang does not allow multiple functions with the same
	// name and different parameters we must use a different function name.
	CreateConsumerWithSelector(dest Destination, selector string) (JMSConsumer, JMSException)

	// CreateBrowser creates a consumer for the specified Destination so that
	// an application can look at messages without removing them.
	CreateBrowser(dest Destination) (QueueBrowser, JMSException)

	// CreateQueue creates a queue object which encapsulates a provider specific
	// queue name.
	//
	// Note that this method does not create the physical queue in the JMS
	// provider. Creating a physical queue is typically an administrative task
	// performed by an administrator using provider-specific tooling.
	CreateQueue(queueName string) Queue

	// CreateTextMessage creates a message object that is used to send a string
	// from one application to another.
	CreateTextMessage() TextMessage

	// CreateTextMessageWithString creates an initialized text message object
	// containing the string that needs to be sent.
	//
	// Note that since Golang does not allow multiple functions with the same
	// name and different parameters we must use a different function name.
	CreateTextMessageWithString(txt string) TextMessage

	// CreateBytesMessage creates a message object that is used to send a slice
	// of bytes from one application to another.
	CreateBytesMessage() BytesMessage

	// CreateBytesMessageWithBytes creates a message object that is used to send a slice
	// of bytes from one application to another.
	CreateBytesMessageWithBytes(bytes []byte) BytesMessage

	// Commit confirms all messages sent/received during this transaction.
	Commit() JMSException

	// Rollback releases all messages sent/received during this transaction.
	Rollback() JMSException

	// Closes the connection to the messaging provider.
	//
	// Since the provider typically allocates significant resources on behalf of
	// a connection applications should close these resources when they are not
	// needed.
	Close()
}
