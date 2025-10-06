package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/buntdb"
)

type Editor struct {
	app        *tview.Application
	db         *buntdb.DB
	pages      *tview.Pages
	keysList   *tview.List
	textArea   *tview.TextArea
	statusBar  *tview.TextView
	currentKey string
}

func NewEditor() (*Editor, error) {
	e := &Editor{
		app:       tview.NewApplication(),
		db:        db,
		pages:     tview.NewPages(),
		keysList:  tview.NewList(),
		textArea:  tview.NewTextArea(),
		statusBar: tview.NewTextView(),
	}

	e.setupUI()
	return e, nil
}

func (e *Editor) setupUI() {
	// 状态栏
	e.statusBar.
		SetDynamicColors(true).
		SetText("[yellow]BuntDB Editor[white] | [green]Ctrl+S[white] Save | [green]Ctrl+Q[white] Quit | [green]Ctrl+N[white] New | [green]Ctrl+D[white] Delete")

	// 键列表
	e.keysList.
		ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" Keys ").
		SetTitleAlign(tview.AlignLeft)

	e.loadKeys()

	// 文本编辑器
	e.textArea.
		SetBorder(true).
		SetTitle(" Value Editor ").
		SetTitleAlign(tview.AlignLeft)

	// 键列表选择事件
	e.keysList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		e.loadValue(mainText)
	})

	// 主布局
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(e.keysList, 0, 1, true).
			AddItem(e.textArea, 0, 3, false), 0, 1, true).
		AddItem(e.statusBar, 1, 0, false)

	// 全局快捷键
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			e.saveCurrentValue()
			return nil
		case tcell.KeyCtrlQ:
			e.quit()
			return nil
		case tcell.KeyCtrlN:
			e.createNewKey()
			return nil
		case tcell.KeyCtrlD:
			e.deleteCurrentKey()
			return nil
		case tcell.KeyTab:
			// 在列表和编辑器之间切换焦点
			if e.app.GetFocus() == e.keysList {
				e.app.SetFocus(e.textArea)
			} else {
				e.app.SetFocus(e.keysList)
			}
			return nil
		}
		return event
	})

	e.pages.AddPage("main", flex, true, true)
	e.app.SetRoot(e.pages, true)
}

func (e *Editor) loadKeys() {
	e.keysList.Clear()

	var count int
	e.db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(key, value string) bool {
			// 截断显示值
			displayValue := value
			if len(displayValue) > 30 {
				displayValue = displayValue[:30] + "..."
			}
			displayValue = strings.ReplaceAll(displayValue, "\n", "\\n")

			e.keysList.AddItem(key, displayValue, 0, nil)
			count++
			return true
		})
	})

	e.updateStatus(fmt.Sprintf("Loaded %d keys", count))
}

func (e *Editor) loadValue(key string) {
	e.currentKey = key

	var value string
	err := e.db.View(func(tx *buntdb.Tx) error {
		var err error
		value, err = tx.Get(key)
		return err
	})

	if err != nil {
		e.updateStatus(fmt.Sprintf("[red]Error loading key: %v", err))
		return
	}

	text, err := formatYamlText(value)
	if err != nil {
		e.updateStatus(fmt.Sprintf("[red]Error parsing key: %v", err))
		return
	}

	e.textArea.SetText(text, true)
	e.textArea.SetTitle(fmt.Sprintf(" Editing: %s ", key))
	e.updateStatus(fmt.Sprintf("Loaded key: %s (%d bytes)", key, len(value)))
	e.app.SetFocus(e.textArea)
}

func (e *Editor) saveCurrentValue() {
	if e.currentKey == "" {
		e.updateStatus("[yellow]No key selected")
		return
	}

	value := e.textArea.GetText()

	err := e.db.Update(func(tx *buntdb.Tx) error {
		formated, err := formatJsonText(value)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(e.currentKey, formated, nil)
		return err
	})

	if err != nil {
		e.updateStatus(fmt.Sprintf("[red]Error saving: %v", err))
		return
	}

	e.updateStatus(fmt.Sprintf("[green]Saved: %s (%d bytes)", e.currentKey, len(value)))
	e.loadKeys() // 刷新列表
}

func (e *Editor) createNewKey() {
	// 创建输入对话框
	form := tview.NewForm()
	form.
		AddInputField("Key name:", "", 40, nil, nil).
		AddButton("Create", func() {
			keyInput := form.GetFormItem(0).(*tview.InputField)
			newKey := keyInput.GetText()

			if newKey == "" {
				e.updateStatus("[yellow]Key name cannot be empty")
				e.pages.RemovePage("dialog")
				return
			}

			// 创建新键
			err := e.db.Update(func(tx *buntdb.Tx) error {
				_, _, err := tx.Set(newKey, "", nil)
				return err
			})

			if err != nil {
				e.updateStatus(fmt.Sprintf("[red]Error creating key: %v", err))
			} else {
				e.updateStatus(fmt.Sprintf("[green]Created key: %s", newKey))
				e.loadKeys()
				e.currentKey = newKey
				e.textArea.SetText("", true)
				e.textArea.SetTitle(fmt.Sprintf(" Editing: %s ", newKey))
				e.app.SetFocus(e.textArea)
			}

			e.pages.RemovePage("dialog")
		}).
		AddButton("Cancel", func() {
			e.pages.RemovePage("dialog")
		})

	form.SetBorder(true).SetTitle(" Create New Key ").SetTitleAlign(tview.AlignLeft)

	// 居中显示对话框
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 7, 1, true).
			AddItem(nil, 0, 1, false), 50, 1, true).
		AddItem(nil, 0, 1, false)

	e.pages.AddPage("dialog", modal, true, true)
}

func (e *Editor) deleteCurrentKey() {
	if e.currentKey == "" {
		e.updateStatus("[yellow]No key selected")
		return
	}

	keyToDelete := e.currentKey

	// 确认对话框
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete key '%s'?", keyToDelete)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				err := e.db.Update(func(tx *buntdb.Tx) error {
					_, err := tx.Delete(keyToDelete)
					return err
				})

				if err != nil {
					e.updateStatus(fmt.Sprintf("[red]Error deleting: %v", err))
				} else {
					e.updateStatus(fmt.Sprintf("[green]Deleted: %s", keyToDelete))
					e.currentKey = ""
					e.textArea.SetText("", true)
					e.textArea.SetTitle(" Value Editor ")
					e.loadKeys()
				}
			}
			e.pages.RemovePage("confirm")
		})

	e.pages.AddPage("confirm", modal, true, true)
}

func (e *Editor) updateStatus(message string) {
	e.statusBar.SetText(fmt.Sprintf("[yellow]BuntDB Editor[white] | %s | [green]Tab[white] Switch | [green]Ctrl+S[white] Save | [green]Ctrl+N[white] New | [green]Ctrl+D[white] Del | [green]Ctrl+Q[white] Quit", message))
}

func (e *Editor) quit() {
	modal := tview.NewModal().
		SetText("Quit BuntDB Editor?").
		AddButtons([]string{"Quit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				e.app.Stop()
			}
			e.pages.RemovePage("quit")
		})

	e.pages.AddPage("quit", modal, true, true)
}

func (e *Editor) Run() error {
	defer e.db.Close()
	return e.app.Run()
}

func formatYamlText(input string) (string, error) {
	var node interface{}

	err := yaml.Unmarshal([]byte(input), &node)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// 重新编码
	output := bytes.NewBuffer(nil)
	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	err = encoder.Encode(node)

	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}
	return output.String(), nil
}

func formatJsonText(input string) (string, error) {
	var node interface{}
	err := yaml.Unmarshal([]byte(input), &node)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	output, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}
	return string(output), nil
}
