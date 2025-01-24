package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var currentHighlightStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57")).
	Bold(true)

type model struct {
	table        table.Model
	fileCounts   map[string]int  // Stores the counts of all files
	showUpdated  bool            // Tracks whether to show the updated text
	changes      []string        // Stores all the changes made for the selected file
	highlightPos int             // Tracks the highlight position within the last 5 changes
	dirInput     textinput.Model // Input field for directory path
	dir          string          // Stores the directory path
	state        string          // Tracks the current state of the application ("input" or "main")
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter directory path"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	return model{
		dirInput: ti,
		state:    "input",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case "input":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.dir = m.dirInput.Value()
				if m.dir == "" {
					return m, nil
				}
				m.state = "main"

				filenames := []string{"Music.xml", "Event.xml", "Chara.xml", "NamePlate.xml", "AvatarAccessory.xml"}
				fileCounts, err := countSpecificXMLFiles(m.dir, filenames)
				if err != nil {
					log.Printf("Error scanning directory: %v\n", err)
					return m, tea.Quit
				}
				m.fileCounts = fileCounts

				columns := []table.Column{
					{Title: "Option", Width: 10},
					{Title: "File Name", Width: 20},
					{Title: "Total Number", Width: 15},
				}

				rows := []table.Row{
					{"1", "Music.xml", fmt.Sprintf("%d", fileCounts["Music.xml"])},
					{"2", "Event.xml", fmt.Sprintf("%d", fileCounts["Event.xml"])},
					{"3", "Chara.xml", fmt.Sprintf("%d", fileCounts["Chara.xml"])},
					{"4", "NamePlate.xml", fmt.Sprintf("%d", fileCounts["NamePlate.xml"])},
					{"5", "AvatarAccessory.xml", fmt.Sprintf("%d", fileCounts["AvatarAccessory.xml"])},
				}

				t := table.New(
					table.WithColumns(columns),
					table.WithRows(rows),
					table.WithFocused(true),
					table.WithHeight(6),
				)

				s := table.DefaultStyles()
				s.Header = s.Header.
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color("240")).
					BorderBottom(true).
					Bold(false)
				s.Selected = s.Selected.
					Foreground(lipgloss.Color("229")).
					Background(lipgloss.Color("57")).
					Bold(false)
				t.SetStyles(s)

				m.table = t
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

		m.dirInput, cmd = m.dirInput.Update(msg)
		return m, cmd

	case "main":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				if !m.showUpdated {
					m.changes = nil

					err := filepath.Walk(m.dir, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						selectedRow := m.table.SelectedRow()
						if len(selectedRow) == 0 {
							return nil
						}

						filename := selectedRow[1]

						if !info.IsDir() && info.Name() == filename {
							data, err := os.ReadFile(path)
							if err != nil {
								log.Printf("Error reading file %s: %v\n", path, err)
								return nil
							}

							content := string(data)
							var updatedContent string

							switch filename {
							case "Music.xml":
								if strings.Contains(content, "<firstLock>true</firstLock>") {
									// Unlock: Change <firstLock>true</firstLock> to <firstLock>false</firstLock>
									updatedContent = strings.Replace(content, "<firstLock>true</firstLock>", "<firstLock>false</firstLock>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <firstLock>true</firstLock> to <firstLock>false</firstLock>", path))
								} else if strings.Contains(content, "<firstLock>false</firstLock>") {
									// Relock: Change <firstLock>false</firstLock> to <firstLock>true</firstLock>
									updatedContent = strings.Replace(content, "<firstLock>false</firstLock>", "<firstLock>true</firstLock>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <firstLock>false</firstLock> to <firstLock>true</firstLock>", path))
								}
							case "Event.xml":
								if strings.Contains(content, "<alwaysOpen>true</alwaysOpen>") {
									updatedContent = strings.Replace(content, "<alwaysOpen>true</alwaysOpen>", "<alwaysOpen>false</alwaysOpen>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <alwaysOpen>true</alwaysOpen> to <alwaysOpen>false</alwaysOpen>", path))
								} else if strings.Contains(content, "<alwaysOpen>false</alwaysOpen>") {
									updatedContent = strings.Replace(content, "<alwaysOpen>false</alwaysOpen>", "<alwaysOpen>true</alwaysOpen>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <alwaysOpen>false</alwaysOpen> to <alwaysOpen>true</alwaysOpen>", path))
								}
							case "Chara.xml":
								if strings.Contains(content, "<defaultHave>true</defaultHave>") {
									// Relock: Change <defaultHave>true</defaultHave> to <defaultHave>false</defaultHave>
									updatedContent = strings.Replace(content, "<defaultHave>true</defaultHave>", "<defaultHave>false</defaultHave>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <defaultHave>true</defaultHave> to <defaultHave>false</defaultHave>", path))
								} else if strings.Contains(content, "<defaultHave>false</defaultHave>") {
									// Unlock: Change <defaultHave>false</defaultHave> to <defaultHave>true</defaultHave>
									updatedContent = strings.Replace(content, "<defaultHave>false</defaultHave>", "<defaultHave>true</defaultHave>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <defaultHave>false</defaultHave> to <defaultHave>true</defaultHave>", path))
								}
							case "NamePlate.xml":
								if strings.Contains(content, "<defaultHave>true</defaultHave>") {
									// Relock: Change <defaultHave>true</defaultHave> to <defaultHave>false</defaultHave>
									updatedContent = strings.Replace(content, "<defaultHave>true</defaultHave>", "<defaultHave>false</defaultHave>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <defaultHave>true</defaultHave> to <defaultHave>false</defaultHave>", path))
								} else if strings.Contains(content, "<defaultHave>false</defaultHave>") {
									// Unlock: Change <defaultHave>false</defaultHave> to <defaultHave>true</defaultHave>
									updatedContent = strings.Replace(content, "<defaultHave>false</defaultHave>", "<defaultHave>true</defaultHave>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <defaultHave>false</defaultHave> to <defaultHave>true</defaultHave>", path))
								}
							case "AvatarAccessory.xml":
								if strings.Contains(content, "<defaultHave>true</defaultHave>") {
									// Relock: Change <defaultHave>true</defaultHave> to <defaultHave>false</defaultHave>
									updatedContent = strings.Replace(content, "<defaultHave>true</defaultHave>", "<defaultHave>false</defaultHave>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <defaultHave>true</defaultHave> to <defaultHave>false</defaultHave>", path))
								} else if strings.Contains(content, "<defaultHave>false</defaultHave>") {
									// Unlock: Change <defaultHave>false</defaultHave> to <defaultHave>true</defaultHave>
									updatedContent = strings.Replace(content, "<defaultHave>false</defaultHave>", "<defaultHave>true</defaultHave>", 1)
									m.changes = append(m.changes, fmt.Sprintf("Updated %s: Changed <defaultHave>false</defaultHave> to <defaultHave>true</defaultHave>", path))
								}
							default:
								return nil
							}
							err = os.WriteFile(path, []byte(updatedContent), 0644)
							if err != nil {
								log.Printf("Error writing modified XML to file %s: %v\n", path, err)
								return nil
							}
						}
						return nil
					})
					if err != nil {
						log.Printf("Error walking directory: %v\n", err)
					}

					if len(m.changes) > 5 {
						m.changes = m.changes[len(m.changes)-5:]
					}

					m.showUpdated = true
					m.highlightPos = len(m.changes) - 1
				}
			case "b":
				if m.showUpdated {
					m.showUpdated = false
				}
			case "up":
				if m.showUpdated && m.highlightPos > 0 {
					m.highlightPos--
				}
			case "down":
				if m.showUpdated && m.highlightPos < len(m.changes)-1 {
					m.highlightPos++
				}
			case "1":
				m.table.SetCursor(0)
			case "2":
				m.table.SetCursor(1)
			case "3":
				m.table.SetCursor(2)
			case "4":
				m.table.SetCursor(3)
			case "5":
				m.table.SetCursor(4)
			}
		}

		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case "input":
		return fmt.Sprintf(
			"Enter the directory path:\n\n%s\n\nPress Enter to continue, or 'q' to quit.\n",
			m.dirInput.View(),
		)

	case "main":
		if m.showUpdated {
			if len(m.changes) == 0 {
				return baseStyle.Render("No changes made.") + "\nPress 'b' to go back.\n"
			}

			var changesText strings.Builder
			for i, change := range m.changes {
				if i == m.highlightPos {
					changesText.WriteString(currentHighlightStyle.Render(change) + "\n")
				} else {
					changesText.WriteString(change + "\n")
				}
			}

			return baseStyle.Render(changesText.String()) + "\nPress '↑' and '↓' to move highlight, 'b' to go back.\n"
		}
		return baseStyle.Render(m.table.View()) + "\nPress '1'-'5' to select, 'enter' to modify selected file, 'q' to quit.\n"
	}

	return ""
}

func countSpecificXMLFiles(dir string, filenames []string) (map[string]int, error) {
	counts := make(map[string]int)
	for _, filename := range filenames {
		counts[filename] = 0
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, filename := range filenames {
				if info.Name() == filename {
					counts[filename]++
				}
			}
		}
		return nil
	})
	return counts, err
}

func main() {
	m := initialModel()

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
