package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// task (items for list)

// app model
type model struct {
	list      list.Model
	width     int
	height    int
	view      string
	questions []string
	input     textinput.Model
	counter   int
	tempTask  Task
}

func NewList(width, height int) (list.Model, error) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	file, err := os.Open("./tasks.json")
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()

	var tasks []JsonTask

	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		return l, err
	}

	for i, t := range tasks {
		var tempTask Task
		tempTask.title = t.Title
		tempTask.date = t.Date
		l.InsertItem(i, tempTask)
	}

	l.Title = "Tasks"

	return l, nil
}

func Tasks2Json(task []Task) []JsonTask {
	var jsontasks []JsonTask
	for _, t := range task {
		var jsontask JsonTask
		jsontask.Title = t.title
		jsontask.Date = t.date
		jsontasks = append(jsontasks, jsontask)
	}
	return jsontasks
}

func WriteTasks(tasks []JsonTask) {
	file, err := os.Create("./tasks.json")
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(tasks); err != nil {
		panic("encoding error: " + err.Error())
	}
	// Close the file
	if err := file.Close(); err != nil {
		panic(err.Error())
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func New() *model {
	m := model{view: "list", questions: []string{"task", "date"}, counter: 0}
	log.Print(len(m.questions))
	return &m
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		l, err := NewList(msg.Width, msg.Height)
		if err != nil {
			panic(err.Error())
		}
		m.list = l
		m.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if msg.String() == " " {
			if m.view == "list" {
				current_task := m.list.SelectedItem().(Task)
				if string(current_task.title[0]) == "*" {
					current_task.title = strings.Replace(current_task.title, "*", "x", 1)
				} else if string(current_task.title[0]) == "x" {
					current_task.title = strings.Replace(current_task.title, "x", "*", 1)
				}
				m.list.SetItem(m.list.Cursor(), current_task)
			}
		}
		if msg.String() == "a" {
			if m.view == "list" {
				m.view = "add"
				m.input = textinput.New()
				m.input.Focus()
				log.Print(m.counter)
			}
		}
		if msg.String() == "enter" {
			if m.view == "add" {
				if m.counter < len(m.questions) {
					// Set task properties based on counter
					if m.tempTask.title == "" {
						m.tempTask.title = "* " + m.input.Value()
						m.counter++
					} else {
						m.tempTask.date = m.input.Value()
						// Insert task into list and switch back to "list" view
						m.list.InsertItem(0, m.tempTask)
						m.view = "list"
						m.tempTask = Task{}
						m.counter = 0 // Reset counter
						items := m.list.Items()
						var tasks []Task
						for _, i := range items {
							t := i.(Task)
							tasks = append(tasks, t)
						}
						jsonTasks := Tasks2Json(tasks)
						WriteTasks(jsonTasks)
					}
					m.input.Reset()
				}
			}
		}
	}
	var cmd tea.Cmd
	if m.view == "add" {
		m.input, cmd = m.input.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m *model) View() string {
	if m.view == "list" {
		return m.list.View()
	} else if m.view == "add" {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, m.questions[m.counter], m.input.View()))
	} else {
		return m.list.View()
	}
}

func main() {
	m := New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	p.Run()
}
