package wifitriggers

import (
	"net"
)

// Represents a Cond bound to an Action. Can be created using a fluent syntax:
//
//   wifitriggers.If(someCond).Then(someAction)
//
type Binding func(connectedClients []net.HardwareAddr) Action

type BindingChain Binding

var NilBindingChain = BindingChain(
	func(connectedClients []net.HardwareAddr) Action { return NilAction })

// Type alias used to create a Binding. Intended to be used as
//
//   wifitriggers.If(someCond).Then(someAction)
//
type If Cond

func (i If) Then(action Action) Binding {
	return func(connectedClients []net.HardwareAddr) Action {
		if (Cond(i))(connectedClients) {
			return action
		} else {
			return NilAction
		}
	}
}

// Adds a Binding to the chain.
func (c BindingChain) AddBinding(b Binding) BindingChain {
	return func(connectedClients []net.HardwareAddr) Action {
		return c(connectedClients).And(b(connectedClients))
	}
}

// Combines two chains.
func (c BindingChain) And(other BindingChain) BindingChain {
	return func(connectedClients []net.HardwareAddr) Action {
		return c(connectedClients).And(other(connectedClients))
	}
}
