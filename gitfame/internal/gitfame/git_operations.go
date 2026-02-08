package gitfame

import (
	"fmt"
	"os/exec"
	"strings"
)

func getRepositoryFiles(repoPath string, revision string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "ls-tree", "-r", "--name-only", revision)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(output), "\n")

	var result []string

	for _, file := range files {
		if file != "" {
			result = append(result, file)
		}
	}

	return result, nil
}

func getFileContent(repoPath string, revision string, filePath string) ([]byte, error) {
	cmd := exec.Command(
		"git",
		"-C",
		repoPath,
		"cat-file",
		"-p",
		fmt.Sprintf("%s:%s", revision, filePath),
	)

	return cmd.Output()
}

func getEmptyFileCommitInfo(
	repoPath string,
	revision string,
	filePath string,
) (string, string, string, error) {
	cmd := exec.Command(
		"git",
		"-C",
		repoPath,
		"log",
		"-1",
		"--format=%H%x00%an%x00%cn",
		revision,
		"--",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return "", "", "", err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "\x00")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid log output format")
	}

	return parts[0], parts[1], parts[2], nil
}

func getBlameOutput(repoPath string, revision string, filePath string) ([]byte, error) {
	cmd := exec.Command(
		"git",
		"-C",
		repoPath,
		"blame",
		"--line-porcelain",
		revision,
		"--",
		filePath,
	)

	return cmd.Output()
}

func isHash(s string) bool {
	if len(s) != 40 {
		return false
	}

	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}

	return true
}
