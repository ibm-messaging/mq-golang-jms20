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

// JMSException represents an interface for returning details of a
// condition that has caused a function call to fail.
//
// It includes provider-specific Reason and ErrorCode attributes, and an
// optional reference to an Error describing the low-level problem.
type JMSException interface {
	GetReason() string
	GetErrorCode() string
	GetLinkedError() error
}

// JMSExceptionImpl is a struct that implements the JMSException interface
type JMSExceptionImpl struct {
	reason    string
	errorCode string
	linkedErr error
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

	ex := JMSExceptionImpl{
		reason:    reason,
		errorCode: errorCode,
		linkedErr: linkedErr,
	}

	return ex
}
