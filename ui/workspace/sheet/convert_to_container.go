/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package sheet

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type containerConversionListUndoEdit = *unison.UndoEdit[*containerConversionList]

type containerConversionList struct {
	Owner widget.Rebuildable
	List  []*containerConversion
}

func (c *containerConversionList) Apply() {
	for _, one := range c.List {
		one.Apply()
	}
	c.Owner.Rebuild(true)
}

type containerConversion struct {
	Target *gurps.Equipment
	Type   string
}

func newContainerConversion(target *gurps.Equipment) *containerConversion {
	return &containerConversion{
		Target: target,
		Type:   target.Type,
	}
}

func (c *containerConversion) Apply() {
	c.Target.Type = c.Type
}

func canConvertToContainer(table *unison.Table[*ntable.Node[*gurps.Equipment]]) bool {
	for _, row := range table.SelectedRows(false) {
		if eqp := row.Data(); eqp != nil && !eqp.Container() {
			return true
		}
	}
	return false
}

func convertToContainer(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.Equipment]]) {
	before := &containerConversionList{Owner: owner}
	after := &containerConversionList{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if eqp := row.Data(); eqp != nil && !eqp.Container() {
			before.List = append(before.List, newContainerConversion(eqp))
			eqp.Type += gurps.ContainerKeyPostfix
			after.List = append(after.List, newContainerConversion(eqp))
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			mgr.Add(&unison.UndoEdit[*containerConversionList]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Convert to Container"),
				UndoFunc:   func(edit containerConversionListUndoEdit) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit containerConversionListUndoEdit) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		owner.Rebuild(true)
	}
}
