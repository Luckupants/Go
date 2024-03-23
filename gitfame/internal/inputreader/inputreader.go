package inputreader

import (
	"fmt"
	"gitlab.com/slon/shad-go/gitfame/internal/information"
	"os"
	"os/exec"
	"slices"

	"github.com/spf13/pflag"
)

func ParseFlags() information.InputInfo {
	curPath, err := exec.Command("pwd").Output()
	if err != nil {
		fmt.Printf("Couldn't get working path : %v", err)
	}
	info := information.InputInfo{
		FlagPath:         pflag.String("repository", string(curPath), "Path to git repository"),
		FlagCommit:       pflag.String("revision", "HEAD", "Commit to work with"),
		FlagOrderBy:      pflag.String("order-by", "lines", "Sort output by"),
		FlagUseCommitter: pflag.Bool("use-committer", false, "Use committer over author"),
		FlagFormat:       pflag.String("format", "tabular", "Format for output"),
		FlagExtensions:   pflag.StringSlice("extensions", nil, "Extensions to work with"),
		FlagLanguages:    pflag.StringSlice("languages", nil, "Languages to work with"),
		FlagExclude:      pflag.StringSlice("exclude", nil, "Patterns to exclude"),
		FlagRestrict:     pflag.StringSlice("restrict-to", nil, "Exclude files that don't match any given pattern"),
		FlagGoroutines:   pflag.Int("goroutines", 8, "Amount of goroutines working"),
	}
	pflag.Parse()
	return info
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CheckInput can Exit if something goes wrong...
func CheckInput(info information.InputInfo) error {
	exists, err := exists(*info.FlagPath)
	if err != nil {
		return fmt.Errorf("error working with given repository path: %v", err)
	} else if !exists {
		return fmt.Errorf("given repository path does not exists")
	}
	orders := []string{"lines", "commits", "files"}
	if !slices.Contains(orders, *info.FlagOrderBy) {
		return fmt.Errorf("invalid order. Must be one of these types : %v", orders)
	}
	formats := []string{"tabular", "csv", "json", "json-lines"}
	if !slices.Contains(formats, *info.FlagFormat) {
		return fmt.Errorf("invalid format. Must be one of these types : %v", formats)
	}
	for _, extension := range *info.FlagExtensions {
		if len(extension) == 0 {
			return fmt.Errorf("detected empty extension")
		}
	}
	for _, language := range *info.FlagLanguages {
		if len(language) == 0 {
			return fmt.Errorf("detected empty language")
		}
	}
	for _, exclusion := range *info.FlagExclude {
		if len(exclusion) == 0 {
			return fmt.Errorf("detected empty exclusion")
		}
	}
	for _, restriction := range *info.FlagRestrict {
		if len(restriction) == 0 {
			return fmt.Errorf("detected empty restriction")
		}
	}
	return nil
}
