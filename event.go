// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

// EventHandler is an interface for handling events in a widget.
type EventHandler interface {
	// On registers a callback function to be called when the event occurs.
	On(func())
}

type eventHandler struct {
}

func (o *eventHandler) On(f func()) {
	f()
}

type nullEventHandler struct{}

func (n *nullEventHandler) On(func()) {}
