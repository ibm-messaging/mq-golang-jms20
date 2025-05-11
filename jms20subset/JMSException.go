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

// JMSException represents an interface for returning details of a
// condition that has caused a function call to fail.
//
// It includes provider-specific Reason and ErrorCode attributes, and an
// optional reference to an Error describing the low-level problem.
type JMSException interface {
	GetReason() string
	GetErrorCode() string
	GetLinkedError() error
	Error() string
}

// JMSExceptionImpl is a struct that implements the JMSException interface
type JMSExceptionImpl struct {
	reason        string
	errorCode     string
	linkedErr     error
	messageLength int
}

// GetReason returns the provider-specific reason string describing the error.
func (ex JMSExceptionImpl) GetReason() string {

	return ex.reason

}

// GetErrorCode returns the provider-specific error code describing the error.
func (ex JMSExceptionImpl) GetErrorCode() string {

	return ex.errorCode

}

// GetLinkedError returns the linked Error object representing the low-level
// problem, if one has been provided.
func (ex JMSExceptionImpl) GetLinkedError() error {

	return ex.linkedErr

}

// GetMessageLength gives the data length that was returned by a Receive (GET) call,
// for example if the message was truncated, or could not be read because the receive
// buffer was too small.
func (ex JMSExceptionImpl) GetMessageLength() int {

	return ex.messageLength

}

// Error allows the JMSExceptionImpl struct to be treated as a Golang error,
// while also returning a human readable string representation of the error.
func (ex JMSExceptionImpl) Error() string {

	var linkedStr string
	if ex.linkedErr != nil {
		linkedStr = ex.linkedErr.Error()
	}

	return "{errorCode=" + ex.errorCode + ", reason=" + ex.reason + ", linkedErr=" + linkedStr + "}"

}

// CreateJMSException is a helper function for creating a JMSException
func CreateJMSException(reason string, errorCode string, linkedErr error) JMSException {
	return CreateJMSExceptionWithExtraParams(reason, errorCode, linkedErr, 0)
}

// CreateJMSException is a helper function for creating a JMSException
func CreateJMSExceptionWithExtraParams(reason string, errorCode string, linkedErr error, messageLength int) JMSException {

	ex := JMSExceptionImpl{
		reason:        reason,
		errorCode:     errorCode,
		linkedErr:     linkedErr,
		messageLength: messageLength,
	}

	return ex
}
