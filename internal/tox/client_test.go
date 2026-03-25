package tox

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileKey(t *testing.T) {
	key1 := fileKey{friendID: 1, fileID: 10}
	key2 := fileKey{friendID: 1, fileID: 10}
	key3 := fileKey{friendID: 2, fileID: 10}

	// Keys with same values should be equal.
	if key1 != key2 {
		t.Error("fileKey with same values should be equal")
	}

	// Keys with different values should not be equal.
	if key1 == key3 {
		t.Error("fileKey with different friendID should not be equal")
	}
}

func TestOutgoingFile_Fields(t *testing.T) {
	of := outgoingFile{
		filename: "test.txt",
		data:     []byte("hello"),
		sent:     3,
	}

	if of.filename != "test.txt" {
		t.Errorf("outgoingFile.filename = %q, want %q", of.filename, "test.txt")
	}
	if string(of.data) != "hello" {
		t.Errorf("outgoingFile.data = %q, want %q", string(of.data), "hello")
	}
	if of.sent != 3 {
		t.Errorf("outgoingFile.sent = %d, want 3", of.sent)
	}
}

func TestIncomingFile_Fields(t *testing.T) {
	inf := incomingFile{
		filename: "received.txt",
		size:     1024,
		data:     []byte("data"),
		received: 512,
	}

	if inf.filename != "received.txt" {
		t.Errorf("incomingFile.filename = %q, want %q", inf.filename, "received.txt")
	}
	if inf.size != 1024 {
		t.Errorf("incomingFile.size = %d, want 1024", inf.size)
	}
	if inf.received != 512 {
		t.Errorf("incomingFile.received = %d, want 512", inf.received)
	}
}

func TestUniqueFilename_NotExists(t *testing.T) {
	// Test with a path that doesn't exist.
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "unique.txt")

	got := uniqueFilename(path)
	if got != path {
		t.Errorf("uniqueFilename for non-existent path = %q, want %q", got, path)
	}
}

func TestUniqueFilename_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "existing.txt")

	// Create the file.
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	got := uniqueFilename(path)
	expected := filepath.Join(tmpDir, "existing_1.txt")
	if got != expected {
		t.Errorf("uniqueFilename for existing path = %q, want %q", got, expected)
	}
}

func TestUniqueFilename_MultipleExist(t *testing.T) {
	tmpDir := t.TempDir()
	base := filepath.Join(tmpDir, "test.txt")

	// Create multiple files.
	for i := 0; i < 3; i++ {
		var path string
		if i == 0 {
			path = base
		} else {
			path = filepath.Join(tmpDir, "test_"+string(rune('0'+i))+".txt")
		}
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	got := uniqueFilename(base)
	expected := filepath.Join(tmpDir, "test_3.txt")
	if got != expected {
		t.Errorf("uniqueFilename with multiple existing = %q, want %q", got, expected)
	}
}

func TestUniqueFilename_WithExtension(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "document.pdf")

	// Create the file.
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	got := uniqueFilename(path)
	expected := filepath.Join(tmpDir, "document_1.pdf")
	if got != expected {
		t.Errorf("uniqueFilename with extension = %q, want %q", got, expected)
	}
}

func TestUniqueFilename_NoExtension(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "README")

	// Create the file.
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	got := uniqueFilename(path)
	expected := filepath.Join(tmpDir, "README_1")
	if got != expected {
		t.Errorf("uniqueFilename without extension = %q, want %q", got, expected)
	}
}

func TestProfilePath_Default(t *testing.T) {
	// Save and clear custom path.
	original, hadOriginal := os.LookupEnv("MTOX_PROFILE_PATH")
	os.Unsetenv("MTOX_PROFILE_PATH")
	defer func() {
		if hadOriginal {
			os.Setenv("MTOX_PROFILE_PATH", original)
		}
	}()

	path := ProfilePath()
	if path == "" {
		t.Error("ProfilePath() returned empty string")
	}

	// Should contain profile.tox.
	if filepath.Base(path) != "profile.tox" {
		t.Errorf("ProfilePath() = %q, should end with profile.tox", path)
	}

	// Should contain .config/mtox directory.
	dir := filepath.Dir(path)
	if filepath.Base(dir) != "mtox" {
		t.Errorf("ProfilePath() dir = %q, should end with mtox", dir)
	}
}

func TestIsAnonOnlyMode_Values(t *testing.T) {
	original, hadOriginal := os.LookupEnv("MTOX_ANON_ONLY")
	defer func() {
		if hadOriginal {
			os.Setenv("MTOX_ANON_ONLY", original)
		} else {
			os.Unsetenv("MTOX_ANON_ONLY")
		}
	}()

	tests := []struct {
		value    string
		setEnv   bool
		expected bool
	}{
		{"1", true, true},
		{"0", true, false},
		{"true", true, false}, // Only "1" is treated as true
		{"yes", true, false},  // Only "1" is treated as true
		{"", true, false},
		{"", false, false}, // Unset
	}

	for _, tt := range tests {
		if tt.setEnv {
			os.Setenv("MTOX_ANON_ONLY", tt.value)
		} else {
			os.Unsetenv("MTOX_ANON_ONLY")
		}

		got := IsAnonOnlyMode()
		if got != tt.expected {
			t.Errorf("IsAnonOnlyMode() with value %q (set=%v) = %v, want %v",
				tt.value, tt.setEnv, got, tt.expected)
		}
	}
}

func TestFileTransferEvents_Interface(t *testing.T) {
	// Verify all file transfer event types implement ToxEvent.
	events := []ToxEvent{
		FileRecvRequestEvent{FriendID: 1, FileID: 1, Kind: 0, FileSize: 100, Filename: "test.txt"},
		FileRecvChunkEvent{FriendID: 1, FileID: 1, Position: 0, Data: []byte("data")},
		FileChunkRequestEvent{FriendID: 1, FileID: 1, Position: 0, Length: 1024},
		FileSendCompleteEvent{FriendID: 1, FileID: 1, Filename: "sent.txt"},
		FileRecvCompleteEvent{FriendID: 1, FileID: 1, Filename: "recv.txt", SavePath: "/tmp/recv.txt"},
		FileTransferErrorEvent{FriendID: 1, FileID: 1, Filename: "error.txt", Error: "failed"},
	}

	for i, e := range events {
		e.toxEvent() // Should not panic
		if e == nil {
			t.Errorf("events[%d] is nil", i)
		}
	}

	if len(events) != 6 {
		t.Errorf("Expected 6 file transfer event types, got %d", len(events))
	}
}

func TestFileRecvRequestEvent_Fields(t *testing.T) {
	e := FileRecvRequestEvent{
		FriendID: 42,
		FileID:   7,
		Kind:     0,
		FileSize: 1024,
		Filename: "document.pdf",
	}

	if e.FriendID != 42 {
		t.Errorf("FileRecvRequestEvent.FriendID = %d, want 42", e.FriendID)
	}
	if e.FileID != 7 {
		t.Errorf("FileRecvRequestEvent.FileID = %d, want 7", e.FileID)
	}
	if e.Kind != 0 {
		t.Errorf("FileRecvRequestEvent.Kind = %d, want 0", e.Kind)
	}
	if e.FileSize != 1024 {
		t.Errorf("FileRecvRequestEvent.FileSize = %d, want 1024", e.FileSize)
	}
	if e.Filename != "document.pdf" {
		t.Errorf("FileRecvRequestEvent.Filename = %q, want %q", e.Filename, "document.pdf")
	}
}

func TestFileRecvCompleteEvent_Fields(t *testing.T) {
	e := FileRecvCompleteEvent{
		FriendID: 3,
		FileID:   5,
		Filename: "image.png",
		SavePath: "/home/user/downloads/image.png",
	}

	if e.FriendID != 3 {
		t.Errorf("FileRecvCompleteEvent.FriendID = %d, want 3", e.FriendID)
	}
	if e.FileID != 5 {
		t.Errorf("FileRecvCompleteEvent.FileID = %d, want 5", e.FileID)
	}
	if e.Filename != "image.png" {
		t.Errorf("FileRecvCompleteEvent.Filename = %q, want %q", e.Filename, "image.png")
	}
	if e.SavePath != "/home/user/downloads/image.png" {
		t.Errorf("FileRecvCompleteEvent.SavePath = %q, want %q", e.SavePath, "/home/user/downloads/image.png")
	}
}

func TestFileTransferErrorEvent_Fields(t *testing.T) {
	e := FileTransferErrorEvent{
		FriendID: 1,
		FileID:   2,
		Filename: "broken.zip",
		Error:    "connection lost",
	}

	if e.FriendID != 1 {
		t.Errorf("FileTransferErrorEvent.FriendID = %d, want 1", e.FriendID)
	}
	if e.FileID != 2 {
		t.Errorf("FileTransferErrorEvent.FileID = %d, want 2", e.FileID)
	}
	if e.Filename != "broken.zip" {
		t.Errorf("FileTransferErrorEvent.Filename = %q, want %q", e.Filename, "broken.zip")
	}
	if e.Error != "connection lost" {
		t.Errorf("FileTransferErrorEvent.Error = %q, want %q", e.Error, "connection lost")
	}
}

func TestClientConstants(t *testing.T) {
	// Verify important constants have sensible values.
	if profileDir == "" {
		t.Error("profileDir constant is empty")
	}
	if profileFile == "" {
		t.Error("profileFile constant is empty")
	}
	if eventBufSize < 1 {
		t.Errorf("eventBufSize = %d, should be >= 1", eventBufSize)
	}
}

func TestBootstrapNodes_Valid(t *testing.T) {
	if len(bootstrapNodes) == 0 {
		t.Fatal("bootstrapNodes is empty")
	}

	for i, node := range bootstrapNodes {
		if node.host == "" {
			t.Errorf("bootstrapNodes[%d].host is empty", i)
		}
		if node.port == 0 {
			t.Errorf("bootstrapNodes[%d].port is 0", i)
		}
		// Public keys should be 64 hex characters (32 bytes).
		if len(node.pubkey) != 64 {
			t.Errorf("bootstrapNodes[%d].pubkey length = %d, want 64", i, len(node.pubkey))
		}
		// Verify it's valid hex.
		for j, c := range node.pubkey {
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
				t.Errorf("bootstrapNodes[%d].pubkey[%d] = %c, not valid hex", i, j, c)
			}
		}
	}
}

func TestFileSendCompleteEvent_Fields(t *testing.T) {
	e := FileSendCompleteEvent{
		FriendID: 10,
		FileID:   20,
		Filename: "upload.txt",
	}

	if e.FriendID != 10 {
		t.Errorf("FileSendCompleteEvent.FriendID = %d, want 10", e.FriendID)
	}
	if e.FileID != 20 {
		t.Errorf("FileSendCompleteEvent.FileID = %d, want 20", e.FileID)
	}
	if e.Filename != "upload.txt" {
		t.Errorf("FileSendCompleteEvent.Filename = %q, want %q", e.Filename, "upload.txt")
	}
}

func TestFileRecvChunkEvent_Fields(t *testing.T) {
	data := []byte("chunk data here")
	e := FileRecvChunkEvent{
		FriendID: 5,
		FileID:   15,
		Position: 1024,
		Data:     data,
	}

	if e.FriendID != 5 {
		t.Errorf("FileRecvChunkEvent.FriendID = %d, want 5", e.FriendID)
	}
	if e.FileID != 15 {
		t.Errorf("FileRecvChunkEvent.FileID = %d, want 15", e.FileID)
	}
	if e.Position != 1024 {
		t.Errorf("FileRecvChunkEvent.Position = %d, want 1024", e.Position)
	}
	if string(e.Data) != string(data) {
		t.Errorf("FileRecvChunkEvent.Data = %q, want %q", string(e.Data), string(data))
	}
}

func TestFileChunkRequestEvent_Fields(t *testing.T) {
	e := FileChunkRequestEvent{
		FriendID: 8,
		FileID:   18,
		Position: 2048,
		Length:   512,
	}

	if e.FriendID != 8 {
		t.Errorf("FileChunkRequestEvent.FriendID = %d, want 8", e.FriendID)
	}
	if e.FileID != 18 {
		t.Errorf("FileChunkRequestEvent.FileID = %d, want 18", e.FileID)
	}
	if e.Position != 2048 {
		t.Errorf("FileChunkRequestEvent.Position = %d, want 2048", e.Position)
	}
	if e.Length != 512 {
		t.Errorf("FileChunkRequestEvent.Length = %d, want 512", e.Length)
	}
}
