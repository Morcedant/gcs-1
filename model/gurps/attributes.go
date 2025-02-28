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

package gurps

import (
	"bytes"
	"sort"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/crc"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

// Attributes holds a set of Attribute objects.
type Attributes struct {
	Set map[string]*Attribute
}

// NewAttributes creates a new Attributes.
func NewAttributes(entity *Entity) *Attributes {
	a := &Attributes{Set: make(map[string]*Attribute)}
	for attrID, def := range entity.SheetSettings.Attributes.Set {
		a.Set[attrID] = NewAttribute(entity, attrID, def.Order)
	}
	return a
}

// MarshalJSON implements json.Marshaler.
func (a *Attributes) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(a.List())
	return buffer.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Attributes) UnmarshalJSON(data []byte) error {
	var list []*Attribute
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	a.Set = make(map[string]*Attribute, len(list))
	for i, one := range list {
		one.Order = i
		a.Set[one.ID()] = one
	}
	return nil
}

// Clone a copy of this.
func (a *Attributes) Clone(entity *Entity) *Attributes {
	clone := &Attributes{Set: make(map[string]*Attribute)}
	for k, v := range a.Set {
		clone.Set[k] = v.Clone(entity)
	}
	return clone
}

// List returns the map of Attribute objects as an ordered list.
func (a *Attributes) List() []*Attribute {
	list := make([]*Attribute, 0, len(a.Set))
	for _, v := range a.Set {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order < list[j].Order
	})
	return list
}

// CRC64 calculates a CRC-64 for this data.
func (a *Attributes) CRC64() uint64 {
	c := crc.Number(0, len(a.Set))
	for _, one := range a.List() {
		c = one.crc64(c)
	}
	return c
}

// Cost returns the points spent for the specified Attribute.
func (a *Attributes) Cost(attrID string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		return attr.PointCost()
	}
	return 0
}

// Current resolves the given attribute ID to its current value, or fxp.Min.
func (a *Attributes) Current(attrID string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		return attr.Current()
	}
	if v, err := fxp.FromString(attrID); err == nil {
		return v
	}
	return fxp.Min
}

// Maximum resolves the given attribute ID to its maximum value, or fxp.Min.
func (a *Attributes) Maximum(attrID string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		return attr.Maximum()
	}
	if v, err := fxp.FromString(attrID); err == nil {
		return v
	}
	return fxp.Min
}

// PoolThreshold resolves the given attribute ID and state to the value for its pool threshold, or fxp.Min.
func (a *Attributes) PoolThreshold(attrID, state string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		if def := attr.AttributeDef(); def != nil {
			for _, one := range def.Thresholds {
				if strings.EqualFold(one.State, state) {
					return one.Threshold(attr.Entity)
				}
			}
		}
	}
	return fxp.Min
}
