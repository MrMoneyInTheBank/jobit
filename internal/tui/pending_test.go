package tui

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"testing"
)

func TestPendingWritesBlockQuit(t *testing.T) {
	for _, deleting := range []bool{false, true} {
		j := InitJobList(fixture())
		j.saving = !deleting
		j.deleting = deleting
		for _, k := range []string{"ctrl+c", "q"} {
			cmd := key(&j, k)
			if cmd != nil {
				t.Fatalf("%s quit during write", k)
			}
		}
		for _, msg := range []tea.Msg{tea.QuitMsg{}, tea.InterruptMsg{}} {
			blocked := ProtectPendingWrite(&j, msg)
			if _, ok := blocked.(pendingQuitMsg); !ok {
				t.Fatal("signal bypassed protection")
			}
			j.Update(blocked)
			if !j.quitBlocked {
				t.Fatal("missing pending notice")
			}
		}
		j.saving = false
		j.deleting = false
		if ProtectPendingWrite(&j, tea.QuitMsg{}) == nil {
			t.Fatal("idle quit blocked")
		}
	}
}

func TestPendingPopupClearsAfterCompletion(t *testing.T) {
	for _, deleting := range []bool{false, true} {
		j := InitJobList(fixture())
		j.saving = !deleting
		j.deleting = deleting
		if deleting {
			j.screen = deleteScreen
		} else {
			j.screen = addScreen
		}
		key(&j, "ctrl+c")
		for _, size := range [][2]int{{80, 24}, {24, 9}, {10, 3}} {
			j.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			view := j.View()
			if !strings.Contains(strings.ToLower(view), "wait") && size[0] >= 24 {
				t.Fatal("missing wait notice")
			}
			if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
				t.Fatalf("popup overflow at %v", size)
			}
		}
		if deleting {
			j.Update(deletedMsg{err: errors.New("failed")})
		} else {
			j.Update(savedMsg{err: errors.New("failed")})
		}
		if j.quitBlocked || j.Busy() {
			t.Fatal("completion did not dismiss popup")
		}
	}
}
