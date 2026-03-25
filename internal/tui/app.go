package tui

import (
	"fmt"
	"os"
	"path/filepath"
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

// contactsPanelWidth is the fixed width of the contacts panel.
const contactsPanelWidth = 22

// modalKind identifies which modal dialog is open.
type modalKind int

const (
	modalNone modalKind = iota
	modalAddFriend
	modalFriendRequest
	modalFileRequest
)

// pendingRequest holds an unresolved incoming friend request.
type pendingRequest struct {
	PublicKey [32]byte
	Message   string
}

// pendingFileRequest holds an unresolved incoming file transfer request.
type pendingFileRequest struct {
	FriendID uint32
	FileID   uint32
	Filename string
	FileSize uint64
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

	// file transfer state
	pendingFileReqs  []pendingFileRequest
	activeFileReqIdx int

	// current conversation
	activeFriendID  uint32
	activeFriend    string
	historyByFriend map[uint32][]chatMessage
	lastTypingState bool

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
		client:          client,
		modalInput:      mi,
		historyByFriend: make(map[uint32][]chatMessage),
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

// tickCmd returns a command that triggers periodic UI updates.
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
		return a.handleFriendRequestEvent(m)

	case toxclient.FriendMessageEvent:
		a.handleIncomingMessage(m.FriendID, m.Message, false)
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendConnectionStatusEvent:
		a.contacts.updateConnectionStatus(m.FriendID, m.Status)
		a.refreshContacts()
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendNameEvent:
		a.handleFriendNameChange(m)
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendStatusMessageEvent:
		a.refreshContacts()
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FriendTypingEvent:
		a.handleFriendTyping(m)
		return a, waitForToxEvent(a.client.Events())

	case toxclient.SelfConnectionStatusEvent:
		a.statusBar.connectionStatus = m.Status
		return a, waitForToxEvent(a.client.Events())

	case toxclient.AnonymityStatusEvent:
		a.handleAnonymityStatusChange(m)
		return a, waitForToxEvent(a.client.Events())

	// File transfer events
	case toxclient.FileRecvRequestEvent:
		return a.handleFileRecvRequest(m)

	case toxclient.FileRecvCompleteEvent:
		a.handleFileRecvComplete(m)
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FileSendCompleteEvent:
		a.handleFileSendComplete(m)
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FileTransferErrorEvent:
		a.setNotification(fmt.Sprintf("File transfer failed: %s", m.Error))
		return a, waitForToxEvent(a.client.Events())

	case toxclient.FileRecvChunkEvent:
		return a, waitForToxEvent(a.client.Events())

	case toxclient.TickEvent:
		a.handleTick()
		return a, tickCmd()
	}

	return a, nil
}

// handleFriendRequestEvent processes an incoming friend request.
func (a App) handleFriendRequestEvent(m toxclient.FriendRequestEvent) (tea.Model, tea.Cmd) {
	a.pendingReqs = append(a.pendingReqs, pendingRequest{
		PublicKey: m.PublicKey,
		Message:   m.Message,
	})
	if a.modal == modalNone {
		a.openFriendRequestModal()
	}
	return a, waitForToxEvent(a.client.Events())
}

// handleFriendNameChange updates the display name when a friend changes theirs.
func (a *App) handleFriendNameChange(m toxclient.FriendNameEvent) {
	a.contacts.updateName(m.FriendID, m.Name)
	if m.FriendID == a.activeFriendID {
		a.activeFriend = m.Name
		a.chat.friendName = m.Name
	}
}

// handleFriendTyping updates the typing indicator for the active conversation.
func (a *App) handleFriendTyping(m toxclient.FriendTypingEvent) {
	if m.FriendID == a.activeFriendID {
		a.chat.isTyping = m.IsTyping
		a.chat.refreshViewport()
	}
}

// handleAnonymityStatusChange processes Tor/I2P status updates.
func (a *App) handleAnonymityStatusChange(m toxclient.AnonymityStatusEvent) {
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
}

// handleFileRecvRequest processes an incoming file transfer request.
func (a App) handleFileRecvRequest(m toxclient.FileRecvRequestEvent) (tea.Model, tea.Cmd) {
	a.pendingFileReqs = append(a.pendingFileReqs, pendingFileRequest{
		FriendID: m.FriendID,
		FileID:   m.FileID,
		Filename: m.Filename,
		FileSize: m.FileSize,
	})
	if a.modal == modalNone {
		a.openFileRequestModal()
	}
	return a, waitForToxEvent(a.client.Events())
}

// handleFileRecvComplete processes a completed incoming file transfer.
func (a *App) handleFileRecvComplete(m toxclient.FileRecvCompleteEvent) {
	name := a.friendName(m.FriendID)
	a.handleIncomingFileMessage(m.FriendID, fmt.Sprintf("Received file: %s → %s", m.Filename, m.SavePath))
	a.setNotification(fmt.Sprintf("File from %s saved: %s", name, m.SavePath))
}

// handleFileSendComplete processes a completed outgoing file transfer.
func (a *App) handleFileSendComplete(m toxclient.FileSendCompleteEvent) {
	a.handleOutgoingFileMessage(m.FriendID, fmt.Sprintf("File sent: %s", m.Filename))
	a.setNotification(fmt.Sprintf("File sent: %s", m.Filename))
}

// handleTick processes periodic UI updates.
func (a *App) handleTick() {
	a.refreshContacts()
	if !a.notifyExpiry.IsZero() && time.Now().After(a.notifyExpiry) {
		a.notification = ""
	}
}

// handleResize processes window size changes.
func (a App) handleResize(m tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	a.width = m.Width
	a.height = m.Height

	statusH := 1
	mainH := a.height - statusH

	chatW := a.width - contactsPanelWidth

	if !a.ready {
		a.contacts = newContactsPanel(contactsPanelWidth, mainH)
		a.chat = newChatPanel(chatW, mainH)
		a.statusBar = newStatusBar(a.width)
		a.statusBar.selfAddress = a.client.SelfAddress()
		a.contacts.focused = true
		a.focus = focusContacts
		a.ready = true

		a.refreshContacts()
		a.client.Bootstrap()
	} else {
		a.contacts.width = contactsPanelWidth
		a.contacts.height = mainH
		a.chat.resize(chatW, mainH)
		a.statusBar.width = a.width
	}

	return a, nil
}

// handleKey processes keyboard input.
func (a App) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle global shortcuts first.
	if model, cmd, handled := a.handleGlobalShortcut(m); handled {
		return model, cmd
	}

	// Handle modal-specific keys.
	if model, cmd, handled := a.handleModalKey(m); handled {
		return model, cmd
	}

	// Handle panel input.
	return a.handlePanelInput(m)
}

// handleGlobalShortcut processes application-wide keyboard shortcuts.
// Returns (model, cmd, handled) where handled indicates if the key was consumed.
func (a App) handleGlobalShortcut(m tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch m.String() {
	case "ctrl+c", "ctrl+q":
		model, cmd := a.quit()
		return model, cmd, true
	case "ctrl+s":
		a.handleSaveProfile()
		return a, nil, true
	case "ctrl+n":
		if a.modal == modalNone {
			a.openAddFriendModal()
		}
		return a, nil, true
	case "ctrl+g":
		a.setNotification("Group chat not yet supported.")
		return a, nil, true
	case "tab":
		a.toggleFocus()
		return a, nil, true
	case "esc":
		if a.modal != modalNone {
			a.closeModal()
		}
		return a, nil, true
	}
	return a, nil, false
}

// handleSaveProfile attempts to save the Tox profile and notifies the user.
func (a *App) handleSaveProfile() {
	if err := a.client.Save(); err != nil {
		a.setNotification(fmt.Sprintf("Save failed: %v", err))
	} else {
		a.setNotification("Profile saved.")
	}
}

// handleModalKey processes keys when a modal dialog is active.
// Returns (model, cmd, handled) where handled indicates if the key was consumed.
func (a App) handleModalKey(m tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch m.String() {
	case "enter":
		return a.handleModalEnter()
	case "r":
		return a.handleModalReject()
	}

	// Handle text input for add friend modal.
	if a.modal == modalAddFriend {
		var cmd tea.Cmd
		a.modalInput, cmd = a.modalInput.Update(m)
		return a, cmd, true
	}

	return a, nil, false
}

// handleModalEnter processes the enter key in modal dialogs.
func (a App) handleModalEnter() (tea.Model, tea.Cmd, bool) {
	switch a.modal {
	case modalAddFriend:
		model, cmd := a.submitAddFriend()
		return model, cmd, true
	case modalFriendRequest:
		model, cmd := a.acceptFriendRequest()
		return model, cmd, true
	case modalFileRequest:
		model, cmd := a.acceptFileRequest()
		return model, cmd, true
	}
	return a, nil, false
}

// handleModalReject processes the reject key in modal dialogs.
func (a App) handleModalReject() (tea.Model, tea.Cmd, bool) {
	switch a.modal {
	case modalFriendRequest:
		model, cmd := a.rejectFriendRequest()
		return model, cmd, true
	case modalFileRequest:
		model, cmd := a.rejectFileRequest()
		return model, cmd, true
	}
	return a, nil, false
}

// handlePanelInput processes keyboard input for the active panel.
func (a App) handlePanelInput(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.focus == focusContacts {
		a.handleContactsInput(m)
	} else {
		return a.handleChatInput(m)
	}
	return a, nil
}

// handleContactsInput processes keyboard input for the contacts panel.
func (a *App) handleContactsInput(m tea.KeyMsg) {
	selected, friendID := a.contacts.update(m)
	if selected {
		a.selectFriend(friendID)
	}
}

// handleChatInput processes keyboard input for the chat panel.
func (a App) handleChatInput(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	text, submitted := a.chat.update(m)
	if submitted {
		return a.sendMessage(text)
	}
	a.updateTypingStatus()
	return a, nil
}

// updateTypingStatus notifies the remote peer of typing state changes.
func (a *App) updateTypingStatus() {
	if a.activeFriendID == 0 {
		return
	}
	isTyping := len(a.chat.input.Value()) > 0
	if isTyping != a.lastTypingState {
		_ = a.client.SetTyping(a.activeFriendID, isTyping)
		a.lastTypingState = isTyping
	}
}

// handleMouse processes mouse events.
func (a App) handleMouse(m tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Determine if click is in contacts or chat area.
	if m.X < contactsPanelWidth {
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
	// Save current friend's history before switching.
	if a.activeFriendID != 0 {
		a.historyByFriend[a.activeFriendID] = a.chat.history
	}
	a.activeFriendID = friendID
	a.activeFriend = a.friendName(friendID)
	a.lastTypingState = false
	// Restore target friend's history (may be nil if first time).
	savedHistory := a.historyByFriend[friendID]
	a.chat.setFriendWithHistory(friendID, a.activeFriend, savedHistory)
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
		// Store message in history map for non-active friend.
		a.historyByFriend[friendID] = append(a.historyByFriend[friendID], msg)
		a.contacts.incrementUnread(friendID)
	}
}

// sendMessage sends the current input as a message to the active friend.
// It also handles the /file command for sending files.
func (a App) sendMessage(text string) (tea.Model, tea.Cmd) {
	if a.activeFriendID == 0 {
		return a, nil
	}

	// Handle /file command
	if strings.HasPrefix(text, "/file ") {
		filePath := strings.TrimPrefix(text, "/file ")
		filePath = strings.TrimSpace(filePath)
		return a.sendFile(filePath)
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

// sendFile initiates a file transfer to the active friend.
func (a App) sendFile(filePath string) (tea.Model, tea.Cmd) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		a.setNotification(fmt.Sprintf("Cannot read file: %v", err))
		return a, nil
	}

	// Get filename from path
	filename := filepath.Base(filePath)

	// Initiate the transfer
	_, err = a.client.FileSend(a.activeFriendID, filename, data)
	if err != nil {
		a.setNotification(fmt.Sprintf("File send failed: %v", err))
		return a, nil
	}

	// Add a message to the chat indicating file transfer started
	msg := chatMessage{
		ts:       time.Now(),
		senderID: 0,
		name:     "You",
		body:     fmt.Sprintf("📤 Sending file: %s (%d bytes)", filename, len(data)),
	}
	a.chat.addMessage(msg)
	a.setNotification(fmt.Sprintf("Sending file: %s", filename))
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

// openFileRequestModal shows the incoming file request dialog.
func (a *App) openFileRequestModal() {
	if len(a.pendingFileReqs) == 0 {
		return
	}
	a.modal = modalFileRequest
	a.activeFileReqIdx = 0
}

// acceptFileRequest accepts the current pending file transfer.
func (a App) acceptFileRequest() (tea.Model, tea.Cmd) {
	if len(a.pendingFileReqs) == 0 {
		a.closeModal()
		return a, nil
	}
	req := a.pendingFileReqs[a.activeFileReqIdx]
	if err := a.client.FileAccept(req.FriendID, req.FileID, req.FileSize, req.Filename); err != nil {
		a.setNotification(fmt.Sprintf("Accept failed: %v", err))
	} else {
		name := a.friendName(req.FriendID)
		a.setNotification(fmt.Sprintf("Receiving file from %s: %s", name, req.Filename))
		// Add a message to the chat
		a.handleIncomingFileMessage(req.FriendID, fmt.Sprintf("📥 Receiving file: %s (%d bytes)", req.Filename, req.FileSize))
	}
	a.pendingFileReqs = append(a.pendingFileReqs[:a.activeFileReqIdx], a.pendingFileReqs[a.activeFileReqIdx+1:]...)
	if len(a.pendingFileReqs) == 0 {
		a.closeModal()
	}
	return a, nil
}

// rejectFileRequest rejects the current pending file transfer.
func (a App) rejectFileRequest() (tea.Model, tea.Cmd) {
	if len(a.pendingFileReqs) == 0 {
		a.closeModal()
		return a, nil
	}
	req := a.pendingFileReqs[a.activeFileReqIdx]
	_ = a.client.FileReject(req.FriendID, req.FileID)
	a.pendingFileReqs = append(a.pendingFileReqs[:a.activeFileReqIdx], a.pendingFileReqs[a.activeFileReqIdx+1:]...)
	if len(a.pendingFileReqs) == 0 {
		a.closeModal()
	}
	return a, nil
}

// handleIncomingFileMessage adds a file-related message to the chat history.
func (a *App) handleIncomingFileMessage(friendID uint32, text string) {
	name := a.friendName(friendID)
	msg := chatMessage{
		ts:       time.Now(),
		senderID: friendID,
		name:     name,
		body:     text,
	}
	if friendID == a.activeFriendID {
		a.chat.addMessage(msg)
	} else {
		a.historyByFriend[friendID] = append(a.historyByFriend[friendID], msg)
		a.contacts.incrementUnread(friendID)
	}
}

// handleOutgoingFileMessage adds an outgoing file-related message to the chat.
func (a *App) handleOutgoingFileMessage(friendID uint32, text string) {
	msg := chatMessage{
		ts:       time.Now(),
		senderID: 0,
		name:     "You",
		body:     text,
	}
	if friendID == a.activeFriendID {
		a.chat.addMessage(msg)
	} else {
		a.historyByFriend[friendID] = append(a.historyByFriend[friendID], msg)
	}
}

// formatFileSize formats a byte count as a human-readable string.
func formatFileSize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// setNotification displays a temporary notification message.
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
	content := a.renderModalContent()
	if content == "" {
		return base
	}

	dialog := modalStyle.Render(content)
	return a.overlayDialogOnBase(base, dialog)
}

// renderModalContent generates the content string for the current modal.
func (a App) renderModalContent() string {
	switch a.modal {
	case modalAddFriend:
		return a.renderAddFriendModal()
	case modalFriendRequest:
		return a.renderFriendRequestModal()
	case modalFileRequest:
		return a.renderFileRequestModal()
	}
	return ""
}

// renderAddFriendModal renders the Add Friend dialog content.
func (a App) renderAddFriendModal() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s",
		panelTitle.Render("Add Friend"),
		a.modalPrompt,
		a.modalInput.View(),
	)
}

// renderFriendRequestModal renders the Friend Request dialog content.
func (a App) renderFriendRequestModal() string {
	if len(a.pendingReqs) == 0 {
		return ""
	}
	req := a.pendingReqs[a.activeReqIdx]
	pkHex := fmt.Sprintf("%x", req.PublicKey)
	if len(pkHex) > 20 {
		pkHex = pkHex[:20] + "..."
	}
	return fmt.Sprintf("%s\n\nFrom: %s\nMessage: %s\n\n[Enter] Accept  [R] Reject  [Esc] Dismiss",
		panelTitle.Render("Friend Request"),
		pkHex,
		req.Message,
	)
}

// renderFileRequestModal renders the Incoming File dialog content.
func (a App) renderFileRequestModal() string {
	if len(a.pendingFileReqs) == 0 {
		return ""
	}
	req := a.pendingFileReqs[a.activeFileReqIdx]
	name := a.friendName(req.FriendID)
	return fmt.Sprintf("%s\n\nFrom: %s\nFile: %s\nSize: %s\n\n[Enter] Accept  [R] Reject  [Esc] Dismiss",
		panelTitle.Render("Incoming File"),
		name,
		req.Filename,
		formatFileSize(req.FileSize),
	)
}

// overlayDialogOnBase positions and renders a dialog box over the base view.
func (a App) overlayDialogOnBase(base, dialog string) string {
	dialogW := lipgloss.Width(dialog)
	dialogH := lipgloss.Height(dialog)

	x := max(0, (a.width-dialogW)/2)
	y := max(0, (a.height-dialogH)/2)

	lines := strings.Split(base, "\n")
	dialogLines := strings.Split(dialog, "\n")

	for i, dl := range dialogLines {
		row := y + i
		lines = a.insertDialogLine(lines, row, x, dl)
	}

	return strings.Join(lines, "\n")
}

// insertDialogLine inserts a dialog line at the specified position in the view.
func (a App) insertDialogLine(lines []string, row, x int, dialogLine string) []string {
	// Ensure the row exists.
	for row >= len(lines) {
		lines = append(lines, "")
	}

	line := lines[row]
	// Pad line to at least x characters.
	runes := []rune(line)
	for len(runes) < x {
		runes = append(runes, ' ')
	}

	// Replace characters at position x with dialog line.
	dlRunes := []rune(dialogLine)
	end := x + len(dlRunes)
	for len(runes) < end {
		runes = append(runes, ' ')
	}
	copy(runes[x:], dlRunes)
	lines[row] = string(runes)

	return lines
}
