//go:build !solution

package fileleak

import (
	"fmt"
	"os"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

func getFiles(entries []os.DirEntry) map[string]string {
	answer := make(map[string]string)
	for _, value := range entries {
		if !value.IsDir() {
			path := fmt.Sprintf("/proc/self/fd/%s", value.Name())
			link, err := os.Readlink(path)
			if err == nil {
				answer[value.Name()] = link
			}
		}
	}
	return answer
}

func checkLeak(t testingT, start, end map[string]string) {
	for fd := range end {
		if _, ok := start[fd]; !ok || start[fd] != end[fd] {
			t.Errorf("leak detected")
			return
		}
	}
}

func VerifyNone(t testingT) {
	startEntries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Errorf("can't open proc")
		return
	}
	start := getFiles(startEntries)
	t.Cleanup(func() {
		endEntries, err := os.ReadDir("/proc/self/fd")
		if err != nil {
			t.Errorf("can't open proc")
		}
		end := getFiles(endEntries)
		checkLeak(t, start, end)
	})
}
