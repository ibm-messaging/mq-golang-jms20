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

// Destination encapsulates a provider-specific address, which is typically
// specialized as either a Queue or a Topic.
type Destination interface {

	// GetDestinationName returns the name of the destination represented by this
	// object.
	//
	// In Java JMS this interface contains no methods and is only a parent for the
	// Queue and Topic interfaces, however in Golang an interface with no methods
	// is automatically implemented by every object, so we need to define at least
	// one method here in order to make it meet the JMS style semantics.
	GetDestinationName() string
}
