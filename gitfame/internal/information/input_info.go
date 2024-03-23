package information

type InputInfo struct {
	FlagPath         *string
	FlagCommit       *string
	FlagOrderBy      *string
	FlagUseCommitter *bool
	FlagFormat       *string
	FlagExtensions   *[]string
	FlagLanguages    *[]string
	FlagExclude      *[]string
	FlagRestrict     *[]string
	FlagGoroutines   *int
}
