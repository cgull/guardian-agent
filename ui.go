package guardianagent

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/howeyc/gopass"
	i "github.com/sternhenri/interact"
)

type UI interface {
	Ask(prompt Prompt) (int, error)
	Confirm(msg string) bool
	Inform(msg string)
	Alert(msg string)
	AskPassword(msg string) ([]byte, error)
}

type TerminalUI struct{}
type FancyTerminalUI struct{}
type AskPassUI struct{}

type Prompt struct {
	Question string
	Choices  []string
}

func (TerminalUI) Ask(params Prompt) (reply int, err error) {
	reply = -1
	var convErr error

	for convErr != nil || reply <= 0 || reply > len(params.Choices) {
		fmt.Print(formatPrompt(params))
		reader := bufio.NewReader(os.Stdin)
		sReply, _ := reader.ReadString('\n')
		reply, convErr = strconv.Atoi(strings.TrimSpace(sReply))
	}

	return
}

func (tui TerminalUI) Confirm(msg string) bool {
	prompt := Prompt{Question: msg, Choices: []string{"Yes", "No"}}
	ans, err := tui.Ask(prompt)
	return err == nil && ans == 1
}

func (TerminalUI) Inform(msg string) {
	fmt.Println(msg)
}

func (TerminalUI) Alert(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}

func (TerminalUI) AskPassword(msg string) ([]byte, error) {
	fmt.Println(msg)
	return gopass.GetPasswd()
}

func formatPrompt(params Prompt) (formattedPrompt string) {
	var buf bytes.Buffer
	buf.WriteString(params.Question)
	for i, v := range params.Choices {
		buf.WriteString(fmt.Sprintf("\n    %d) %s", i+1, v))
	}
	buf.WriteString("\n\nAnswer (enter a number): ")
	formattedPrompt = buf.String()
	return
}

func mapToChoice(vs []string) []i.Choice {
	vsm := make([]i.Choice, len(vs))
	for j, v := range vs {
		vsm[j] = i.Choice{Text: v}
	}
	return vsm
}

func (FancyTerminalUI) Ask(params Prompt) (reply int, err error) {
	var resp int64

	i.Run(&i.Interact{
		Questions: []*i.Question{
			{
				Quest: i.Quest{
					Msg: params.Question,
					Choices: i.Choices{
						Alternatives: mapToChoice(params.Choices),
					},
				},
				Action: func(c i.Context) interface{} {
					resp, _ = c.Ans().Int()
					return nil
				},
			},
		},
	})
	reply = int(resp)
	return
}

func (FancyTerminalUI) Inform(msg string) {
	fmt.Println(msg)
}

func (FancyTerminalUI) Alert(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}

func (FancyTerminalUI) AskPassword(msg string) ([]byte, error) {
	fmt.Println(msg)
	return gopass.GetPasswd()
}

func (tui FancyTerminalUI) Confirm(msg string) bool {
	prompt := Prompt{Question: msg, Choices: []string{"Yes", "No"}}
	ans, err := tui.Ask(prompt)
	return err == nil && ans == 1
}

func (AskPassUI) Ask(params Prompt) (reply int, err error) {
	reply = -1
	var convErr error

	for convErr != nil || reply <= 0 || reply > len(params.Choices) { // 1 indexed
		cmd := exec.Command("ssh-askpass", formatPrompt(params))
		out, err := cmd.Output()
		if err != nil {
			return reply, err
		}
		sReply := strings.TrimSpace(string(out))
		reply, convErr = strconv.Atoi(sReply)
	}

	return
}

func (AskPassUI) Inform(msg string) {
	fmt.Println(msg)
}

func (AskPassUI) Alert(msg string) {
	cmd := exec.Command("ssh-askpass", msg)
	cmd.Run()
}

func (AskPassUI) AskPassword(msg string) ([]byte, error) {
	cmd := exec.Command("ssh-askpass", msg)
	out, err := cmd.Output()
	if err != nil {
		return out, err
	}
	return []byte(strings.TrimSpace(string(out))), nil
}

func (apui AskPassUI) Confirm(msg string) bool {
	cmd := exec.Command("ssh-askpass", msg)
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	outStr := strings.ToLower(strings.TrimSpace(string(out)))
	return len(outStr) == 0 || outStr == "yes"
}
