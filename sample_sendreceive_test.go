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
	"testing"

	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/stretchr/testify/assert"
)

/*
 * Demonstrate a fully worked example of a send/receive application including
 * good practice around error handling.
 */
func TestSampleSendReceiveWithErrorHandling(t *testing.T) {

	// Create a ConnectionFactory using some property files
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()

	if cfErr != nil {
		fmt.Println("Error during CreateConnecionFactory: ")
		fmt.Println(cfErr)
	}

	// Creates a connection to the queue manager.
	// We use a "defer" call to make sure that the connection is closed at the end of the method
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	// Check for errors - note that typically the application would terminate or
	// short cut its execution path once an error like this has been detected.
	if errCtx != nil {

		fmt.Println("Error during CreateContext: ")
		fmt.Println(errCtx)

	} else {

		// Create a Queue object that points at an IBM MQ queue
		queue := context.CreateQueue("DEV.QUEUE.1")

		// Send a message to the queue that contains the specified text string
		errSend := context.CreateProducer().SendString(queue, "My first message")

		// Check for errors. In this case we'll allow the application to continue,
		// but there won't be a message to receive.
		if errSend != nil {
			fmt.Println("Error during SendString: ")
			fmt.Println(errSend)
		}

		// Create a consumer, using Defer to make sure it gets closed at the end of the method
		consumer, errCons := context.CreateConsumer(queue)
		if consumer != nil {
			defer consumer.Close()
		}

		// Consumer was created successfully
		if errCons == nil {

			// Receive a message from the queue and return the string from the message body
			rcvBody, rcvErr := consumer.ReceiveStringBodyNoWait()

			if rcvErr != nil {
				fmt.Println("Error received:")
				fmt.Println(rcvErr)
			} else {

				// Successful creation of the Consumer
				if rcvBody != nil {
					fmt.Println("Received text string: " + *rcvBody)
				} else {
					fmt.Println("No message received")
				}

			}

		} else {
			fmt.Println("Error during CreateConsumer: ")
			fmt.Println(errCons)
		}

	}

}

/*
 * Demonstrate a send/receive application with limited error handling - this code
 * is used as part of the Readme for this repo in order to illustrate the key
 * functional steps using this JMS style library.
 */
func TestSampleSendReceiveWithLimitedErrorHandling(t *testing.T) {

	// Create a ConnectionFactory using some external property files
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	if cfErr != nil {
		fmt.Println(cfErr)
	}

	// Creates a connection to the queue manager.
	// We use a "defer" call to make sure that the connection is closed at the end of the method
	context, ctxErr := cf.CreateContext()
	if ctxErr != nil {
		fmt.Println(ctxErr)
	}
	if context != nil {
		defer context.Close()
	}

	// Create a Queue object that points at an IBM MQ queue
	queue := context.CreateQueue("DEV.QUEUE.1")

	// Send a message to the queue that contains the specified text string
	context.CreateProducer().SendString(queue, "My first message")

	// Create a consumer, using Defer to make sure it gets closed at the end of the method
	consumer, conErr := context.CreateConsumer(queue)
	if conErr != nil {
		fmt.Println(conErr)
	}
	if consumer != nil {
		defer consumer.Close()
	}

	// Receive a message from the queue and return the string from the message body
	rcvBody, rcvErr := consumer.ReceiveStringBodyNoWait()

	if rcvErr != nil {
		fmt.Println("Error received:")
		fmt.Println(rcvErr)
	}

	if rcvBody != nil {
		fmt.Println("Received text string: " + *rcvBody)
	} else {
		fmt.Println("No message received")
	}

}

/*
 * Test for receiving a message when there is no message available.
 */
func TestReceiveNoWaitWithoutMessage(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	if cfErr != nil {
		fmt.Println(cfErr)
	}

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	if ctxErr != nil {
		fmt.Println(ctxErr)
	}
	if context != nil {
		defer context.Close()
	}

	// Equivalent to a JNDI lookup or other declarative definition
	queue := context.CreateQueue("DEV.QUEUE.2")

	// Try to receive a message when there isn't one available.
	consumer, conErr := context.CreateConsumer(queue)
	if conErr != nil {
		fmt.Println(conErr)
	}
	if consumer != nil {
		defer consumer.Close()
	}

	rcvBody, rcvErr := consumer.ReceiveStringBodyNoWait()

	if rcvErr != nil {

		fmt.Println("Error received:")
		fmt.Println(rcvErr)

	} else {

		// No error on the receive, so we can check for the message.
		assert.Nil(t, rcvBody)
		if rcvBody != nil {
			fmt.Println("Received unexpected message: " + *rcvBody)
		}

	}

}
