// Copyright (c) IBM Corporation 2019.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

//
package mqjms

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

// CreateConnectionFactoryFromDefaultJSONFiles is a utility method that creates
// a JMS ConnectionFactory object that is populated with properties from values
// stored in two external files on the file system.
//
// This method reads the following files;
//   - $HOME/Downloads/connection_info.json   for host/port/channel information
//   - $HOME/Downloads/apiKey.json            for username/password information
//
// If your queue manager is hosted on the IBM MQ on Cloud service then you can
// download these two files directly from the IBM Cloud service console.
// Alternative you can use the example files provided in the /config-samples
// directory of this repository and populate them with the details of your own
// queue manager.
func CreateConnectionFactoryFromDefaultJSONFiles() (cf ConnectionFactoryImpl, err error) {
	return CreateConnectionFactoryFromJSON("", "")
}

// CreateConnectionFactoryFromJSON is a utility method that creates
// a JMS ConnectionFactory object that is populated with properties from values
// stored in two external files on the file system.
//
// The calling application provides the fully qualified path to the location of
// the file as the two parameters. If empty string is provided then the default
// location and name is assumed as follows;
//   - $HOME/Downloads/connection_info.json   for host/port/channel information
//   - $HOME/Downloads/apiKey.json            for username/password information
//
// If your queue manager is hosted on the IBM MQ on Cloud service then you can
// download these two files directly from the IBM Cloud service console.
// Alternative you can use the example files provided in the /config-samples
// directory of this repository and populate them with the details of your own
// queue manager.
func CreateConnectionFactoryFromJSON(connectionInfoLocn string, apiKeyLocn string) (cf ConnectionFactoryImpl, err error) {

	// If the caller has not explicitly specified a path in which to find these
	// files then assume that they are in the default location (/Downloads)
	if connectionInfoLocn == "" {
		connectionInfoLocn = os.Getenv("HOME") + "/Downloads/connection_info.json"
	}

	if apiKeyLocn == "" {
		apiKeyLocn = os.Getenv("HOME") + "/Downloads/applicationApiKey.json"
	}

	// Attempt to read the connection info file at the specified location.
	// If we get an error then there is no way to proceed successfully, so
	// terminate the program immediately.
	connInfoContent, err := ioutil.ReadFile(connectionInfoLocn)
	if err != nil {
		log.Print("Error reading file from " + connectionInfoLocn)
		return ConnectionFactoryImpl{}, err
	}

	apiKeyContent, err := ioutil.ReadFile(apiKeyLocn)
	if err != nil {
		log.Print("Error reading file from " + apiKeyLocn)
		return ConnectionFactoryImpl{}, err
	}

	// Having successfully opened the connection info file, unmarshall the
	// JSON into a map and parse out the individual values that we need to
	// use in order to populate the ConnectionFactory object.
	var connInfoMap map[string]*json.RawMessage
	err = json.Unmarshal(connInfoContent, &connInfoMap)

	if err != nil {
		log.Print("Failure during unmarshalling file from JSON: " + connectionInfoLocn)
		return ConnectionFactoryImpl{}, err
	}

	var qmName, hostname, appChannel string
	var port int

	qmName, errQM := parseStringValueFromJSON("queueManagerName", connInfoMap, connectionInfoLocn)
	if errQM != nil {
		return ConnectionFactoryImpl{}, errQM
	}

	hostname, errHost := parseStringValueFromJSON("hostname", connInfoMap, connectionInfoLocn)
	if errHost != nil {
		return ConnectionFactoryImpl{}, errHost
	}

	port, errPort := parseIntValueFromJSON("listenerPort", connInfoMap, connectionInfoLocn)
	if errPort != nil {
		return ConnectionFactoryImpl{}, errPort
	}

	appChannel, errChannel := parseStringValueFromJSON("applicationChannelName", connInfoMap, connectionInfoLocn)
	if errChannel != nil {
		return ConnectionFactoryImpl{}, errChannel
	}

	// Now unmarshall and parse out the values from the api key file (that
	// contains the username/password credentials).
	var apiKeyMap map[string]*json.RawMessage
	err = json.Unmarshal(apiKeyContent, &apiKeyMap)

	if err != nil {
		log.Print("Failure during unmarshalling file from JSON: " + apiKeyLocn)
		return ConnectionFactoryImpl{}, err
	}

	var username, password string

	username, errUser := parseStringValueFromJSON("mqUsername", apiKeyMap, apiKeyLocn)
	if errUser != nil {
		return ConnectionFactoryImpl{}, errUser
	}

	password, errPassword := parseStringValueFromJSON("apiKey", apiKeyMap, apiKeyLocn)
	if errPassword != nil {
		return ConnectionFactoryImpl{}, errPassword
	}

	// Use the parsed values to initialize the attributes of the Impl object.
	cf = ConnectionFactoryImpl{
		QMName:      qmName,
		Hostname:    hostname,
		PortNumber:  port,
		ChannelName: appChannel,
		UserName:    username,
		Password:    password,
	}

	// Give the populated ConnectionFactory back to the caller.
	return cf, nil

}

// Extract a specified string value from the map that we generated from a JSON object
func parseStringValueFromJSON(attributeName string, mapData map[string]*json.RawMessage, fileName string) (value string, err error) {

	var valueStr string

	if mapData[attributeName] == nil {
		return "", errors.New("Unable to find " + attributeName + " in " + fileName)
	}

	err = json.Unmarshal(*mapData[attributeName], &valueStr)

	return valueStr, err

}

// Extract a specified int value from the map that we generated from a JSON object
func parseIntValueFromJSON(attributeName string, mapData map[string]*json.RawMessage, fileName string) (value int, err error) {

	var valueNum int

	if mapData[attributeName] == nil {
		return 0, errors.New("Unable to find " + attributeName + " in " + fileName)
	}

	err = json.Unmarshal(*mapData[attributeName], &valueNum)

	return valueNum, err

}
