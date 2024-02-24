package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
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

func SortJsonTasks(tasks []JsonTask) []JsonTask {
	sort.Slice(tasks, func(i, j int) bool {
		date1, _ := time.Parse("02-01-2006", tasks[i].Date) // Parse date string into time.Time object
		date2, _ := time.Parse("02-01-2006", tasks[j].Date) // Parse date string into time.Time object
		return date1.Before(date2)
	})
	for i := range tasks {
		tasks[i].Index = i
	}
	// panic(tasks)
	return tasks
}

var inputStyles = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).Width(50).Height(1)
var listStyles = lipgloss.NewStyle()

func NewList(width, height int) (list.Model, error) {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	usrhd, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	file, err := os.Open(usrhd + "/todotea/tasks.json")
	if err != nil {
		panic(err.Error())
	}

	defer file.Close()

	var tasks []JsonTask

	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		return l, err
	}
	tasks = SortJsonTasks(tasks)
	for i, t := range tasks {
		var tempTask Task
		tempTask.title = t.Title
		tempTask.date = t.Date
		tempTask.index = t.Index
		l.InsertItem(i, tempTask)
	}

	l.Title = "Tasks"
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("ctrl+t"),
				key.WithHelp("ctrl+t", "add item"),
			),
			key.NewBinding(
				key.WithKeys("backspace/d"),
				key.WithHelp("backspace/d", "delete item"),
			),
			key.NewBinding(
				key.WithKeys("space"),
				key.WithHelp("space", "set item status"),
			),
			key.NewBinding(
				key.WithKeys("ctrl+e"),
				key.WithHelp("ctrl+e", "edit item"),
			),
		}
	}

	return l, nil
}

func Tasks2Json(task []Task) []JsonTask {
	var jsontasks []JsonTask
	for _, t := range task {
		var jsontask JsonTask
		jsontask.Title = t.title
		jsontask.Date = t.date
		jsontask.Index = t.index
		jsontasks = append(jsontasks, jsontask)
	}
	return jsontasks
}

func WriteTasks(tasks []JsonTask) {
	usrhd, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	file, err := os.Create(usrhd + "/todotea/tasks.json")
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
	return &m
}

func getTasksFromItems(items []list.Item) []Task {
	var tasks []Task
	for _, i := range items {
		t := i.(Task)
		tasks = append(tasks, t)
	}
	return tasks
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
				items := m.list.Items()
				var tasks []Task = getTasksFromItems(items)
				jsonTasks := Tasks2Json(tasks)
				WriteTasks(jsonTasks)
			}
		}
		if msg.String() == "ctrl+t" {
			if m.view == "list" {
				m.view = "add"
				m.input = textinput.New()
				m.input.Focus()
			}
		}
		if msg.String() == "ctrl+q" {
			if m.view == "add" {
				m.view = "list"
			} else {
				tea.Quit()
			}
		}
		if msg.String() == "d" || msg.String() == "backspace" {
			if m.view == "list" {
				selectedItem := m.list.SelectedItem().(Task)
				indexToRemove := selectedItem.index
				m.list.RemoveItem(indexToRemove)

				// Update indexes of tasks after removal
				items := m.list.Items()
				for i, item := range items {
					task := item.(Task)
					task.index = i
					items[i] = task
				}

				// Convert tasks to JSON and write to file
				var tasks []Task = getTasksFromItems(items)
				jsonTasks := Tasks2Json(tasks)
				jsonTasks = SortJsonTasks(jsonTasks)
				WriteTasks(jsonTasks)
			}
		}
		if msg.String() == "enter" {
			if m.view == "add" {
				if m.input.Value() != "" && m.tempTask.title == "" {
					m.tempTask.title = "* " + m.input.Value()
					m.counter++
					m.input.Reset()
				} else if m.input.Value() != "" {
					if strings.ToLower(m.input.Value()) == "today" {
						today := time.Now().Format("02-01-2006")
						m.tempTask.date = today
					} else if strings.ToLower(m.input.Value()) == "tomorrow" {
						now := time.Now()
						tomorrow := now.Add(24 * time.Hour)
						m.tempTask.date = tomorrow.Format("02-01-2006")
					} else {
						m.tempTask.date = m.input.Value()
					}

					// Insert task into list and switch back to "list" view
					m.view = "list"
					items := m.list.Items()
					var tasks []Task = getTasksFromItems(items)
					var index int
					m.tempTask.index = index
					tasks = append(tasks, m.tempTask)
					m.tempTask = Task{}
					m.counter = 0
					jsonTasks := Tasks2Json(tasks)
					jsonTasks = SortJsonTasks(jsonTasks)
					WriteTasks(jsonTasks)
					m.input.Reset()
					l, err := NewList(m.width, m.height)
					if err != nil {
						panic(err.Error())
					}
					m.list = l
				}
			} else if m.view == "edit" {
				if m.input.Value() != "" && m.tempTask.title == "" {
					m.tempTask.title = m.input.Value()
					m.counter++
					m.input.SetValue(m.list.SelectedItem().(Task).Description())
				} else if m.input.Value() != "" {
					if strings.ToLower(m.input.Value()) == "today" {
						today := time.Now().Format("02-01-2006")
						m.tempTask.date = today
						log.Print(m.tempTask.date)
					} else if strings.ToLower(m.input.Value()) == "tomorrow" {
						now := time.Now()
						tomorrow := now.Add(24 * time.Hour)
						m.tempTask.date = tomorrow.Format("02-01-2006")
					} else {
						m.tempTask.date = m.input.Value()
					}

					// Insert task into list and switch back to "list" view
					m.list.RemoveItem(m.list.Cursor())
					m.list.InsertItem(0, m.tempTask)

					m.view = "list"
					m.tempTask = Task{}
					m.counter = 0 // Reset counter
					items := m.list.Items()
					var tasks []Task = getTasksFromItems(items)
					jsonTasks := Tasks2Json(tasks)
					jsonTasks = SortJsonTasks(jsonTasks)
					WriteTasks(jsonTasks)
					m.input.Reset()
					l, err := NewList(m.width, m.height)
					if err != nil {
						panic(err.Error())
					}
					m.list = l
				}
			}
		}
		if msg.String() == "ctrl+e" {
			if m.view == "list" {
				m.view = "edit"
				m.input = textinput.New()
				m.input.Focus()
				m.input.SetValue(m.list.SelectedItem().(Task).Title())
			}
		}

	}
	var cmd tea.Cmd
	if m.view == "add" || m.view == "edit" {
		m.input, cmd = m.input.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m *model) View() string {
	if m.view == "list" {
		return m.list.View()
	} else if m.view == "add" || m.view == "edit" {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, m.questions[m.counter], inputStyles.Render(m.input.View())))
	} else {
		return lipgloss.JoinVertical(lipgloss.Center, m.list.View())
	}
}

func main() {
	m := New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	p.Run()
}
