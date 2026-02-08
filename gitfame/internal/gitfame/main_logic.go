package gitfame

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func Run(options Options, output io.Writer, errorOutput io.Writer) error {
	if err := validateOptions(options); err != nil {
		fmt.Fprintln(errorOutput, err.Error())
		return err
	}

	allowedExtensions, err := buildExtensionFilter(options.Exts)
	if err != nil {
		return err
	}

	allowedLanguageExtensions, useLanguageFilter, err := buildLanguageFilter(
		options.Langs,
		errorOutput,
	)
	if err != nil {
		return err
	}

	files, err := getRepositoryFiles(options.Repo, options.Rev)
	if err != nil {
		fmt.Fprintln(errorOutput, err)
		return err
	}

	authorData, err := processFiles(
		files,
		options,
		allowedExtensions,
		allowedLanguageExtensions,
		useLanguageFilter,
		errorOutput,
	)
	if err != nil {
		return err
	}

	results := calculateResults(authorData)
	sortResults(results, options.Order)

	return formatOutput(results, options.Format, output)
}

func validateOptions(options Options) error {
	if options.Order != "lines" && options.Order != "commits" && options.Order != "files" {
		return fmt.Errorf("invalid value for --order-by")
	}

	if options.Format != "tabular" && options.Format != "csv" && options.Format != "json" &&
		options.Format != "json-lines" {
		return fmt.Errorf("invalid value for --format")
	}

	return nil
}

func buildExtensionFilter(extensions []string) (map[string]bool, error) {
	allowed := make(map[string]bool)

	for _, ext := range extensions {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}

		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}

		ext = strings.ToLower(ext)
		allowed[ext] = true
	}

	return allowed, nil
}

func buildLanguageFilter(languages []string, errorOutput io.Writer) (map[string]bool, bool, error) {
	allowed := make(map[string]bool)
	useFilter := false

	if len(languages) == 0 {
		return allowed, useFilter, nil
	}

	languageDefinitions, err := loadLanguageDefinitions()
	if err != nil {
		fmt.Fprintln(errorOutput, err)
		return nil, false, err
	}

	requestedLanguages := make(map[string]bool)
	originalLanguages := make([]string, 0, len(languages))

	for _, lang := range languages {
		lang = strings.TrimSpace(lang)
		if lang == "" {
			continue
		}

		originalLanguages = append(originalLanguages, lang)
		requestedLanguages[strings.ToLower(lang)] = true
	}

	foundLanguages := make(map[string]bool)

	for _, definition := range languageDefinitions {
		nameLower := strings.ToLower(strings.TrimSpace(definition.Name))
		if !requestedLanguages[nameLower] {
			continue
		}

		foundLanguages[nameLower] = true
		useFilter = true

		for _, ext := range definition.Extensions {
			ext = strings.TrimSpace(ext)
			if ext == "" {
				continue
			}

			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}

			ext = strings.ToLower(ext)
			allowed[ext] = true
		}
	}

	for _, original := range originalLanguages {
		if !foundLanguages[strings.ToLower(original)] {
			fmt.Fprintln(errorOutput, "warning: unknown language", original)
		}
	}

	return allowed, useFilter, nil
}

func loadLanguageDefinitions() ([]LanguageDefinition, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("cannot get caller info")
	}

	configPath := filepath.Join(
		filepath.Dir(currentFile),
		"..", "..", "configs", "language_extensions.json",
	)

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var definitions []LanguageDefinition
	if err := json.NewDecoder(file).Decode(&definitions); err != nil {
		return nil, err
	}

	return definitions, nil
}

func processFiles(
	files []string,
	options Options,
	allowedExtensions map[string]bool,
	allowedLanguageExtensions map[string]bool,
	useLanguageFilter bool,
	errorOutput io.Writer,
) (map[string]*AuthorData, error) {
	authorData := make(map[string]*AuthorData)

	for _, file := range files {
		if !shouldProcessFile(
			file,
			allowedExtensions,
			allowedLanguageExtensions,
			useLanguageFilter,
			options.Rests,
			options.Excls,
		) {
			continue
		}

		content, err := getFileContent(options.Repo, options.Rev, file)
		if err != nil {
			fmt.Fprintln(errorOutput, err)
			return nil, err
		}

		if len(content) == 0 {
			if err := processEmptyFile(authorData, options, file); err != nil {
				fmt.Fprintln(errorOutput, err)
				return nil, err
			}

			continue
		}

		if err := processFileWithContent(authorData, options, file); err != nil {
			fmt.Fprintln(errorOutput, err)
			return nil, err
		}
	}

	return authorData, nil
}

func processEmptyFile(authorData map[string]*AuthorData, options Options, filePath string) error {
	hash, author, committer, err := getEmptyFileCommitInfo(options.Repo, options.Rev, filePath)
	if err != nil {
		return err
	}

	if hash == "" {
		return nil
	}

	name := author
	if options.UseCmtr {
		name = committer
	}

	if name == "" {
		return nil
	}

	data := authorData[name]
	if data == nil {
		data = &AuthorData{
			name:    name,
			commits: make(map[string]struct{}),
			files:   make(map[string]struct{}),
		}
		authorData[name] = data
	}

	data.commits[hash] = struct{}{}
	data.files[filePath] = struct{}{}

	return nil
}

func processFileWithContent(
	authorData map[string]*AuthorData,
	options Options,
	filePath string,
) error {
	output, err := getBlameOutput(options.Repo, options.Rev, filePath)
	if err != nil {
		return err
	}

	var (
		currentHash, currentAuthor, currentCommitter string
		currentLines                                 int
	)

	processCurrentBlock := func() {
		if currentHash == "" || currentLines == 0 {
			return
		}

		name := currentAuthor
		if options.UseCmtr {
			name = currentCommitter
		}

		if name == "" {
			return
		}

		data := authorData[name]
		if data == nil {
			data = &AuthorData{
				name:    name,
				commits: make(map[string]struct{}),
				files:   make(map[string]struct{}),
			}
			authorData[name] = data
		}

		data.lines += currentLines
		data.commits[currentHash] = struct{}{}
		data.files[filePath] = struct{}{}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		if line[0] == '\t' {
			currentLines++
			continue
		}

		parts := strings.SplitN(line, " ", 4)
		if len(parts) >= 3 && isHash(parts[0]) {
			if currentHash != "" {
				processCurrentBlock()
			}

			currentHash = parts[0]
			currentAuthor = ""
			currentCommitter = ""
			currentLines = 0

			continue
		}

		parts = strings.SplitN(line, " ", 2)
		key := parts[0]

		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}

		switch key {
		case "author":
			currentAuthor = value
		case "committer":
			currentCommitter = value
		}
	}

	if currentHash != "" {
		processCurrentBlock()
	}

	return nil
}

func calculateResults(authorData map[string]*AuthorData) []AuthorStats {
	results := make([]AuthorStats, 0, len(authorData))

	for _, data := range authorData {
		results = append(results, AuthorStats{
			Name:    data.name,
			Lines:   data.lines,
			Commits: len(data.commits),
			Files:   len(data.files),
		})
	}

	return results
}

func sortResults(results []AuthorStats, order string) {
	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]

		switch order {
		case "lines":
			if a.Lines != b.Lines {
				return a.Lines > b.Lines
			}

			if a.Commits != b.Commits {
				return a.Commits > b.Commits
			}

			if a.Files != b.Files {
				return a.Files > b.Files
			}

		case "commits":
			if a.Commits != b.Commits {
				return a.Commits > b.Commits
			}

			if a.Lines != b.Lines {
				return a.Lines > b.Lines
			}

			if a.Files != b.Files {
				return a.Files > b.Files
			}

		case "files":
			if a.Files != b.Files {
				return a.Files > b.Files
			}

			if a.Lines != b.Lines {
				return a.Lines > b.Lines
			}

			if a.Commits != b.Commits {
				return a.Commits > b.Commits
			}
		}

		return a.Name < b.Name
	})
}

func formatOutput(results []AuthorStats, format string, output io.Writer) error {
	switch format {
	case "tabular":
		return formatTabularOutput(results, output)
	case "csv":
		return formatCSVOutput(results, output)
	case "json":
		return formatJSONOutput(results, output)
	case "json-lines":
		return formatJSONLinesOutput(results, output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
