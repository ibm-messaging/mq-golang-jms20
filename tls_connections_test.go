/*
 * Copyright (c) IBM Corporation 2019
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	"fmt"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
	"testing"
)

// This test file contains tests that demonstrate how to create TLS connections
// to a queue manager.
//
// Note that instructions on how to configure the queue manager TLS channels
// and the corresponding certificate keystores required by these tests can be
// found in the ./tls-samples/README.md file.

/*
 * Test that we can connect successfully if we provide the correct anonymous
 * ("ony way") TLS configuration.
 */
func TestAnonymousTLSConnection(t *testing.T) {

	cf, err := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, err)

	// Override the connection settings to point to the anonymous ("one way") TLS
	// channel (must be configured on the queue manager)
	cf.ChannelName = "TLS.ANON.SVRCONN"

	// Set the channel settings that tells the client what TLS configuration to use
	// to connect to the queue manager.
	cf.TLSCipherSpec = "ANY_TLS12"
	cf.KeyRepository = "./tls-samples/anon-tls" // points to .kdb file

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	if errCtx != nil && errCtx.GetReason() == "MQRC_UNKNOWN_CHANNEL_NAME" {
		// See ./tls-samples/README.md for details on how to configure the required channel.
		fmt.Println("Skipping TestAnonymousTLSConnection as required channel is not defined.")
		return
	}

	if errCtx != nil && errCtx.GetReason() == "MQRC_NOT_AUTHORIZED" {
		// See ./tls-samples/README.md for details on how to configure the required channel.
		fmt.Println("TLS connection was successfully negotiated, but client was blocked from connecting.")
		// Allow test to fail below.
	}

	// This connection should have been created successfully.
	assert.Nil(t, errCtx)

}

/*
 * Test that we can connect successfully if we provide the correct mutual
 * TLS configuration.
 */
func TestMutualTLSConnection(t *testing.T) {

	cf, err := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, err)

	// Override the connection settings to point to the mutual TLS
	// channel (must be configured on the queue manager)
	cf.ChannelName = "TLS.MUTUAL.SVRCONN"

	// Set the channel settings that tells the client what TLS configuration to use
	// to connect to the queue manager.
	cf.TLSCipherSpec = "ANY_TLS12"
	cf.TLSClientAuth = mqjms.TLSClientAuth_REQUIRED
	cf.CertificateLabel = "SampleClientA"         // point to the client certificate
	cf.KeyRepository = "./tls-samples/mutual-tls" // points to .kdb file

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	if errCtx != nil && errCtx.GetReason() == "MQRC_UNKNOWN_CHANNEL_NAME" {
		// See ./tls-samples/README.md for details on how to configure the required channel.
		fmt.Println("Skipping TestMutualTLSConnection as required channel is not defined.")
		return
	}

	if errCtx != nil && errCtx.GetReason() == "MQRC_NOT_AUTHORIZED" {
		// See ./tls-samples/README.md for details on how to configure the required channel.
		fmt.Println("TLS connection was successfully negotiated, but client was blocked from connecting.")
		// Allow test to fail below.
	}

	// This connection should have been created successfully.
	assert.Nil(t, errCtx)
	// MQRC_SSL_INITIALIZATION_ERROR is most likely to indicate that the client certificate
	// is not trusted by the queue manager - make sure you have run "REFRESH SECURITY TYPE(SSL)".
	//
	// Also see the queue manager error log for more details.

}

/*
 * Test that we get the correct error message when connecting to a TLS channel
 * without providing any TLS configuration.
 */
func TestNonTLSConnectionFails(t *testing.T) {

	cf, err := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, err)

	// Override the connection settings to point to the anonymous ("one way") TLS
	// channel (must be configured on the queue manager)
	cf.ChannelName = "TLS.ANON.SVRCONN"

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	// We expect this test to fail because we are trying to connect to a TLS channel
	// without providing any TLS information.
	assert.NotNil(t, errCtx)

	if errCtx.GetReason() == "MQRC_UNKNOWN_CHANNEL_NAME" {
		// See ./tls-samples/README.md for details on how to configure the required channel.
		fmt.Println("Skipping TestNonTLSConnectionFails as required channel is not defined.")
		return
	}

	assert.Equal(t, "MQRC_SSL_INITIALIZATION_ERROR", errCtx.GetReason())
	assert.Equal(t, "2393", errCtx.GetErrorCode())

}

/*
 * Test that we get the correct error message when connecting to a mutual TLS channel
 * when we only provide anonymous TLS configuration.
 */
func TestAnonConfigToMutualTLSConnectionFails(t *testing.T) {

	cf, err := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, err)

	// Override the connection settings to point to the anonymous ("one way") TLS
	// channel (must be configured on the queue manager)
	cf.ChannelName = "TLS.MUTUAL.SVRCONN"

	// Set the channel settings that tells the client what TLS configuration to use
	// to connect to the queue manager.
	cf.TLSCipherSpec = "ANY_TLS12"
	cf.KeyRepository = "./tls-samples/anon-tls" // Deliberately pointing to "wrong" config.

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	// We expect this test to fail because we are trying to connect to a TLS channel
	// without providing any TLS information.
	assert.NotNil(t, errCtx)

	if errCtx.GetReason() == "MQRC_UNKNOWN_CHANNEL_NAME" {
		// See ./tls-samples/README.md for details on how to configure the required channel.
		fmt.Println("Skipping TestAnonConfigToMutualTLSConnectionFails as required channel is not defined.")
		return
	}

	assert.Equal(t, "MQRC_SSL_INITIALIZATION_ERROR", errCtx.GetReason())
	assert.Equal(t, "2393", errCtx.GetErrorCode())

}

/*
 * Test that we get an error code when we give an invalid value for the
 * TLSClientAuth parameter.
 */
func TestInvalidClientAuthValue(t *testing.T) {

	cf, err := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, err)

	// Override the connection settings to point to the anonymous ("one way") TLS
	// channel (must be configured on the queue manager)
	cf.ChannelName = "TLS.ANON.SVRCONN"

	// Set the channel settings that tells the client what TLS configuration to use
	// to connect to the queue manager.
	cf.TLSCipherSpec = "ANY_TLS12"
	cf.TLSClientAuth = "INVALID_VALUE!"
	cf.KeyRepository = "./tls-samples/anon-tls" // points to .kdb file

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	assert.NotNil(t, errCtx)
	if errCtx != nil {
		assert.Equal(t, "MQRC_CD_ERROR", errCtx.GetReason())
		assert.Equal(t, "2277", errCtx.GetErrorCode())
	}

}
