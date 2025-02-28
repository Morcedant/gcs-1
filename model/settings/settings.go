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

package settings

import (
	"context"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xio/fs/paths"
	"github.com/richardwilkes/unison"
)

const maxRecentFiles = 20

var global *Settings

// NavigatorSettings holds settings for the navigator view.
type NavigatorSettings struct {
	DividerPosition float32  `json:"divider_position"`
	OpenRowKeys     []string `json:"open_row_keys,omitempty"`
}

// Settings holds the application settings.
type Settings struct {
	LastSeenGCSVersion string               `json:"last_seen_gcs_version,omitempty"`
	General            *settings.General    `json:"general,omitempty"`
	LibrarySet         library.Libraries    `json:"libraries,omitempty"`
	LibraryExplorer    NavigatorSettings    `json:"library_explorer"`
	RecentFiles        []string             `json:"recent_files,omitempty"`
	LastDirs           map[string]string    `json:"last_dirs,omitempty"`
	PageRefs           PageRefs             `json:"page_refs,omitempty"`
	KeyBindings        KeyBindings          `json:"key_bindings,omitempty"`
	WorkspaceFrame     *unison.Rect         `json:"workspace_frame,omitempty"`
	Colors             theme.Colors         `json:"colors"`
	Fonts              theme.Fonts          `json:"fonts"`
	QuickExports       *gurps.QuickExports  `json:"quick_exports,omitempty"`
	Sheet              *gurps.SheetSettings `json:"sheet_settings,omitempty"`
}

// Default returns new default settings.
func Default() *Settings {
	return &Settings{
		LastSeenGCSVersion: cmdline.AppVersion,
		General:            settings.NewGeneral(),
		LibrarySet:         library.NewLibraries(),
		LibraryExplorer:    NavigatorSettings{DividerPosition: 300},
		LastDirs:           make(map[string]string),
		QuickExports:       gurps.NewQuickExports(),
		Sheet:              gurps.FactorySheetSettings(),
	}
}

// Global returns the global settings.
func Global() *Settings {
	if global == nil {
		dice.GURPSFormat = true
		if err := jio.LoadFromFile(context.Background(), Path(), &global); err != nil {
			global = Default()
		}
		global.EnsureValidity()
		gurps.SettingsProvider = global
		gurps.InstallEvaluatorFunctions(fxp.EvalFuncs)
		global.Colors.MakeCurrent()
		global.Fonts.MakeCurrent()
	}
	return global
}

// Save to the standard path.
func (s *Settings) Save() error {
	return jio.SaveToFile(context.Background(), Path(), s)
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (s *Settings) EnsureValidity() {
	if s.General == nil {
		s.General = settings.NewGeneral()
	} else {
		s.General.EnsureValidity()
	}
	if len(s.LibrarySet) == 0 {
		s.LibrarySet = library.NewLibraries()
	}
	if s.LastDirs == nil {
		s.LastDirs = make(map[string]string)
	}
	if s.QuickExports == nil {
		s.QuickExports = gurps.NewQuickExports()
	}
	if s.Sheet == nil {
		s.Sheet = gurps.FactorySheetSettings()
	} else {
		s.Sheet.EnsureValidity()
	}
}

// ListRecentFiles returns the current list of recently opened files. Files that are no longer readable for any reason
// are omitted.
func (s *Settings) ListRecentFiles() []string {
	list := make([]string, 0, len(s.RecentFiles))
	for _, one := range s.RecentFiles {
		if fs.FileIsReadable(one) {
			list = append(list, one)
		}
	}
	if len(list) != len(s.RecentFiles) {
		s.RecentFiles = make([]string, len(list))
		copy(s.RecentFiles, list)
	}
	return list
}

// AddRecentFile adds a file path to the list of recently opened files.
func (s *Settings) AddRecentFile(filePath string) {
	ext := path.Ext(filePath)
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == toolbox.MacOS || runtime.GOOS == toolbox.WindowsOS {
		ext = strings.ToLower(ext)
	}
	for _, one := range library.AcceptableExtensions() {
		if one == ext {
			full, err := filepath.Abs(filePath)
			if err != nil {
				return
			}
			if fs.FileIsReadable(full) {
				for i, f := range s.RecentFiles {
					if f == full {
						copy(s.RecentFiles[i:], s.RecentFiles[i+1:])
						s.RecentFiles[len(s.RecentFiles)-1] = ""
						s.RecentFiles = s.RecentFiles[:len(s.RecentFiles)-1]
						break
					}
				}
				s.RecentFiles = append(s.RecentFiles, "")
				copy(s.RecentFiles[1:], s.RecentFiles)
				s.RecentFiles[0] = full
				if len(s.RecentFiles) > maxRecentFiles {
					s.RecentFiles = s.RecentFiles[:maxRecentFiles]
				}
			}
			return
		}
	}
}

// GeneralSettings implements gurps.SettingsProvider.
func (s *Settings) GeneralSettings() *settings.General {
	return s.General
}

// SheetSettings implements gurps.SettingsProvider.
func (s *Settings) SheetSettings() *gurps.SheetSettings {
	return s.Sheet
}

// Libraries implements gurps.SettingsProvider.
func (s *Settings) Libraries() library.Libraries {
	return s.LibrarySet
}

// Path returns the path to our settings file.
func Path() string {
	return filepath.Join(paths.AppDataDir(), cmdline.AppCmdName+"_prefs.json")
}
