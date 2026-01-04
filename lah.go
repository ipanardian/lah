// Package lah provides a modern replacement for the Unix ls command.
// It displays file listings in a beautiful table format with colors,
// git integration, and human-readable file sizes.
//
// GitHub: https://github.com/ipanardian/lah
// Author: Ipan Ardian
// Version: v1.0.0
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ipanardian/lah/box/table"
)

type FileInfo struct {
	Name      string
	Path      string
	Size      int64
	Mode      fs.FileMode
	ModTime   time.Time
	IsDir     bool
	IsHidden  bool
	GitStatus string
}

type Config struct {
	SortModified bool
	Reverse      bool
	ShowGit      bool
}

func main() {
	var config Config

	var rootCmd = &cobra.Command{
		Use:   "lah [path]",
		Short: "A beautiful ls replacement with table formatting",
		Long: `lah is a modern replacement for ls with box-drawn tables, colors, and git integration.

GitHub: https://github.com/ipanardian/lah
Author: Ipan Ardian
Version: v1.0.0`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			if err := listDirectory(path, config); err != nil {
				log.Fatalf("Error: %v", err)
			}
		},
	}

	rootCmd.Flags().BoolVarP(&config.SortModified, "sort-modified", "t", false, "sort by modified time (newest first)")
	rootCmd.Flags().BoolVarP(&config.Reverse, "reverse", "r", false, "reverse sort order")
	rootCmd.Flags().BoolVarP(&config.ShowGit, "git", "g", false, "show git status inline")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showColoredHelp(cmd)
	})

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func listDirectory(path string, config Config) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var files []FileInfo
	now := time.Now()

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			Size:     info.Size(),
			Mode:     info.Mode(),
			ModTime:  info.ModTime(),
			IsDir:    entry.IsDir(),
			IsHidden: strings.HasPrefix(entry.Name(), "."),
		}

		if config.ShowGit {
			fileInfo.GitStatus = getGitStatus(fileInfo.Path)
		}

		files = append(files, fileInfo)
	}

	sortFiles(files, config)

	printTable(files, now, config)

	return nil
}

func sortFiles(files []FileInfo, config Config) {
	if config.SortModified {
		sort.Slice(files, func(i, j int) bool {
			if config.Reverse {
				return files[i].ModTime.Before(files[j].ModTime)
			}
			return files[i].ModTime.After(files[j].ModTime)
		})
	} else {
		sort.Slice(files, func(i, j int) bool {
			if files[i].IsDir != files[j].IsDir {
				return files[i].IsDir
			}

			result := strings.Compare(strings.ToLower(files[i].Name), strings.ToLower(files[j].Name))
			if config.Reverse {
				return result > 0
			}
			return result < 0
		})
	}
}

func getGitStatus(filePath string) string {
	dir := filepath.Dir(filePath)
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return ""
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return ""
	}

	relPath, err := filepath.Rel(dir, filePath)
	if err != nil {
		return ""
	}

	status, err := worktree.Status()
	if err != nil {
		return ""
	}

	fileStatus := status.File(relPath)
	if fileStatus.Worktree == git.Unmodified && fileStatus.Staging == git.Unmodified {
		return "(clean)"
	}

	var added, deleted int
	if fileStatus.Worktree == git.Added || fileStatus.Staging == git.Added {
		added++
	}
	if fileStatus.Worktree == git.Deleted || fileStatus.Staging == git.Deleted {
		deleted++
	}
	if fileStatus.Worktree == git.Modified || fileStatus.Staging == git.Modified {
		added++
	}

	if added > 0 || deleted > 0 {
		return fmt.Sprintf("+%d -%d", added, deleted)
	}

	return ""
}

func getTerminalWidth() int {
	if width := os.Getenv("COLUMNS"); width != "" {
		if w, err := strconv.Atoi(width); err == nil && w > 0 {
			return w - 10
		}
	}

	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width - 10
	}

	if cmd := exec.Command("tput", "cols"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if w, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil && w > 0 {
				return w - 10
			}
		}
	}

	return 70
}

func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			j := i + 1
			if j < len(s) && s[j] == '[' {
				j++
				for j < len(s) && (s[j] < 'a' || s[j] > 'z') && (s[j] < 'A' || s[j] > 'Z') {
					j++
				}
				j++
			}
			i = j
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

func printTable(files []FileInfo, now time.Time, config Config) {
	if len(files) == 0 {
		return
	}

	terminalWidth := max(getTerminalWidth(), 40)

	data := make([][]string, len(files)+1)

	headers := []string{"Name", "Size", "Modified", "Perms"}
	if config.ShowGit {
		headers = append(headers, "Git")
	}
	data[0] = headers

	for i, file := range files {
		row := []string{
			formatName(file),
			formatSize(file.Size, file.IsDir),
			formatModified(file.ModTime, now),
			formatPermissions(file.Mode),
		}
		if config.ShowGit {
			row = append(row, formatGitStatus(file.GitStatus))
		}
		data[i+1] = row
	}

	displayWidths := calculateDisplayWidths(data)
	mins, maxs := columnConstraints(config.ShowGit)
	for i := range displayWidths {
		if i < len(mins) && mins[i] > 0 && displayWidths[i] < mins[i] {
			displayWidths[i] = mins[i]
		}
		if i < len(maxs) && maxs[i] > 0 && displayWidths[i] > maxs[i] {
			displayWidths[i] = maxs[i]
		}
	}

	// Ensure terminal can display at least the minimum widths
	minContentWidth := 0
	for i := range displayWidths {
		minContentWidth += lookupMin(mins, i, 4)
	}
	minBorderWidth := (len(displayWidths)-1)*3 + 2
	if terminalWidth < minContentWidth+minBorderWidth {
		fmt.Println("Terminal is too small to display the table. Please widen your terminal window.")
		return
	}

	// Adjust widths to fit terminal
	totalContentWidth := 0
	for _, w := range displayWidths {
		totalContentWidth += w
	}
	borderWidth := (len(displayWidths)-1)*3 + 2
	totalWidth := totalContentWidth + borderWidth

	if totalWidth > terminalWidth {
		excess := totalWidth - terminalWidth
		totalShrinkable := 0

		for i, w := range displayWidths {
			if i != 1 && i != 3 { // Don't shrink Size and Perms columns
				minWidth := lookupMin(mins, i, 4)
				if w-minWidth > 0 {
					totalShrinkable += w - minWidth
				}
			}
		}

		// Shrink columns proportionally
		for i := range displayWidths {
			if i != 1 && i != 3 { // Don't shrink Size and Perms columns
				minWidth := lookupMin(mins, i, 4)
				shrinkable := displayWidths[i] - minWidth
				if shrinkable > 0 && totalShrinkable > 0 {
					shrinkAmount := (shrinkable * excess) / totalShrinkable
					if shrinkAmount > shrinkable {
						shrinkAmount = shrinkable
					}
					displayWidths[i] -= shrinkAmount
					// Ensure we don't go below minimum
					if displayWidths[i] < minWidth {
						displayWidths[i] = minWidth
					}
					excess -= shrinkAmount
					totalShrinkable -= shrinkable
				}
			}
		}
	}

	table := table.NewTableWithWidths(data, displayWidths)
	table.SetBorderStyle(0)
	table.SetHeaderStyle(1)
	table.SetHeaderColor(color.New(color.FgCyan, color.Bold))
	table.SetBorderColor(color.New(color.FgGreen))

	table.Print()
}

func calculateDisplayWidths(data [][]string) []int {
	if len(data) == 0 {
		return nil
	}

	rows := len(data)
	cols := len(data[0])
	widths := make([]int, cols)

	for i := range rows {
		for j := range cols {
			displayText := stripANSI(data[i][j])
			width := utf8.RuneCountInString(displayText)
			if width > widths[j] {
				widths[j] = width
			}
		}
	}

	return widths
}

func columnConstraints(showGit bool) ([]int, []int) {
	// Columns: Name, Size, Modified, Perms, (Git)
	mins := []int{15, 6, 10, 10}
	maxs := []int{50, 10, 15, 12}
	if showGit {
		mins = append(mins, 6)
		maxs = append(maxs, 12)
	}
	return mins, maxs
}

func lookupMin(mins []int, idx int, fallback int) int {
	if idx < len(mins) && mins[idx] > 0 {
		return mins[idx]
	}
	return fallback
}

func formatName(file FileInfo) string {
	name := file.Name

	if file.IsDir {
		return color.New(color.FgBlue, color.Bold).Sprint(name)
	}

	if file.Mode.Perm()&0111 != 0 {
		return color.New(color.FgRed).Sprint(name)
	}

	if file.IsHidden {
		return color.New(color.FgYellow).Sprint(name)
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go", ".rs", ".py", ".js", ".ts", ".jsx", ".tsx":
		return color.New(color.FgGreen).Sprint(name)
	case ".md", ".txt", ".rst":
		return color.New(color.FgYellow).Sprint(name)
	case ".yml", ".yaml", ".json", ".toml", ".ini":
		return color.New(color.FgMagenta).Sprint(name)
	default:
		return color.New(color.FgWhite).Sprint(name)
	}
}

func formatSize(size int64, isDir bool) string {
	if isDir {
		return color.New(color.FgCyan).Sprint("-")
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	result := fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])

	return color.New(color.FgHiWhite).Sprint(result)
}

func formatModified(t time.Time, now time.Time) string {
	duration := now.Sub(t)

	var c *color.Color
	var text string

	if duration < 0 {
		c = color.New(color.FgBlue)
		text = "future"
	} else if duration < time.Minute {
		c = color.New(color.FgGreen)
		text = fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		c = color.New(color.FgGreen)
		text = fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		c = color.New(color.FgYellow)
		text = fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else if duration < 7*24*time.Hour {
		c = color.New(color.FgHiYellow)
		text = fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	} else if duration < 30*24*time.Hour {
		c = color.New(color.FgRed)
		text = fmt.Sprintf("%d weeks ago", int(duration.Hours()/(24*7)))
	} else if duration < 365*24*time.Hour {
		c = color.New(color.FgHiRed)
		text = fmt.Sprintf("%d months ago", int(duration.Hours()/(24*30)))
	} else {
		c = color.New(color.FgHiBlack)
		text = fmt.Sprintf("%d years ago", int(duration.Hours()/(24*365)))
	}

	return c.Sprint(text)
}

func formatPermissions(mode fs.FileMode) string {
	perm := mode.Perm()

	var result strings.Builder

	switch {
	case mode&fs.ModeDir != 0:
		result.WriteString(color.New(color.FgCyan, color.Bold).Sprint("d"))
	case mode&fs.ModeSymlink != 0:
		result.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("l"))
	case mode&fs.ModeDevice != 0:
		if mode&fs.ModeCharDevice != 0 {
			result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("c"))
		} else {
			result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("b"))
		}
	case mode&fs.ModeNamedPipe != 0:
		result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("p"))
	case mode&fs.ModeSocket != 0:
		result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("s"))
	default:
		result.WriteString(color.New(color.FgCyan).Sprint("-"))
	}

	for i := 8; i >= 0; i-- {
		bit := perm >> uint(i) & 1
		var c *color.Color

		switch (8 - i) % 3 {
		case 0:
			if bit == 1 {
				c = color.New(color.FgGreen, color.Bold)
				result.WriteString(c.Sprint("r"))
			} else {
				c = color.New(color.FgHiBlack)
				result.WriteString(c.Sprint("-"))
			}
		case 1:
			if bit == 1 {
				c = color.New(color.FgYellow, color.Bold)
				result.WriteString(c.Sprint("w"))
			} else {
				c = color.New(color.FgHiBlack)
				result.WriteString(c.Sprint("-"))
			}
		case 2:
			if bit == 1 {
				if mode&fs.ModeSetuid != 0 {
					c = color.New(color.FgMagenta, color.Bold)
					result.WriteString(c.Sprint("s"))
				} else if mode&fs.ModeSetgid != 0 {
					c = color.New(color.FgMagenta, color.Bold)
					result.WriteString(c.Sprint("s"))
				} else if mode&fs.ModeSticky != 0 {
					c = color.New(color.FgRed, color.Bold)
					result.WriteString(c.Sprint("t"))
				} else {
					c = color.New(color.FgRed, color.Bold)
					result.WriteString(c.Sprint("x"))
				}
			} else {
				c = color.New(color.FgHiBlack)
				result.WriteString(c.Sprint("-"))
			}
		}
	}

	return result.String()
}

func formatGitStatus(status string) string {
	if status == "" {
		return ""
	}

	if status == "(clean)" {
		return color.New(color.FgGreen).Sprint(status)
	}

	if strings.Contains(status, "+") {
		return color.New(color.FgGreen, color.Bold).Sprint(status)
	}

	return color.New(color.FgYellow).Sprint(status)
}

func showColoredHelp(_ *cobra.Command) {
	fmt.Printf("\n%s %s\n\n",
		color.New(color.FgCyan, color.Bold).Sprint("lah v1.0.0"),
		color.New(color.FgHiWhite).Sprint("- A beautiful ls replacement"),
	)
	fmt.Printf("%s\n", color.New(color.FgHiBlack).Sprint("GitHub: https://github.com/ipanardian/lah"))
	fmt.Printf("%s\n\n", color.New(color.FgHiBlack).Sprint("Author: Ipan Ardian"))

	fmt.Printf("%s\n\n", color.New(color.FgWhite).Sprint("USAGE:"))
	fmt.Printf("  lah [path] [flags]\n\n")

	fmt.Printf("%s\n", color.New(color.FgWhite, color.Bold).Sprint("FLAGS:"))

	flags := []struct {
		flag, desc string
	}{
		{"-t, --sort-modified", "sort by modified time (newest first)"},
		{"-r, --reverse", "reverse sort order"},
		{"-g, --git", "show git status inline"},
		{"-h, --help", "show this help message"},
	}

	for _, f := range flags {
		fmt.Printf("  %s\t%s\n",
			color.New(color.FgCyan, color.Bold).Sprintf("%-20s", f.flag),
			color.New(color.FgHiWhite).Sprint(f.desc),
		)
	}

	fmt.Printf("\n%s\n", color.New(color.FgWhite, color.Bold).Sprint("EXAMPLES:"))
	examples := []string{
		"lah",
		"lah -t",
		"lah -tr",
		"lah -g",
		"lah -tg",
	}

	for _, ex := range examples {
		fmt.Printf("  %s\n", color.New(color.FgGreen).Sprint(ex))
	}

	fmt.Println()
}
