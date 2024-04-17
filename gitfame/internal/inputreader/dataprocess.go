package inputreader

import (
	"encoding/json"
	"fmt"
	"gitlab.com/slon/shad-go/gitfame/configs"
	"gitlab.com/slon/shad-go/gitfame/internal/information"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func GetFiles(info information.InputInfo) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", *info.FlagCommit, ":", *info.FlagPath, "--format=%(objecttype) %(path)")
	cmd.Dir = *info.FlagPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var answer []string
	splittedOut := strings.Split(string(out), "\n")
	splittedOut = splittedOut[:len(splittedOut)-1]
	for _, line := range splittedOut {
		if strings.Fields(line)[0] == "blob" {
			answer = append(answer, strings.Join(strings.Fields(line)[1:], " "))
		}
	}
	for i := 0; i < len(answer); i++ {
		answer[i] = filepath.Join(*info.FlagPath, answer[i])
	}
	return answer, nil
}

func CheckExtension(fileName string, extensions []string) bool {
	for _, extension := range extensions {
		if strings.HasSuffix(strings.TrimSpace(fileName), strings.TrimSpace(extension)) {
			return true
		}
	}
	return false
}

func FilterExtensions(fileNames []string, extensions []string) []string {
	if len(extensions) == 0 {
		return fileNames
	}
	var answer []string
	for _, fileName := range fileNames {
		if CheckExtension(fileName, extensions) {
			answer = append(answer, fileName)
		}
	}
	return answer
}

type LanguageInfo struct {
	Name       string   `json:"Name"`
	FileType   string   `json:"type"`
	Extensions []string `json:"extensions"`
}

func FilterLanguages(fileNames []string, info information.InputInfo) ([]string, error) {
	var langInfo []LanguageInfo
	err := json.Unmarshal(configs.Languages, &langInfo)
	if err != nil {
		return nil, err
	}
	var allowedExtensions []string
	for _, language := range *info.FlagLanguages {
		index := slices.IndexFunc(langInfo, func(languageInfo LanguageInfo) bool {
			return strings.EqualFold(languageInfo.Name, language)
		})
		if index == -1 {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("language not found: %v", language))
			continue
		}
		allowedExtensions = append(allowedExtensions, langInfo[index].Extensions...)
	}
	fileNames = FilterExtensions(fileNames, allowedExtensions)
	return fileNames, nil
}

func ExcludePatterns(fileNames []string, patterns []string, sourcePath string) ([]string, error) {
	var answer []string
	for _, fileName := range fileNames {
		ok := false
		for _, pattern := range patterns {
			matches, err := filepath.Match(filepath.Join(sourcePath, pattern), fileName)
			ok = matches
			if err != nil {
				return nil, err
			}
			if ok {
				break
			}
		}
		if !ok {
			answer = append(answer, fileName)
		}
	}
	return answer, nil
}

func Restrict(fileNames []string, patterns []string, sourcePath string) ([]string, error) {
	var answer []string
	if len(patterns) == 0 {
		answer = fileNames
	}
	for _, fileName := range fileNames {
		for _, pattern := range patterns {
			matches, err := filepath.Match(filepath.Join(sourcePath, pattern), fileName)
			if err != nil {
				return nil, err
			}
			if matches {
				answer = append(answer, fileName)
				break
			}
		}
	}
	return answer, nil
}
