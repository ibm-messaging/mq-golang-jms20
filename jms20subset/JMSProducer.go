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

// JMSProducer is a simple object used to send messages on behalf of a
// JMSContext. It provides various methods to send a message to a specified
// Destination. It also provides methods to allow message options to be
// specified prior to sending a message.
type JMSProducer interface {

	// Send a message to the specified Destination, using any message options
	// that are defined on this JMSProducer.
	Send(dest Destination, msg Message) JMSException

	// Send a TextMessage with the specified body to the specified Destination
	// using any message options that are defined on this JMSProducer.
	//
	// Note that since Golang does not allow multiple functions with the same
	// name and different parameters we must use a different function name.
	SendString(dest Destination, body string) JMSException

	// SetDeliveryMode sets the delivery mode of messages sent using this
	// JMSProducer - for example whether a message is persistent or non-persistent.
	//
	// Permitted arguments to this method are jms20subset.DeliveryMode_PERSISTENT
	// and jms20subset.DeliveryMode_NON_PERSISTENT.
	SetDeliveryMode(mode int) JMSProducer

	// GetDeliveryMode returns the delivery mode of messages that are send using
	// this JMSProducer.
	//
	// Typical values returned by this method include
	// jms20subset.DeliveryMode_PERSISTENT and jms20subset.DeliveryMode_NON_PERSISTENT
	GetDeliveryMode() int

	// SetTimeToLive sets the time to live (in milliseconds) for messages that
	// are sent using this JMSProducer.
	//
	// The expiration time of a message is the sum of this time to live and the
	// time at which the message is sent. A value of zero means that the message
	// never expires.
	SetTimeToLive(timeToLive int) JMSProducer

	// GetTimeToLive returns the time to live (in milliseconds) that will be
	// applied to messages that are sent using this JMSProducer.
	GetTimeToLive() int
}
