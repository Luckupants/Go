//go:build !solution

package ciletters

import (
	_ "embed"
	"strings"
	"text/template"
)

func LastRows(log string) []string {
	answer := strings.Split(log, "\n")
	return answer[max(0, len(answer)-10):]
}

//go:embed lolkekcheburek.txt
var ff string

func MakeLetter(n *Notification) (string, error) {
	tmpl := template.New("test")
	tmpl = tmpl.Funcs(template.FuncMap{
		"last_rows": LastRows,
	})
	var err error
	tmpl, err = tmpl.Parse(ff)
	if err != nil {
		return "", err
	}
	answer := strings.Builder{}
	err = tmpl.Execute(&answer, n)
	return answer.String(), err
}
