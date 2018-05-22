// Mgmt
// Copyright (C) 2013-2018+ James Shubin and the project contributors
// Written by James Shubin <james@shubin.ca> and the project contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package resources

import (
	"fmt"

	"github.com/purpleidea/mgmt/engine"
	"github.com/purpleidea/mgmt/engine/traits"
)

func init() {
	engine.RegisterResource("print", func() engine.Res { return &PrintRes{} })
}

// PrintRes is a resource that is useful for printing a message to the screen.
// It will also display a message when it receives a notification. It supports
// automatic grouping.
type PrintRes struct {
	traits.Base // add the base methods without re-implementation
	traits.Groupable
	traits.Recvable
	traits.Refreshable

	init *engine.Init

	Msg string `lang:"msg" yaml:"msg"` // the message to display
}

// Default returns some sensible defaults for this resource.
func (obj *PrintRes) Default() engine.Res {
	return &PrintRes{}
}

// Validate if the params passed in are valid data.
func (obj *PrintRes) Validate() error {
	return nil
}

// Init runs some startup code for this resource.
func (obj *PrintRes) Init(init *engine.Init) error {
	obj.init = init // save for later

	return nil
}

// Close is run by the engine to clean up after the resource is done.
func (obj *PrintRes) Close() error {
	return nil
}

// Watch is the primary listener for this resource and it outputs events.
func (obj *PrintRes) Watch() error {
	// notify engine that we're running
	if err := obj.init.Running(); err != nil {
		return err // exit if requested
	}

	var send = false // send event?
	for {
		select {
		case event, ok := <-obj.init.Events:
			if !ok {
				return nil
			}
			if err := obj.init.Read(event); err != nil {
				return err
			}
		}

		// do all our event sending all together to avoid duplicate msgs
		if send {
			send = false
			if err := obj.init.Event(); err != nil {
				return err // exit if requested
			}
		}
	}
}

// CheckApply method for Print resource. Does nothing, returns happy!
func (obj *PrintRes) CheckApply(apply bool) (checkOK bool, err error) {
	obj.init.Logf("CheckApply: %t", apply)
	if val, exists := obj.init.Recv()["Msg"]; exists && val.Changed {
		// if we received on Msg, and it changed, log message
		obj.init.Logf("CheckApply: Received `Msg` of: %s", obj.Msg)
	}

	if obj.init.Refresh() {
		obj.init.Logf("Received a notification!")
	}
	obj.init.Logf("Msg: %s", obj.Msg)
	if g := obj.GetGroup(); len(g) > 0 { // add any grouped elements
		for _, x := range g {
			print, ok := x.(*PrintRes) // convert from Res
			if !ok {
				panic(fmt.Sprintf("grouped member %v is not a %s", x, obj.Kind()))
			}
			obj.init.Logf("%s: Msg: %s", print, print.Msg)
		}
	}
	return true, nil // state is always okay
}

// Cmp compares two resources and returns an error if they are not equivalent.
func (obj *PrintRes) Cmp(r engine.Res) error {
	if !obj.Compare(r) {
		return fmt.Errorf("did not compare")
	}
	return nil
}

// Compare two resources and return if they are equivalent.
func (obj *PrintRes) Compare(r engine.Res) bool {
	// we can only compare PrintRes to others of the same resource kind
	res, ok := r.(*PrintRes)
	if !ok {
		return false
	}

	if obj.Msg != res.Msg {
		return false
	}
	return true
}

// PrintUID is the UID struct for PrintRes.
type PrintUID struct {
	engine.BaseUID
	name string
}

// UIDs includes all params to make a unique identification of this object.
// Most resources only return one, although some resources can return multiple.
func (obj *PrintRes) UIDs() []engine.ResUID {
	x := &PrintUID{
		BaseUID: engine.BaseUID{Name: obj.Name(), Kind: obj.Kind()},
		name:    obj.Name(),
	}
	return []engine.ResUID{x}
}

// GroupCmp returns whether two resources can be grouped together or not.
func (obj *PrintRes) GroupCmp(r engine.GroupableRes) error {
	_, ok := r.(*PrintRes)
	if !ok {
		return fmt.Errorf("resource is not the same kind")
	}
	return nil // grouped together if we were asked to
}

// UnmarshalYAML is the custom unmarshal handler for this struct.
// It is primarily useful for setting the defaults.
func (obj *PrintRes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawRes PrintRes // indirection to avoid infinite recursion

	def := obj.Default()       // get the default
	res, ok := def.(*PrintRes) // put in the right format
	if !ok {
		return fmt.Errorf("could not convert to PrintRes")
	}
	raw := rawRes(*res) // convert; the defaults go here

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*obj = PrintRes(raw) // restore from indirection with type conversion!
	return nil
}
