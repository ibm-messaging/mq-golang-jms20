// Copyright (c) IBM Corporation 2022.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package mqjms provides the implementation of the JMS style Golang interfaces to communicate with IBM MQ.
package mqjms

import (
	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	ibmmq "github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

// BrowserImpl represents the JMS QueueBrowser object that allows applications
// to peek at messages on a queue without destructively consuming them.
type BrowserImpl struct {
	browseOption *int32
	ConsumerImpl // Browser is a specialized form of consumer
}

// GetEnumeration returns an iterator for browsing the current
// queue messages in the order they would be received.
//
// In this implementation there is exactly one Enumeration per
// QueueBrowser. If an application wants to browse two independent
// copies of the messages it must create two QueueBrowsers.
func (browser *BrowserImpl) GetEnumeration() (jms20subset.MessageIterator, jms20subset.JMSException) {

	// A browser is just an alternative view of a Consumer that
	// presents slightly different functions + behaviour.
	return browser, nil

}

// GetNext returns the next Message that is available
// or else nil if no messages are available.
func (browser *BrowserImpl) GetNext() (jms20subset.Message, jms20subset.JMSException) {

	// Like a ReceiveNoWait, but with Browse turned on.
	gmo := ibmmq.NewMQGMO()
	gmo.Options |= *browser.browseOption

	msg, err := browser.receiveInternal(gmo)

	if err == nil {
		// After we have browsed the first message successfully we move on to asking
		// for the "next" message from this point onwards.
		brse := int32(ibmmq.MQGMO_BROWSE_NEXT)
		browser.browseOption = &brse
	}

	return msg, err
}
