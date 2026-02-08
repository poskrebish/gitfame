package gitfame

type AuthorStats struct {
	Name    string
	Lines   int
	Commits int
	Files   int
}

type Options struct {
	Repo    string
	Rev     string
	Order   string
	UseCmtr bool
	Format  string
	Exts    []string
	Langs   []string
	Excls   []string
	Rests   []string
}

type LanguageDefinition struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Extensions []string `json:"extensions"`
}

type AuthorData struct {
	name    string
	lines   int
	commits map[string]struct{}
	files   map[string]struct{}
}

type JSONAuthor struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}
