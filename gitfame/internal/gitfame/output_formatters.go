package gitfame

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"
)

func formatTabularOutput(results []AuthorStats, output io.Writer) error {
	writer := tabwriter.NewWriter(output, 0, 0, 1, ' ', 0)
	fmt.Fprintln(writer, "Name\tLines\tCommits\tFiles")

	for _, result := range results {
		fmt.Fprintf(
			writer,
			"%s\t%d\t%d\t%d\n",
			result.Name,
			result.Lines,
			result.Commits,
			result.Files,
		)
	}

	return writer.Flush()
}

func formatCSVOutput(results []AuthorStats, output io.Writer) error {
	writer := csv.NewWriter(output)

	if err := writer.Write([]string{"Name", "Lines", "Commits", "Files"}); err != nil {
		return err
	}

	for _, result := range results {
		record := []string{
			result.Name,
			strconv.Itoa(result.Lines),
			strconv.Itoa(result.Commits),
			strconv.Itoa(result.Files),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()

	return writer.Error()
}

func formatJSONOutput(results []AuthorStats, output io.Writer) error {
	jsonAuthors := make([]JSONAuthor, 0, len(results))
	for _, result := range results {
		jsonAuthors = append(jsonAuthors, JSONAuthor(result))
	}

	data, err := json.Marshal(jsonAuthors)
	if err != nil {
		return err
	}

	_, err = output.Write(append(data, '\n'))

	return err
}

func formatJSONLinesOutput(results []AuthorStats, output io.Writer) error {
	for _, result := range results {
		jsonAuthor := JSONAuthor(result)

		data, err := json.Marshal(jsonAuthor)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(output, string(data)); err != nil {
			return err
		}
	}

	return nil
}
