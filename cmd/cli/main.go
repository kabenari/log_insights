package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kabenari/log-insight/pkg/models"
)

type model struct {
	insights []models.AIResult
	cursor   int
	selected int
}

func initialModel() model {
	return model{
		insights: loadInsights(),
		selected: -1,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.insights)-1 {
				m.cursor++
			}
		case "enter", " ":
			// Toggle details view
			if m.selected == m.cursor {
				m.selected = -1 // Close if already open
			} else {
				m.selected = m.cursor // Open details
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if len(m.insights) == 0 {
		return "No insights found yet. Waiting for Worker...\nPress 'q' to quit."
	}

	s := "\n  LOG INSIGHT DASHBOARD \n"
	s += "  =====================\n\n"

	for i, item := range m.insights {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, item.OriginalLog.Level, item.OriginalLog.Message)

		if m.selected == i {
			s += fmt.Sprintf("\n      \033[36mAI Analysis:\033[0m %s\n", item.Analysis)
			s += fmt.Sprintf("      \033[90mTimestamp: %s\033[0m\n\n", item.OriginalLog.Timestamp)
		}
	}

	s += "\n  [Use Arrows to Move] [Enter to Expand] [q to Quit]\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func loadInsights() []models.AIResult {
	file, err := os.Open("insights.jsonl")
	if err != nil {
		return []models.AIResult{}
	}
	defer file.Close()

	var results []models.AIResult
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var res models.AIResult
		if err := json.Unmarshal(scanner.Bytes(), &res); err == nil {
			results = append([]models.AIResult{res}, results...)
		}
	}
	return results
}
