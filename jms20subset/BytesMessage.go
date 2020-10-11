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

// BytesMessage is used to send a message containing a slice of bytes
//
// Instances of this object are created using the functions on the JMSContext
// such as CreateBytesMessage and CreateBytesMessageWithBytes.
type BytesMessage interface {

	// Encapsulate the root Message type so that this interface "inherits" the
	// accessors for standard attributes that apply to all message types, such
	// as GetJMSMessageID.
	Message

	// ReadBytes returns the bytes contained in this message's data.
	ReadBytes() *[]byte

	// WriteBytes sets the bytes for this message's data.
	WriteBytes(bytes []byte)

	GetBodyLength() int
}
