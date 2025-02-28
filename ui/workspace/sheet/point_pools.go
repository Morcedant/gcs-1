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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attribute"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

// PointPoolsPanel holds the contents of the point pools block on the sheet.
type PointPoolsPanel struct {
	unison.Panel
	entity *gurps.Entity
	crc    uint64
}

// NewPointPoolsPanel creates a new point pools panel.
func NewPointPoolsPanel(entity *gurps.Entity) *PointPoolsPanel {
	p := &PointPoolsPanel{entity: entity}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  6,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HSpan:  2,
	})
	p.SetBorder(unison.NewCompoundBorder(&widget.TitledBorder{Title: i18n.Text("Point Pools")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	p.crc = attrs.CRC64()
	p.rebuild(attrs)
	return p
}

func (p *PointPoolsPanel) rebuild(attrs *gurps.AttributeDefs) {
	p.RemoveAllChildren()
	for _, def := range attrs.List() {
		if def.Type != attribute.Pool {
			continue
		}
		attr, ok := p.entity.Attributes.Set[def.ID()]
		if !ok {
			jot.Warnf("unable to locate attribute data for '%s'", def.ID())
			continue
		}
		p.AddChild(p.createPointsField(attr))

		var currentField *widget.DecimalField
		currentField = widget.NewDecimalPageField(nil, "", i18n.Text("Point Pool Current"),
			func() fxp.Int {
				if currentField != nil {
					currentField.SetMinMax(currentField.Min(), attr.Maximum())
				}
				return attr.Current()
			},
			func(v fxp.Int) { attr.Damage = (attr.Maximum() - v).Max(0) }, fxp.Min, attr.Maximum(), true)
		p.AddChild(currentField)

		p.AddChild(widget.NewPageLabel(i18n.Text("of")))

		maximumField := widget.NewDecimalPageField(nil, "", i18n.Text("Point Pool Maximum"),
			func() fxp.Int { return attr.Maximum() },
			func(v fxp.Int) {
				attr.SetMaximum(v)
				currentField.SetMinMax(currentField.Min(), v)
				currentField.Sync()
			}, fxp.Min, fxp.Max, true)
		p.AddChild(maximumField)

		name := widget.NewPageLabel(def.Name)
		if def.FullName != "" {
			name.Tooltip = unison.NewTooltipWithText(def.FullName)
		}
		p.AddChild(name)

		if threshold := attr.CurrentThreshold(); threshold != nil {
			state := widget.NewPageLabel("[" + threshold.State + "]")
			if threshold.Explanation != "" {
				state.Tooltip = unison.NewTooltipWithText(threshold.Explanation)
			}
			p.AddChild(state)
		} else {
			p.AddChild(unison.NewPanel())
		}
	}
}

func (p *PointPoolsPanel) createPointsField(attr *gurps.Attribute) *widget.NonEditablePageField {
	field := widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		if text := "[" + attr.PointCost().String() + "]"; text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
		if def := attr.AttributeDef(); def != nil {
			f.Tooltip = unison.NewTooltipWithText(fmt.Sprintf(i18n.Text("Points spent on %s"), def.CombinedName()))
		}
	})
	field.Font = theme.PageFieldSecondaryFont
	return field
}

// Sync the panel to the current data.
func (p *PointPoolsPanel) Sync() {
	attrs := p.entity.Attributes
	if crc := attrs.CRC64(); crc != p.crc {
		p.crc = crc
		p.rebuild(gurps.SheetSettingsFor(p.entity).Attributes)
		widget.MarkForLayoutWithinDockable(p)
	}
}
