// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin

import (
	"fmt"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/seccomp"
	"github.com/snapcore/snapd/interfaces/udev"
)

const ioPortsControlSummary = `allows access to all I/O ports`

const ioPortsControlBaseDeclarationSlots = `
  io-ports-control:
    allow-installation:
      slot-snap-type:
        - core
    deny-auto-connection: true
`

const ioPortsControlConnectedPlugAppArmor = `
# Description: Allow write access to all I/O ports.
# See 'man 4 mem' for details.

capability sys_rawio, # required by iopl

/dev/port rw,
`

const ioPortsControlConnectedPlugSecComp = `
# Description: Allow changes to the I/O port permissions and
# privilege level of the calling process.  In addition to granting
# unrestricted I/O port access, running at a higher I/O privilege
# level also allows the process to disable interrupts.  This will
# probably crash the system, and is not recommended.
ioperm
iopl
`

// The type for io-ports-control interface
type ioPortsControlInterface struct{}

// Getter for the name of the io-ports-control interface
func (iface *ioPortsControlInterface) Name() string {
	return "io-ports-control"
}

func (iface *ioPortsControlInterface) StaticInfo() interfaces.StaticInfo {
	return interfaces.StaticInfo{
		Summary:              ioPortsControlSummary,
		ImplicitOnCore:       true,
		ImplicitOnClassic:    true,
		BaseDeclarationSlots: ioPortsControlBaseDeclarationSlots,
	}
}

func (iface *ioPortsControlInterface) String() string {
	return iface.Name()
}

// Check validity of the defined slot
func (iface *ioPortsControlInterface) SanitizeSlot(slot *interfaces.Slot) error {
	// Does it have right type?
	if iface.Name() != slot.Interface {
		panic(fmt.Sprintf("slot is not of interface %q", iface))
	}

	// Creation of the slot of this type
	// is allowed only by a gadget or os snap
	if !(slot.Snap.Type == "os") {
		return fmt.Errorf("%s slots only allowed on core snap", iface.Name())
	}
	return nil
}

// Checks and possibly modifies a plug
func (iface *ioPortsControlInterface) SanitizePlug(plug *interfaces.Plug) error {
	if iface.Name() != plug.Interface {
		panic(fmt.Sprintf("plug is not of interface %q", iface))
	}
	// Currently nothing is checked on the plug side
	return nil
}

func (iface *ioPortsControlInterface) AppArmorConnectedPlug(spec *apparmor.Specification, plug *interfaces.Plug, plugAttrs map[string]interface{}, slot *interfaces.Slot, slotAttrs map[string]interface{}) error {
	spec.AddSnippet(ioPortsControlConnectedPlugAppArmor)
	return nil
}

func (iface *ioPortsControlInterface) UDevConnectedPlug(spec *udev.Specification, plug *interfaces.Plug, plugAttrs map[string]interface{}, slot *interfaces.Slot, slotAttrs map[string]interface{}) error {
	const udevRule = `KERNEL=="port", TAG+="%s"`
	for appName := range plug.Apps {
		tag := udevSnapSecurityName(plug.Snap.Name(), appName)
		spec.AddSnippet(fmt.Sprintf(udevRule, tag))
	}
	return nil
}

func (iface *ioPortsControlInterface) SecCompConnectedPlug(spec *seccomp.Specification, plug *interfaces.Plug, plugAttrs map[string]interface{}, slot *interfaces.Slot, slotAttrs map[string]interface{}) error {
	spec.AddSnippet(ioPortsControlConnectedPlugSecComp)
	return nil
}

func (iface *ioPortsControlInterface) AutoConnect(*interfaces.Plug, *interfaces.Slot) bool {
	// Allow what is allowed in the declarations
	return true
}

func init() {
	registerIface(&ioPortsControlInterface{})
}
