package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"errors"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/crypto/ssh"
)


var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
	history = loadHistory()

)


type model struct {
	focusIndex     int
	inputs         []textinput.Model
	cursorMode     cursor.Mode
	suggestions    []string
	showSuggestion string
}


type Inputs struct {
	FilePath   string
	TargetPath string
	FileMode      int
	Mode          int
	Username   string
	Address    string
	Password   string
}


func ConnectClient(username, password, address string) scp.Client {
	clientConfig, _ := auth.PasswordKey(username, password, ssh.InsecureIgnoreHostKey())
	client := scp.NewClient(fmt.Sprintf("%s:22", address), &clientConfig)
	return client

}


type History struct {
	Paths     []string `json:"paths"`
	Targets   []string `json:"targets"`
	Users     []string `json:"users"`
	Addresses []string `json:"addresses"`
}


func loadHistory() History {
	file, err := os.ReadFile("history.json")
	if err != nil {
		return History{}
	}
	var hist History
	err = json.Unmarshal(file, &hist)
	if err != nil {
		return History{}
	}
	return hist
}


func saveHistory(inputs Inputs, h *History) error {
	addUnique := func(slice []string, val string) []string {
		for _, v := range slice {
			if v == val {
				return slice
			}
		}
		return append(slice, val)
	}

	h.Paths = addUnique(h.Paths, inputs.FilePath)
	h.Targets = addUnique(h.Targets, inputs.TargetPath)
	h.Users = addUnique(h.Users, inputs.Username)
	h.Addresses = addUnique(h.Addresses, inputs.Address)

	data, _ := json.MarshalIndent(h, "", "  ")
	err := os.WriteFile("history.json", data, 0644)
	return err
}


func ZipFolder(sourceDir, outputZip string) error {
	zipFile, err := os.Create(outputZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == sourceDir {
			return nil
		}

		relPath := strings.TrimPrefix(path, filepath.Clean(sourceDir)+string(os.PathSeparator))

		if info.IsDir() {
			_, err = zipWriter.Create(relPath + "/")
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})
}


func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, 7),
	}
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.Width = 30
	
		switch i {
		case 0:
			t.Placeholder = "FilePath"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "TargetPath"
		case 2:
			t.Placeholder = "FileMode: 0=file, 1=folder"
		case 3:
			t.Placeholder = "Mode: 0=upload, 1=download"
		case 4:
			t.Placeholder = "Username"
			t.CharLimit = 64
		case 5:
			t.Placeholder = "Address"
			t.CharLimit = 64
		case 6:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
	
		m.inputs[i] = t
	}
	switch m.focusIndex {
	case 0:
			m.suggestions = history.Paths
	case 1:
			m.suggestions = history.Targets
	case 4:
			m.suggestions = history.Users
	case 5:
			m.suggestions = history.Addresses
	}
	return m
}


func (m model) Init() tea.Cmd {
	return textinput.Blink
}


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		case "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)
	
	// Only try to show suggestions if we're on a valid input index
	if m.focusIndex >= 0 && m.focusIndex < len(m.inputs) {
		inputVal := m.inputs[m.focusIndex].Value()
		m.showSuggestion = ""

		// Update suggestions based on focus index
		switch m.focusIndex {
		case 0:
			m.suggestions = history.Paths
		case 1:
			m.suggestions = history.Targets
		case 4:
			m.suggestions = history.Users
		case 5:
			m.suggestions = history.Addresses
		default:
			m.suggestions = nil
		}

		// Find a matching suggestion
		for _, s := range m.suggestions {
			if strings.HasPrefix(s, inputVal) && s != inputVal {
				m.showSuggestion = s
				break
			}
		}

		// Handle tab completion
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "right", "tab":
				if m.showSuggestion != "" {
					m.inputs[m.focusIndex].SetValue(m.showSuggestion)
					m.showSuggestion = ""
				}
			}
		}
	}

	return m, cmd
}


func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i == m.focusIndex && m.showSuggestion != "" {
			b.WriteString(helpStyle.Render(fmt.Sprintf("\n↪ %s\n", m.showSuggestion)))
		}
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}

func ParseModeInt(mode,filemode string) (int, int, error){
	modeInt, err := strconv.Atoi(mode)
	if err != nil {
		return 0,0, errors.New("Invalid Mode. Skill Issues")
	

	}
	if modeInt != 1 && modeInt != 0 {
		return 0,0, errors.New("Invalid Mode. Skill Issues")
	}
	filemodeInt, err := strconv.Atoi(filemode)
	if err != nil {
		return 0,0, errors.New("Invalid Mode. Skill Issues")

	}
	if filemodeInt != 1 && filemodeInt != 0 {
		return 0,0, errors.New("Invalid Mode. Skill Issues")

	}
	return modeInt,filemodeInt,nil
	
}

func DownloadFile(client scp.Client, remotePath, localPath string) error {
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	err = client.CopyFromRemote(context.Background(), file, remotePath)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	return nil
}


func main() {
	cwd, _ := os.Getwd()
	p := tea.NewProgram(initialModel())
	// fmodel, err := p.Run()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	os.Exit(1)
	// }
	// finalModel := fmodel.(model)
	fmodel, err := p.Run()
	if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
	}
	
	if fmodel == nil {
			fmt.Println("No input submitted. Exiting.")
			os.Exit(0)
	}
	
	finalModel, ok := fmodel.(model)
	if !ok {
			fmt.Println("Unexpected error: invalid model type.")
			os.Exit(1)
	}
	
	filemodeInt, modeInt, err  := ParseModeInt(finalModel.inputs[2].Value(),finalModel.inputs[3].Value())
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	inputs := Inputs{
		FilePath:   finalModel.inputs[0].Value(),
		TargetPath: finalModel.inputs[1].Value(),
		Mode:       modeInt,
		FileMode:       filemodeInt,
		Username:   finalModel.inputs[4].Value(),
		Address:    finalModel.inputs[5].Value(),
		Password:   finalModel.inputs[6].Value(),
	}
	client := ConnectClient(inputs.Username, inputs.Password, inputs.Address)
	err = client.Connect()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer client.Close()

	if inputs.Mode== 0{
		
		if inputs.FileMode == 0 {
			parts := strings.Split(inputs.FilePath, "/")
			filename := parts[len(parts)-1]
			file, err := os.Open(inputs.FilePath)
			if err != nil {
				fmt.Println("Failed to open file:", cwd,err)
				return
			}
			defer file.Close()
			remotePath := fmt.Sprintf("%s/%s", strings.TrimRight(inputs.TargetPath, "/"), filename)
			err = client.CopyFile(context.Background(), file, remotePath, "0655")
			if err != nil {
				fmt.Println("Failed to upload file:", err)
			} else {
				fmt.Println("Upload complete!")
			}
			
		} else if inputs.FileMode == 1 {
			parts := strings.Split(inputs.FilePath, "/")
			foldername := parts[len(parts)-1]
			zipfileName := fmt.Sprintf("%s.zip", foldername)
			err := ZipFolder(inputs.FilePath, zipfileName)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			file, err := os.Open(zipfileName)
			if err != nil {
				fmt.Println("Failed to open file:", cwd, err)
				return
			}
			defer file.Close()
			remotePath := fmt.Sprintf("%s/%s", strings.TrimRight(inputs.TargetPath, "/"), zipfileName)
			err = client.CopyFile(context.Background(), file, remotePath, "0655")
			if err != nil {
				fmt.Println("Failed to upload file:", err)
			} else {
				fmt.Println("Upload complete!")
			}
		}
		
	}else if inputs.Mode == 1{
		if inputs.FileMode == 0 {
			parts := strings.Split(inputs.FilePath, "/")
			filename := parts[len(parts)-1]
			localPath := filename
			remotePath := fmt.Sprintf("%s/%s", strings.TrimRight(inputs.TargetPath, "/"), filename)

			err = DownloadFile(client, remotePath, localPath)
			if err != nil {
				fmt.Println("Download failed:",cwd, err)
			} else {
				fmt.Println("Download complete:", cwd, localPath)
			}
		}
		
		
	}
	
	if err := saveHistory(inputs, &history); err != nil {
		fmt.Println("Warning: Failed to save history:", err)
	}
}
