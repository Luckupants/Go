package core

import (
	"errors"
	"gitlab.com/slon/shad-go/gitfame/internal/information"
	"gitlab.com/slon/shad-go/gitfame/internal/inputreader"
	"gitlab.com/slon/shad-go/gitfame/internal/organizedata"
	"gitlab.com/slon/shad-go/gitfame/internal/printdata"
	"gitlab.com/slon/shad-go/gitfame/pkg/progressbar"
	"maps"
	"os/exec"
	"strconv"
	"strings"
)

func skipHeader(data []string) []string {
	var firstWord string
	for firstWord != "filename" {
		firstWord = strings.Split(data[0], " ")[0]
		data = data[1:]
	}
	return data
}

func parseBlame(answer map[string]information.FameInfo, data []string, seenCommits map[string]string, seenNames map[string]struct{}, info information.InputInfo) ([]string, error) {
	linesAmount, err := strconv.Atoi(strings.Split(data[0], " ")[3])
	if err != nil {
		return nil, errors.Join(err, errors.New("error parsing git blame"))
	}
	commit := strings.Split(data[0], " ")[0]
	name := ""
	if _, ok := seenCommits[commit]; !ok {
		if *info.FlagUseCommitter {
			_, name, _ = strings.Cut(data[5], " ")
		} else {
			_, name, _ = strings.Cut(data[1], " ")
		}
		seenCommits[commit] = name
		data = skipHeader(data)[1:]
	} else {
		name = seenCommits[commit]
		data = data[2:]
	}
	seenNames[name] = struct{}{}
	if _, ok := answer[name]; !ok {
		answer[name] = information.FameInfo{Name: name, Commits: make(map[string]struct{})}
	}
	answer[name].Commits[commit] = struct{}{}
	answer[name].LinesAmount += linesAmount
	for i := 0; i < linesAmount-1; i++ {
		commit = strings.Split(data[0], " ")[0]
		answer[name].Commits[commit] = struct{}{}
		data = data[2:]
	}
	return data, nil
}

func getLogData(info information.InputInfo, file string) (string, string, error) {
	cmd := exec.Command("git", "log", *info.FlagCommit, "--max-count=1", "--pretty=format:%an%n%cn%n%H", "--", file)
	cmd.Dir = *info.FlagPath
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	answer := strings.Split(string(output), "\n")
	if *info.FlagUseCommitter {
		return answer[1], answer[2], nil
	}
	return answer[0], answer[2], nil
}

type channelInfo struct {
	file   string
	info   information.InputInfo
	answer map[string]*information.FameInfo
}

func fame(files []string, info information.InputInfo, answer map[string]*information.FameInfo, pb *progressbar.ProgressBar) error {
	for _, file := range files {
		result, err := blame(info, file)
		if err != nil {
			return err
		}
		for _, res := range result {
			if _, ok := answer[res.Name]; !ok {
				answer[res.Name] = res
			} else {
				answer[res.Name].FilesAmount += res.FilesAmount
				answer[res.Name].LinesAmount += res.LinesAmount
				maps.Copy(answer[res.Name].Commits, res.Commits)
			}
		}
	}
	for _, val := range answer {
		val.CommitsAmount = len(val.Commits)
	}
	return nil
}

func blame(info information.InputInfo, file string) (map[string]*information.FameInfo, error) {
	var (
		seenCommits = make(map[string]string)
		seenNames   = make(map[string]struct{})
	)
	cmd := exec.Command("git", "blame", *info.FlagCommit, file, "--porcelain")
	cmd.Dir = *info.FlagPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	data := strings.Split(string(output), "\n")
	if len(data) == 1 {
		name, commit, err2 := getLogData(info, file)
		if err2 != nil {
			return nil, err2
		}
		result := information.FameInfo{Name: name, Commits: make(map[string]struct{})}
		result.Commits[commit] = struct{}{}
		result.FilesAmount++
		return map[string]*information.FameInfo{name: &result}, nil
	}
	result := make(map[string]information.FameInfo)
	for len(data) > 1 {
		data, err = parseBlame(result, data, seenCommits, seenNames, info)
		if err != nil {
			return nil, err
		}
	}
	for name := range seenNames {
		result[name].FilesAmount++
	}
	return result, nil
}

func Execute(info information.InputInfo, pb *progressbar.ProgressBar) error {
	pb.SendMessage("Loading files...")
	files, err := inputreader.GetFiles(info)
	if err != nil {
		return err
	}
	pb.SendMessage("Filtering...")
	files = inputreader.FilterExtensions(files, *info.FlagExtensions)
	files, err = inputreader.FilterLanguages(files, info)
	if err != nil {
		return err
	}
	files, err = inputreader.ExcludePatterns(files, *info.FlagExclude, *info.FlagPath)
	if err != nil {
		return err
	}
	files, err = inputreader.Restrict(files, *info.FlagRestrict, *info.FlagPath)
	if err != nil {
		return err
	}
	answer := make(map[string]*information.FameInfo)
	pb.SendMessage("Looking for commits...")
	err = fame(files, info, answer, pb)
	if err != nil {
		return err
	}
	pb.SendMessage("Preparing output...")
	ans, err := organizedata.PrepareForOutput(answer, info)
	if err != nil {
		return err
	}
	err = printdata.PrintAnswer(ans, info)
	if err != nil {
		return err
	}
	return nil
}
