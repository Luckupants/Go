package information

type FameInfo struct {
	Commits       map[string]struct{} `json:"-"`
	Name          string              `json:"name"`
	LinesAmount   int                 `json:"lines"`
	CommitsAmount int                 `json:"commits"`
	FilesAmount   int                 `json:"files"`
}
