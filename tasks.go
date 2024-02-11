package main

type Task struct {
	title string
	date  string
}
type JsonTask struct {
	Title string `json:"title"`
	Date  string `json:"date"`
}

func (t Task) FilterValue() string {
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.date
}
