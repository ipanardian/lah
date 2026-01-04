package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCalculateDisplayWidths(t *testing.T) {
	data := [][]string{
		{"Name", "Size", "Modified"},
		{"file.txt", "1.2 KB", "2 minutes ago"},
		{"very-long-filename.go", "15.3 KB", "1 hour ago"},
	}

	widths := calculateDisplayWidths(data)

	expected := []int{12, 7, 12}
	for i, w := range widths {
		if i < len(expected) && w < expected[i] {
			t.Errorf("Column %d width %d is less than expected minimum %d", i, w, expected[i])
		}
	}
}

func TestFormatPermissions(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected string
	}{
		{"regular file", 0644, "-rw-r--r--"},
		{"directory", 0755 | os.ModeDir, "drwxr-xr-x"},
		{"executable", 0755, "-rwxr-xr-x"},
		{"symlink", 0777 | os.ModeSymlink, "lrwxrwxrwx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPermissions(tt.mode)
			clean := stripANSI(result)
			if clean != tt.expected {
				t.Errorf("formatPermissions(%v) = %q, want %q", tt.mode, clean, tt.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		isDir    bool
		expected string
	}{
		{"directory", 0, true, "-"},
		{"bytes", 512, false, "512 B"},
		{"kilobytes", 1536, false, "1.5 KB"},
		{"megabytes", 2097152, false, "2.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.size, tt.isDir)
			// Remove ANSI codes for comparison
			clean := stripANSI(result)
			if clean != tt.expected {
				t.Errorf("formatSize(%d, %v) = %q, want %q", tt.size, tt.isDir, clean, tt.expected)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mRed Text\x1b[0m"
	expected := "Red Text"
	result := stripANSI(input)
	if result != expected {
		t.Errorf("stripANSI(%q) = %q, want %q", input, result, expected)
	}
}

func TestSortFiles(t *testing.T) {
	now := time.Now()
	files := []FileInfo{
		{Name: "file.txt", ModTime: now.Add(-time.Hour), IsDir: false},
		{Name: "dir", ModTime: now.Add(-time.Minute), IsDir: true},
		{Name: "another.txt", ModTime: now, IsDir: false},
	}

	config := Config{}
	sortFiles(files, config)

	if !files[0].IsDir {
		t.Error("First item should be a directory")
	}
	if files[0].Name != "dir" {
		t.Errorf("Expected 'dir' first, got %q", files[0].Name)
	}

	if files[1].Name != "another.txt" {
		t.Errorf("Expected 'another.txt' second, got %q", files[1].Name)
	}
}

func TestGetTerminalWidth(t *testing.T) {
	width := getTerminalWidth()
	if width <= 0 {
		t.Error("Terminal width should be positive")
	}
	if width > 1000 {
		t.Errorf("Terminal width %d seems too large", width)
	}
}

func TestFileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	fileInfo := FileInfo{
		Name:  info.Name(),
		Size:  info.Size(),
		IsDir: info.IsDir(),
	}

	if fileInfo.Name != "test.txt" {
		t.Errorf("Expected name 'test.txt', got %q", fileInfo.Name)
	}
	if fileInfo.IsDir {
		t.Error("File should not be a directory")
	}
	if fileInfo.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), fileInfo.Size)
	}
}
