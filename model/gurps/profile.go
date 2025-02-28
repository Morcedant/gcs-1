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
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/ancestry"
	"github.com/richardwilkes/gcs/v5/model/gurps/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

// Standard height and width for the portrait
const (
	PortraitHeight = 96
	PortraitWidth  = 3 * PortraitHeight / 4
)

// Profile holds the profile information for an NPC.
type Profile struct {
	PlayerName        string         `json:"player_name,omitempty"`
	Name              string         `json:"name,omitempty"`
	Title             string         `json:"title,omitempty"`
	Organization      string         `json:"organization,omitempty"`
	Religion          string         `json:"religion,omitempty"`
	Age               string         `json:"age,omitempty"`
	Birthday          string         `json:"birthday,omitempty"`
	Eyes              string         `json:"eyes,omitempty"`
	Hair              string         `json:"hair,omitempty"`
	Skin              string         `json:"skin,omitempty"`
	Handedness        string         `json:"handedness,omitempty"`
	Gender            string         `json:"gender,omitempty"`
	TechLevel         string         `json:"tech_level,omitempty"`
	PortraitData      []byte         `json:"portrait,omitempty"`
	Height            measure.Length `json:"height,omitempty"`
	Weight            measure.Weight `json:"weight,omitempty"`
	SizeModifier      int            `json:"SM,omitempty"`
	SizeModifierBonus fxp.Int        `json:"-"`
	portrait          *unison.Image
}

// Update any derived values.
func (p *Profile) Update(entity *Entity) {
	p.SizeModifierBonus = entity.BonusFor(feature.AttributeIDPrefix+gid.SizeModifier, nil)
}

// Portrait returns the portrait image, if there is one.
func (p *Profile) Portrait() *unison.Image {
	if p.portrait == nil && len(p.PortraitData) != 0 {
		var err error
		if p.portrait, err = unison.NewImageFromBytes(p.PortraitData, 0.5); err != nil {
			jot.Error(errs.NewWithCause("unable to load portrait data", err))
			p.portrait = nil
			p.PortraitData = nil
			return nil
		}
	}
	return p.portrait
}

// AdjustedSizeModifier returns the adjusted size modifier.
func (p *Profile) AdjustedSizeModifier() int {
	return p.SizeModifier + fxp.As[int](p.SizeModifierBonus)
}

// SetAdjustedSizeModifier sets the adjusted size modifier.
func (p *Profile) SetAdjustedSizeModifier(value int) {
	if value != p.AdjustedSizeModifier() {
		// TODO: Need undo logic
		p.SizeModifier = value - fxp.As[int](p.SizeModifierBonus)
	}
}

// AutoFill fills in the default profile entries.
func (p *Profile) AutoFill(entity *Entity) {
	generalSettings := SettingsProvider.GeneralSettings()
	p.TechLevel = generalSettings.DefaultTechLevel
	p.PlayerName = generalSettings.DefaultPlayerName
	a := entity.Ancestry()
	p.Gender = a.RandomGender("")
	p.Age = strconv.Itoa(a.RandomAge(entity, p.Gender, 0))
	p.Eyes = a.RandomEyes(p.Gender, "")
	p.Hair = a.RandomHair(p.Gender, "")
	p.Skin = a.RandomSkin(p.Gender, "")
	p.Handedness = a.RandomHandedness(p.Gender, "")
	p.Height = a.RandomHeight(entity, p.Gender, 0)
	p.Weight = a.RandomWeight(entity, p.Gender, 0)
	p.Name = a.RandomName(ancestry.AvailableNameGenerators(SettingsProvider.Libraries()), p.Gender)
	p.Birthday = generalSettings.CalendarRef(SettingsProvider.Libraries()).RandomBirthday(p.Birthday)
}
