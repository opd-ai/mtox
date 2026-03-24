package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	toxclient "github.com/opd-ai/mtox/internal/tox"
)

// focus panel constants.
const (
	focusContacts = iota
	focusChat
)

// modalKind identifies which modal dialog is open.
type modalKind int

const (
	modalNone modalKind = iota
	modalAddFriend
	modalFriendRequest
)

// pendingRequest holds an unresolved incoming friend request.
type pendingRequest struct {
	PublicKey [32]byte
	Message   string
}

// App is the root bubbletea model.
type App struct {
	client    *toxclient.Client
	contacts  contactsPanel
	chat      chatPanel
	statusBar statusBar
	focus     int
	width     int
	height    int

	// modal state
	modal        modalKind
	modalInput   textinput.Model
	modalPrompt  string
	pendingReqs  []pendingRequest
	activeReqIdx int

	// current conversation
	activeFriendID uint32
	activeFriend   string

	// error/info message
	notification string
	notifyExpiry time.Time

	ready bool
}

// New creates the root App model.
func New(client *toxclient.Client) App {
	mi := textinput.New()
	mi.CharLimit = 512

	a := App{
		client:     client,
		modalInput: mi,
	}
	return a
}

// Init implements tea.Model.
func (a App) Init() tea.Cmd {
	return tea.Batch(
		waitForToxEvent(a.client.Events()),
		tickCmd(),
	)
}

// waitForToxEvent returns a tea.Cmd that blocks until a ToxEvent is available.
func waitForToxEvent(events <-chan toxclient.ToxEvent) tea.Cmd {
	return func() tea.Msg {
		return <-events
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return toxclient.TickEvent{}
	})
}

// Update implements tea.Model.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {

	case tea.WindowSizeMsg:
		return a.handleResize(m)

	case tea.KeyMsg:
		return a.handleKey(m)

	case tea.MouseMsg:
		return a.handleMouse(m)

	// Tox events
	case toxclient.FriendRequestEvent:
		a.pendingReqs = append(a.pendingReqs, pendingRequest{
			PublicKey: m.PublicKey,
			Message:   m.Message,
		})
		if a.modal == modalNone {
			a.openFriendRequestModal()
		}
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendMessageEvent:
		a.handleIncomingMessage(m.FriendID, m.Message, false)
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendConnectionStatusEvent:
		a.contacts.updateConnectionStatus(m.FriendID, m.Status)
		a.refreshContacts()
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendNameEvent:
		a.contacts.updateName(m.FriendID, m.Name)
		if m.FriendID == a.activeFriendID {
			a.activeFriend = m.Name
			a.chat.friendName = m.Name
		}
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendStatusMessageEvent:
		a.refreshContacts()
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendTypingEvent:
		if m.FriendID == a.activeFriendID {
			a.chat.isTyping = m.IsTyping
			a.chat.refreshViewport()
		}
		return a, waitForToxEvent(a.client.Events())

	case toxclient.SelfConnectionStatusEvent:
		a.statusBar.connectionStatus = m.Status
		return a, waitForToxEvent(a.client.Events())

	case toxclient.AnonymityStatusEvent:
		switch m.Network {
		case "tor":
			a.statusBar.torStatus = m.Status
			if m.Status == toxclient.AnonymityAvailable && m.Address != "" {
				a.setNotification("Tor hidden service ready")
			}
		case "i2p":
			a.statusBar.i2pStatus = m.Status
			if m.Status == toxclient.AnonymityAvailable && m.Address != "" {
				a.setNotification("I2P destination ready")
			}
		}
		return a, waitForToxEvent(a.client.Events())

	case toxclient.TickEvent:
		a.refreshContacts()
		// Clear expired notifications.
		if !a.notifyExpiry.IsZero() && time.Now().After(a.notifyExpiry) {
			a.notification = ""
		}
		return a, tickCmd()
	}

	return a, nil
}

// handleResize processes window size changes.
func (a App) handleResize(m tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	a.width = m.Width
	a.height = m.Height

	statusH := 1
	mainH := a.height - statusH

	contactsW := 22
	chatW := a.width - contactsW

	if !a.ready {
		a.contacts = newContactsPanel(contactsW, mainH)
		a.chat = newChatPanel(chatW, mainH)
		a.statusBar = newStatusBar(a.width)
		a.statusBar.selfAddress = a.client.SelfAddress()
		a.contacts.focused = true
		a.focus = focusContacts
		a.ready = true

		a.refreshContacts()
		a.client.Bootstrap()
	} else {
		a.contacts.width = contactsW
		a.contacts.height = mainH
		a.chat.resize(chatW, mainH)
		a.statusBar.width = a.width
	}

	return a, nil
}

// handleKey processes keyboard input.
func (a App) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global shortcuts.
	switch m.String() {
	case "ctrl+c", "ctrl+q":
		return a.quit()
	case "ctrl+s":
		if err := a.client.Save(); err != nil {
			a.setNotification(fmt.Sprintf("Save failed: %v", err))
		} else {
			a.setNotification("Profile saved.")
		}
		return a, nil
	case "ctrl+n":
		if a.modal == modalNone {
			a.openAddFriendModal()
		}
		return a, nil
	case "ctrl+g":
		a.setNotification("Group chat not yet supported.")
		return a, nil
	case "tab":
		a.toggleFocus()
		return a, nil
	case "esc":
		if a.modal != modalNone {
			a.closeModal()
		}
		return a, nil
	case "enter":
		if a.modal == modalAddFriend {
			return a.submitAddFriend()
		}
		if a.modal == modalFriendRequest {
			return a.acceptFriendRequest()
		}
	case "r":
		if a.modal == modalFriendRequest {
			return a.rejectFriendRequest()
		}
	}

	// Modal input.
	if a.modal == modalAddFriend {
		var cmd tea.Cmd
		a.modalInput, cmd = a.modalInput.Update(m)
		return a, cmd
	}

	// Panel input.
	if a.focus == focusContacts {
		selected, friendID := a.contacts.update(m)
		if selected {
			a.selectFriend(friendID)
		}
	} else {
		text, submitted := a.chat.update(m)
		if submitted {
			return a.sendMessage(text)
		}
		// Notify typing status.
		if a.activeFriendID != 0 {
			isTyping := len(a.chat.input.Value()) > 0
			_ = a.client.SetTyping(a.activeFriendID, isTyping)
		}
	}

	return a, nil
}

// handleMouse processes mouse events.
func (a App) handleMouse(m tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Determine if click is in contacts or chat area.
	contactsW := 22
	if m.X < contactsW {
		selected, friendID := a.contacts.update(m)
		if selected {
			a.focus = focusContacts
			a.contacts.focused = true
			a.chat.blur()
			a.selectFriend(friendID)
		}
	} else {
		if m.Action == tea.MouseActionPress {
			a.focus = focusChat
			a.chat.focus()
			a.contacts.focused = false
		}
		a.chat.update(m)
	}
	return a, nil
}

// toggleFocus switches focus between the two panels.
func (a *App) toggleFocus() {
	if a.focus == focusContacts {
		a.focus = focusChat
		a.contacts.focused = false
		a.chat.focus()
	} else {
		a.focus = focusContacts
		a.contacts.focused = true
		a.chat.blur()
	}
}

// selectFriend switches the active conversation.
func (a *App) selectFriend(friendID uint32) {
	if a.activeFriendID == friendID {
		return
	}
	a.activeFriendID = friendID
	a.activeFriend = a.friendName(friendID)
	a.chat.setFriend(friendID, a.activeFriend)
	a.contacts.clearUnread(friendID)
}

// friendName resolves a friend's display name from the contacts panel.
func (a *App) friendName(friendID uint32) string {
	for _, c := range a.contacts.contacts {
		if c.FriendID == friendID {
			return c.Name
		}
	}
	return fmt.Sprintf("Friend %d", friendID)
}

// refreshContacts syncs the contacts panel from the tox client.
func (a *App) refreshContacts() {
	friends := a.client.GetFriends()
	a.contacts.setContacts(friends)
}

// handleIncomingMessage appends a received message to the chat history.
func (a *App) handleIncomingMessage(friendID uint32, text string, isAction bool) {
	name := a.friendName(friendID)
	msg := chatMessage{
		ts:       time.Now(),
		senderID: friendID,
		name:     name,
		body:     text,
		isAction: isAction,
	}
	if friendID == a.activeFriendID {
		a.chat.addMessage(msg)
	} else {
		a.contacts.incrementUnread(friendID)
	}
}

// sendMessage sends the current input as a message to the active friend.
func (a App) sendMessage(text string) (tea.Model, tea.Cmd) {
	if a.activeFriendID == 0 {
		return a, nil
	}
	if err := a.client.SendMessage(a.activeFriendID, text); err != nil {
		a.setNotification(fmt.Sprintf("Send failed: %v", err))
		return a, nil
	}
	_ = a.client.SetTyping(a.activeFriendID, false)
	msg := chatMessage{
		ts:       time.Now(),
		senderID: 0,
		name:     "You",
		body:     text,
	}
	a.chat.addMessage(msg)
	return a, nil
}

// quit saves and exits.
func (a App) quit() (tea.Model, tea.Cmd) {
	_ = a.client.Save()
	a.client.Stop()
	return a, tea.Quit
}

// openAddFriendModal shows the Add Friend dialog.
func (a *App) openAddFriendModal() {
	a.modal = modalAddFriend
	a.modalPrompt = "Enter Tox ID (then press Enter):"
	a.modalInput.SetValue("")
	a.modalInput.Placeholder = "Tox ID..."
	a.modalInput.Focus()
}

// openFriendRequestModal shows the incoming friend request dialog.
func (a *App) openFriendRequestModal() {
	if len(a.pendingReqs) == 0 {
		return
	}
	a.modal = modalFriendRequest
	a.activeReqIdx = 0
}

// closeModal dismisses the current modal.
func (a *App) closeModal() {
	a.modal = modalNone
	a.modalInput.Blur()
}

// submitAddFriend processes the Add Friend form submission.
func (a App) submitAddFriend() (tea.Model, tea.Cmd) {
	addr := strings.TrimSpace(a.modalInput.Value())
	a.closeModal()
	if addr == "" {
		return a, nil
	}
	if _, err := a.client.AddFriend(addr, "Hi! I'd like to add you on mtox."); err != nil {
		a.setNotification(fmt.Sprintf("Add friend failed: %v", err))
	} else {
		a.setNotification("Friend request sent.")
		a.refreshContacts()
	}
	return a, nil
}

// acceptFriendRequest accepts the current pending request.
func (a App) acceptFriendRequest() (tea.Model, tea.Cmd) {
	if len(a.pendingReqs) == 0 {
		a.closeModal()
		return a, nil
	}
	req := a.pendingReqs[a.activeReqIdx]
	if _, err := a.client.AcceptFriend(req.PublicKey); err != nil {
		a.setNotification(fmt.Sprintf("Accept failed: %v", err))
	} else {
		a.setNotification("Friend request accepted.")
		a.refreshContacts()
	}
	a.pendingReqs = append(a.pendingReqs[:a.activeReqIdx], a.pendingReqs[a.activeReqIdx+1:]...)
	if len(a.pendingReqs) == 0 {
		a.closeModal()
	}
	return a, nil
}

// rejectFriendRequest discards the current pending request.
func (a App) rejectFriendRequest() (tea.Model, tea.Cmd) {
	if len(a.pendingReqs) == 0 {
		a.closeModal()
		return a, nil
	}
	a.pendingReqs = append(a.pendingReqs[:a.activeReqIdx], a.pendingReqs[a.activeReqIdx+1:]...)
	if len(a.pendingReqs) == 0 {
		a.closeModal()
	}
	return a, nil
}

func (a *App) setNotification(msg string) {
	a.notification = msg
	a.notifyExpiry = time.Now().Add(4 * time.Second)
}

// View implements tea.Model.
func (a App) View() string {
	if !a.ready {
		return "Initializing mtox...\n"
	}

	contactsView := a.contacts.view()
	chatView := a.chat.view()

	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, contactsView, chatView)

	statusView := a.statusBar.view()

	var view string
	if a.notification != "" {
		notifStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("237")).
			Width(a.width).
			PaddingLeft(1)
		view = lipgloss.JoinVertical(lipgloss.Left,
			mainRow,
			notifStyle.Render(a.notification),
		)
	} else {
		view = lipgloss.JoinVertical(lipgloss.Left, mainRow, statusView)
	}

	// Overlay modal if active.
	if a.modal != modalNone {
		view = a.overlayModal(view)
	}

	return view
}

// overlayModal renders a modal dialog over the current view.
func (a App) overlayModal(base string) string {
	var content string
	switch a.modal {
	case modalAddFriend:
		content = fmt.Sprintf("%s\n\n%s\n\n%s",
			panelTitle.Render("Add Friend"),
			a.modalPrompt,
			a.modalInput.View(),
		)
	case modalFriendRequest:
		if len(a.pendingReqs) == 0 {
			break
		}
		req := a.pendingReqs[a.activeReqIdx]
		pkHex := fmt.Sprintf("%x", req.PublicKey)
		if len(pkHex) > 20 {
			pkHex = pkHex[:20] + "..."
		}
		content = fmt.Sprintf("%s\n\nFrom: %s\nMessage: %s\n\n[Enter] Accept  [R] Reject  [Esc] Dismiss",
			panelTitle.Render("Friend Request"),
			pkHex,
			req.Message,
		)
	}

	if content == "" {
		return base
	}

	dialog := modalStyle.Render(content)
	dialogW := lipgloss.Width(dialog)
	dialogH := lipgloss.Height(dialog)

	x := (a.width - dialogW) / 2
	y := (a.height - dialogH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	lines := strings.Split(base, "\n")
	dialogLines := strings.Split(dialog, "\n")

	for i, dl := range dialogLines {
		row := y + i
		if row >= len(lines) {
			lines = append(lines, "")
		}
		line := lines[row]
		// Pad line to at least x characters.
		for len([]rune(line)) < x {
			line += " "
		}
		// Replace characters at position x with dialog line.
		runes := []rune(line)
		dlRunes := []rune(dl)
		end := x + len(dlRunes)
		for len(runes) < end {
			runes = append(runes, ' ')
		}
		copy(runes[x:], dlRunes)
		lines[row] = string(runes)
	}

	return strings.Join(lines, "\n")
}
