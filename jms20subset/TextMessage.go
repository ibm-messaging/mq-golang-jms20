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

// TextMessage is used to send a message containing a string.
//
// Instances of this object are created using the functions on the JMSContext
// such as CreateTextMessage and CreateTextMessageWithString.
type TextMessage interface {

	// Encapsulate the root Message type so that this interface "inherits" the
	// accessors for standard attributes that apply to all message types, such
	// as GetJMSMessageID.
	Message

	// GetText returns the string containing this message's data.
	GetText() *string

	// SetText sets the string containing this message's data.
	SetText(newBody string)
}
