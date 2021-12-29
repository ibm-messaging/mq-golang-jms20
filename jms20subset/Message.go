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

// Message is the root interface of all JMS Messages. It defines the
// common message header attributes used for all messages.
//
// Instances of message objects are created using the functions on the JMSContext
// such as CreateTextMessage.
type Message interface {

	// GetJMSMessageID returns the ID of the message that uniquely identifies
	// each message sent by the provider.
	GetJMSMessageID() string

	// GetJMSTimestamp returns the message timestamp at which the message was
	// handed off to the provider to be sent.
	GetJMSTimestamp() int64

	// SetJMSCorrelationID sets the correlation ID for the message which can be
	// used to link on message to another. A typical use is to link a response
	// message with its request message.
	SetJMSCorrelationID(correlID string) JMSException

	// GetJMSCorrelationID returns the correlation ID of this message.
	GetJMSCorrelationID() string

	// SetJMSReplyTo sets the Destination to which a reply to this message should
	// be sent. If it is nil then no reply is expected.
	SetJMSReplyTo(dest Destination) JMSException

	// GetJMSReplyTo returns the Destination object to which a reply to this
	// message should be sent.
	GetJMSReplyTo() Destination

	// GetJMSDeliveryMode returns the delivery mode that is specified for this
	// message.
	//
	// Typical values returned by this method include
	// jms20subset.DeliveryMode_PERSISTENT and jms20subset.DeliveryMode_NON_PERSISTENT
	GetJMSDeliveryMode() int

	// SetStringProperty enables an application to set a string-type message property.
	//
	// value is *string which allows a nil value to be specified, to unset an individual
	// property.
	SetStringProperty(name string, value *string) JMSException

	// GetStringProperty returns the string value of a named message property.
	// Returns nil if the named property is not set.
	GetStringProperty(name string) (*string, JMSException)

	// SetIntProperty enables an application to set a int-type message property.
	SetIntProperty(name string, value int) JMSException

	// GetIntProperty returns the int value of a named message property.
	// Returns 0 if the named property is not set.
	GetIntProperty(name string) (int, JMSException)

	// PropertyExists returns true if the named message property exists on this message.
	PropertyExists(name string) (bool, JMSException)

	// GetPropertyNames returns a slice of strings containing the name of every message
	// property on this message.
	// Returns a zero length slice if no message properties are defined.
	GetPropertyNames() ([]string, JMSException)

	// ClearProperties removes all message properties from this message.
	ClearProperties() JMSException
}
