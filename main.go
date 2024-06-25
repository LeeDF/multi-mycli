package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cast"
	"log"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

var configFile = flag.String("c", "./conf.json", "config file path")

func main() {
	flag.Parse()
	m := loadConf()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func loadConf() model {
	all, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("read file err: %s", err.Error())
	}
	var m model
	err = json.Unmarshal(all, &m)
	if err != nil {
		log.Fatalf("unmarshal err: %s", err.Error())
	}
	if m.Cmd == "" {
		m.Cmd = "mysql"
	}
	if m.Cmd != "mysql" && m.Cmd != "mycli" {
		log.Fatalf("cmd must be mysql or mycli")
	}

	if len(m.Mysql) == 0 {
		log.Fatalf("mysql conf is empty")
	}
	return m
}

type MysqlConf struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pwd      string `json:"pwd"`
	Database string `json:"database"`
}

type mysqlFinishedMsg struct{ err error }

func runMysql(cm string, conf MysqlConf) tea.Cmd {
	var args []string
	if cm == "mycli" {
		args = []string{
			"-h", conf.Host, "-P", cast.ToString(conf.Port), "-u", conf.User, "-p", conf.Pwd, "-D", conf.Database,
		}
	}
	if cm == "mysql" {
		args = []string{
			"-h", conf.Host, "-P", cast.ToString(conf.Port), "-u", conf.User, "-p" + conf.Pwd, "-D", conf.Database,
		}
	}
	c := exec.Command(cm, args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return mysqlFinishedMsg{err}
	})
}

type model struct {
	err    error
	cursor int
	Cmd    string      `json:"cmd"`
	Mysql  []MysqlConf `json:"mysql"`
}

func (m model) Init() tea.Cmd {

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.err = nil
			return m, runMysql(m.Cmd, m.Mysql[m.cursor])
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.Mysql)-1 {
				m.cursor++
			}
		}
	case mysqlFinishedMsg:

		if msg.err != nil {
			m.err = msg.err
		}
	}
	return m, nil
}

func (m model) View() string {
	s := ""
	if m.err != nil {
		s += "Error: " + m.err.Error() + "\n"
	}
	s += "select database\n"

	// Iterate over our choices
	for i, conf := range m.Mysql {

		if i == m.cursor {
			s += color.GreenString("* %s/%s \n", conf.Host, conf.Database)
		} else {
			// Render the row
			s += fmt.Sprintf("%s/%s \n", conf.Host, conf.Database)
		}

	}

	// The footer
	s += "\nPress q to quit.\n"
	return s
}
