package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// chatMessage represents a single message in the chat history.
type chatMessage struct {
	ts       time.Time
	senderID uint32 // 0 = self
	name     string
	body     string
	isAction bool
}

// chatPanel manages the chat viewport and input box.
type chatPanel struct {
	friendID    uint32
	friendName  string
	history     []chatMessage
	viewport    viewport.Model
	input       textinput.Model
	isTyping    bool // remote peer is typing
	focused     bool
	width       int
	height      int
	selfName    string
	initialized bool
}

func newChatPanel(width, height int) chatPanel {
	ti := textinput.New()
	ti.Placeholder = "type a message..."
	ti.CharLimit = 1000

	vp := viewport.New(width-4, height-8)

	return chatPanel{
		viewport: vp,
		input:    ti,
		width:    width,
		height:   height,
	}
}

func (c *chatPanel) setFriend(friendID uint32, name string) {
	if c.friendID == friendID {
		return
	}
	c.friendID = friendID
	c.friendName = name
	c.history = nil
	c.refreshViewport()
}

func (c *chatPanel) addMessage(msg chatMessage) {
	c.history = append(c.history, msg)
	c.refreshViewport()
	c.viewport.GotoBottom()
}

func (c *chatPanel) refreshViewport() {
	c.viewport.SetContent(c.renderHistory())
}

func (c *chatPanel) renderHistory() string {
	if len(c.history) == 0 {
		return messagePeer.Render("(no messages yet)")
	}

	var sb strings.Builder
	for _, m := range c.history {
		ts := messageTimestamp.Render(fmt.Sprintf("[%s]", m.ts.Format("15:04")))
		var line string
		if m.isAction {
			line = fmt.Sprintf("%s * %s %s", ts, m.name, messageAction.Render(m.body))
		} else if m.senderID == 0 {
			line = fmt.Sprintf("%s %s: %s", ts, messageSelf.Render("You"), m.body)
		} else {
			line = fmt.Sprintf("%s %s: %s", ts, messagePeer.Render(m.name), m.body)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if c.isTyping {
		sb.WriteString(typingIndicator.Render(fmt.Sprintf("%s is typing...", c.friendName)))
		sb.WriteString("\n")
	}

	return sb.String()
}

// resize adjusts viewport and input dimensions.
func (c *chatPanel) resize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Width = width - 4
	c.viewport.Height = height - 8
	c.input.Width = width - 8
	c.refreshViewport()
}

// update processes input messages; returns (message to send, was submitted).
func (c *chatPanel) update(msg tea.Msg) (string, bool) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if !c.focused {
			return "", false
		}
		switch m.String() {
		case "enter":
			val := strings.TrimSpace(c.input.Value())
			if val != "" {
				c.input.SetValue("")
				return val, true
			}
		default:
			var cmd tea.Cmd
			c.input, cmd = c.input.Update(msg)
			_ = cmd
			c.viewport, cmd = c.viewport.Update(msg)
			_ = cmd
		}
	case tea.MouseMsg:
		var cmd tea.Cmd
		c.viewport, cmd = c.viewport.Update(msg)
		_ = cmd
	}
	return "", false
}

// view renders the chat panel.
func (c chatPanel) view() string {
	var header string
	if c.friendName != "" {
		header = chatHeaderStyle.Render(fmt.Sprintf("Chat with: %s", c.friendName))
	} else {
		header = chatHeaderStyle.Render("(select a contact)")
	}

	divider := strings.Repeat("─", max(0, c.width-4))
	inputLine := fmt.Sprintf("> %s", c.input.View())

	content := strings.Join([]string{
		header,
		divider,
		c.viewport.View(),
		divider,
		inputLine,
	}, "\n")

	innerWidth := c.width - 2
	style := inactivePanel
	if c.focused {
		style = activePanel
	}

	return style.Width(innerWidth).Height(c.height - 2).Render(content)
}

// focus sets the focus state and enables or disables the text input.
func (c *chatPanel) focus() {
	c.focused = true
	c.input.Focus()
}

func (c *chatPanel) blur() {
	c.focused = false
	c.input.Blur()
}
