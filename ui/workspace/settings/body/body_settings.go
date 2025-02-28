package body

import (
	"fmt"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	settings2 "github.com/richardwilkes/gcs/v5/ui/workspace/settings"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var _ widget.GroupedCloser = &bodyDockable{}

type bodyDockable struct {
	settings2.Dockable
	owner         widget.EntityPanel
	targetMgr     *widget.TargetMgr
	undoMgr       *unison.UndoManager
	bodyType      *gurps.Body
	originalCRC   uint64
	toolbar       *unison.Panel
	content       *unison.Panel
	applyButton   *unison.Button
	cancelButton  *unison.Button
	promptForSave bool
}

// ShowBodySettings the Body Settings. Pass in nil to edit the defaults or a sheet to edit the sheet's.
func ShowBodySettings(owner widget.EntityPanel) {
	ws, dc, found := workspace.Activate(func(d unison.Dockable) bool {
		if s, ok := d.(*bodyDockable); ok && owner == s.owner {
			return true
		}
		return false
	})
	if !found && ws != nil {
		d := &bodyDockable{
			owner:         owner,
			promptForSave: true,
		}
		d.Self = d
		d.targetMgr = widget.NewTargetMgr(d)
		if owner != nil {
			entity := d.owner.Entity()
			d.bodyType = entity.SheetSettings.BodyType.Clone(entity, nil)
			d.TabTitle = i18n.Text("Body Type: " + owner.Entity().Profile.Name)
		} else {
			d.bodyType = settings.Global().Sheet.BodyType.Clone(nil, nil)
			d.TabTitle = i18n.Text("Default Body Type")
		}
		d.TabIcon = res.BodyTypeSVG
		d.bodyType.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
		d.originalCRC = d.bodyType.CRC64()
		d.Extensions = []string{".body", ".ghl"}
		d.undoMgr = unison.NewUndoManager(100, func(err error) { jot.Error(err) })
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.ModifiedCallback = d.modified
		d.WillCloseCallback = d.willClose
		d.Setup(ws, dc, d.addToStartToolbar, nil, d.initContent)
	}
}

func (d *bodyDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

func (d *bodyDockable) modified() bool {
	modified := d.originalCRC != d.bodyType.CRC64()
	d.applyButton.SetEnabled(modified)
	d.cancelButton.SetEnabled(modified)
	return modified
}

func (d *bodyDockable) willClose() bool {
	if d.promptForSave && d.originalCRC != d.bodyType.CRC64() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Apply changes made to\n%s?"), d.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			d.apply()
		case unison.ModalResponseCancel:
			return false
		}
	}
	return true
}

func (d *bodyDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

func (d *bodyDockable) addToStartToolbar(toolbar *unison.Panel) {
	d.toolbar = toolbar
	d.applyButton = unison.NewSVGButton(res.CheckmarkSVG)
	d.applyButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Apply Changes"))
	d.applyButton.SetEnabled(false)
	d.applyButton.ClickCallback = func() {
		d.apply()
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.applyButton)

	d.cancelButton = unison.NewSVGButton(res.NotSVG)
	d.cancelButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Discard Changes"))
	d.cancelButton.SetEnabled(false)
	d.cancelButton.ClickCallback = func() {
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.cancelButton)
}

func (d *bodyDockable) initContent(content *unison.Panel) {
	d.content = content
	content.SetBorder(nil)
	content.SetLayout(&unison.FlexLayout{Columns: 1})
	// TODO: Insert content here
}

func (d *bodyDockable) Entity() *gurps.Entity {
	if d.owner != nil {
		return d.owner.Entity()
	}
	return nil
}

func (d *bodyDockable) applyBodyType(bodyType *gurps.Body) {
	d.bodyType = bodyType.Clone(d.Entity(), nil)
	d.sync()
}

func (d *bodyDockable) reset() {
	entity := d.Entity()
	undo := &unison.UndoEdit[*gurps.Body]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Reset Body Type"),
		UndoFunc:   func(e *unison.UndoEdit[*gurps.Body]) { d.applyBodyType(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*gurps.Body]) { d.applyBodyType(e.AfterData) },
		AbsorbFunc: func(e *unison.UndoEdit[*gurps.Body], other unison.Undoable) bool { return false },
		BeforeData: d.bodyType.Clone(entity, nil),
	}
	if d.owner != nil {
		entity.SheetSettings.BodyType = settings.Global().Sheet.BodyType.Clone(entity, nil)
	} else {
		settings.Global().Sheet.BodyType = gurps.FactoryBody()
	}
	d.bodyType.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	undo.AfterData = d.bodyType.Clone(entity, nil)
	d.UndoManager().Add(undo)
	d.sync()
}

func (d *bodyDockable) sync() {
	var focusRefKey string
	if focus := d.Window().Focus(); unison.AncestorOrSelf[*bodyDockable](focus) == d {
		focusRefKey = focus.RefKey
	}
	scrollRoot := d.content.ScrollRoot()
	h, v := scrollRoot.Position()
	d.content.RemoveAllChildren()
	// TODO: Rebuild display here
	d.MarkForLayoutAndRedraw()
	d.ValidateLayout()
	d.MarkModified()
	if focusRefKey != "" {
		if focus := d.targetMgr.Find(focusRefKey); focus != nil {
			focus.RequestFocus()
		} else {
			widget.FocusFirstContent(d.toolbar, d.content)
		}
	}
	scrollRoot.SetPosition(h, v)
}

func (d *bodyDockable) load(fileSystem fs.FS, filePath string) error {
	bodyType, err := gurps.NewBodyFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	bodyType.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	entity := d.Entity()
	undo := &unison.UndoEdit[*gurps.Body]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Load Body Type"),
		UndoFunc:   func(e *unison.UndoEdit[*gurps.Body]) { d.applyBodyType(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*gurps.Body]) { d.applyBodyType(e.AfterData) },
		AbsorbFunc: func(e *unison.UndoEdit[*gurps.Body], other unison.Undoable) bool { return false },
		BeforeData: d.bodyType.Clone(entity, nil),
	}
	d.bodyType = bodyType
	undo.AfterData = d.bodyType.Clone(entity, nil)
	d.UndoManager().Add(undo)
	d.sync()
	return nil
}

func (d *bodyDockable) save(filePath string) error {
	return d.bodyType.Save(filePath)
}

func (d *bodyDockable) apply() {
	d.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	if d.owner == nil {
		settings.Global().Sheet.BodyType = d.bodyType.Clone(nil, nil)
		return
	}
	entity := d.owner.Entity()
	entity.SheetSettings.BodyType = d.bodyType.Clone(entity, nil)
	for _, wnd := range unison.Windows() {
		if ws := workspace.FromWindow(wnd); ws != nil {
			ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
				for _, one := range dc.Dockables() {
					if s, ok := one.(gurps.SheetSettingsResponder); ok {
						s.SheetSettingsUpdated(entity, true)
					}
				}
				return false
			})
		}
	}
}
