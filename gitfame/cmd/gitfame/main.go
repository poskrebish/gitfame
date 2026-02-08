//go:build !solution

package main

import (
	"os"

	"github.com/spf13/pflag"
	"gitlab.com/slon/shad-go/06-BHW-Gitfame/gitfame/internal/gitfame"
)

func main() {
	repo := pflag.String("repository", ".", "")
	rev := pflag.String("revision", "HEAD", "")
	order := pflag.String("order-by", "lines", "")
	cmtr := pflag.Bool("use-committer", false, "")
	frmt := pflag.String("format", "tabular", "")
	exts := pflag.StringSlice("extensions", nil, "")
	langs := pflag.StringSlice("languages", nil, "")
	excls := pflag.StringSlice("exclude", nil, "")
	rests := pflag.StringSlice("restrict-to", nil, "")

	pflag.Parse()

	opts := gitfame.Options{
		Repo:    *repo,
		Rev:     *rev,
		Order:   *order,
		UseCmtr: *cmtr,
		Format:  *frmt,
		Exts:    *exts,
		Langs:   *langs,
		Excls:   *excls,
		Rests:   *rests,
	}

	if err := gitfame.Run(opts, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
