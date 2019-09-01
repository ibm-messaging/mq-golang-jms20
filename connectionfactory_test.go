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
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

/*
 * Test the ability to populate a ConnectionFactory from contents of the two
 * JSON files stored on the file system.
 */
func TestLoadCFFromJSON(t *testing.T) {

	// Loads CF parameters from connection_info.json and apiKey.json in the Downloads directory
	cf, err := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()

	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, mqjms.TransportType_CLIENT, cf.TransportType)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	if errCtx != nil {
		log.Fatal(errCtx)
	}

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("DEV.QUEUE.1")

	// Send a message!
	context.CreateProducer().SendString(queue, "My first message")

	// Clean up the message to avoid problems with other tests.
	consumer, errCons := context.CreateConsumer(queue)

	if errCons != nil {
		log.Fatal(errCons)
	}

	consumer.ReceiveStringBodyNoWait()
	consumer.Close()

}

/*
 * Illustrate the ability to populate a ConnectionFactory object using code
 * of the application developer's choosing, which could be extended to implement
 * other helper methods for creating ConnectionFactory objects.
 */
func TestProgrammaticConnFactory(t *testing.T) {

	// Initialise the attributes of the CF in whatever way you like
	cf := mqjms.ConnectionFactoryImpl{
		QMName:      "QM_ONE",
		Hostname:    "random.hostname.com",
		PortNumber:  1414,
		ChannelName: "SYSTEM.APP.SVRCONN",
		UserName:    os.Getenv("MQ_SAMP_USERID"),
		Password:    os.Getenv("MQ_SAMP_PASSWORD"),
	}

	assert.NotNil(t, cf)

}
