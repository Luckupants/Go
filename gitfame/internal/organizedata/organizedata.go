package organizedata

import (
	"errors"
	"gitlab.com/slon/shad-go/gitfame/internal/information"
	"sort"
)

type outputOrder struct {
	data []information.FameInfo
	less func(fi1 information.FameInfo, fi2 information.FameInfo) bool
}

func (oo outputOrder) Len() int           { return len(oo.data) }
func (oo outputOrder) Less(i, j int) bool { return oo.less(oo.data[i], oo.data[j]) }
func (oo outputOrder) Swap(i, j int)      { oo.data[i], oo.data[j] = oo.data[j], oo.data[i] }

func PrepareForOutput(answer map[string]*information.FameInfo, info information.InputInfo) ([]information.FameInfo, error) {
	var ans []information.FameInfo
	for _, val := range answer {
		ans = append(ans, *val)
	}
	oo := outputOrder{data: ans}
	switch *info.FlagOrderBy {
	case "lines":
		oo.less = func(fi1 information.FameInfo, fi2 information.FameInfo) bool {
			if fi1.LinesAmount != fi2.LinesAmount {
				return fi1.LinesAmount > fi2.LinesAmount
			}
			if fi1.CommitsAmount != fi2.CommitsAmount {
				return fi1.CommitsAmount > fi2.CommitsAmount
			}
			if fi1.FilesAmount != fi2.FilesAmount {
				return fi1.FilesAmount > fi2.FilesAmount
			}
			return fi1.Name < fi2.Name
		}
	case "commits":
		oo.less = func(fi1 information.FameInfo, fi2 information.FameInfo) bool {
			if fi1.CommitsAmount != fi2.CommitsAmount {
				return fi1.CommitsAmount > fi2.CommitsAmount
			}
			if fi1.LinesAmount != fi2.LinesAmount {
				return fi1.LinesAmount > fi2.LinesAmount
			}
			if fi1.FilesAmount != fi2.FilesAmount {
				return fi1.FilesAmount > fi2.FilesAmount
			}
			return fi1.Name < fi2.Name
		}
	case "files":
		oo.less = func(fi1 information.FameInfo, fi2 information.FameInfo) bool {
			if fi1.FilesAmount != fi2.FilesAmount {
				return fi1.FilesAmount > fi2.FilesAmount
			}
			if fi1.LinesAmount != fi2.LinesAmount {
				return fi1.LinesAmount > fi2.LinesAmount
			}
			if fi1.CommitsAmount != fi2.CommitsAmount {
				return fi1.CommitsAmount > fi2.CommitsAmount
			}
			return fi1.Name < fi2.Name
		}
	default:
		return nil, errors.New("incorrect output order")
	}
	sort.Sort(oo)
	return ans, nil
}
