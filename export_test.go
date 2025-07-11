// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

func IDPartFromCaller() string {
	pc := caller()
	return idPartFromCaller(pc)
}

func (d *DebugUI) ContainerCounter() int {
	return len(d.ctx.idToContainer)
}
