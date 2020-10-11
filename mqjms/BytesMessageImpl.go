// Copyright (c) IBM Corporation 2020.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package mqjms provides the implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqjms

// BytesMessageImpl contains the IBM MQ specific attributes necessary to
// present a message that carries a slice of bytes
type BytesMessageImpl struct {
	bodyBytes   *[]byte
	MessageImpl // embed the "parent" message object that defines the basic behaviour
}

// ReadBytes returns the string that is contained in this BytesMessage.
func (msg *BytesMessageImpl) ReadBytes() *[]byte {

	if msg.bodyBytes == nil {
		return &[]byte{}
	}
	return msg.bodyBytes

}

// WriteBytes stores the supplied slice of bytes so that it can be transmitted as part
// of this BytesMessage.
func (msg *BytesMessageImpl) WriteBytes(bytes []byte) {

	msg.bodyBytes = &bytes

}

// GetBodyLength returns the length of the bytes that are stored in this message
func (msg *BytesMessageImpl) GetBodyLength() int {

	length := 0

	if msg.bodyBytes != nil {
		length = len(*msg.bodyBytes)
	}

	return length

}
