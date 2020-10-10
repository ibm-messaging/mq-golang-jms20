// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package mqjms provides the implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqjms

// TextMessageImpl contains the IBM MQ specific attributes necessary to
// present a message that carries a string.
type TextMessageImpl struct {
	bodyStr     *string
	MessageImpl // embed the "parent" message object that defines the basic behaviour
}

// GetText returns the string that is contained in this TextMessage.
func (msg *TextMessageImpl) GetText() *string {

	return msg.bodyStr

}

// SetText stores the supplied string so that it can be transmitted as part
// of this TextMessage.
func (msg *TextMessageImpl) SetText(newBody string) {

	msg.bodyStr = &newBody

}
