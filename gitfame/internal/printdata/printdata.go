package printdata

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/slon/shad-go/gitfame/internal/information"
	"os"
	"strconv"
	"text/tabwriter"
)

func PrintTabular(output []information.FameInfo) error {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 1, 1, ' ', 0)
	_, err := fmt.Fprintln(w, "Name\tLines\tCommits\tFiles")
	if err != nil {
		return err
	}
	for _, value := range output {
		_, err = fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", value.Name, value.LinesAmount, value.CommitsAmount, value.FilesAmount)
		if err != nil {
			return err
		}
	}
	err = w.Flush()
	return err
}

func PrintCSV(output []information.FameInfo) error {
	w := csv.NewWriter(os.Stdout)
	toWrite := []string{"Name", "Lines", "Commits", "Files"}
	err := w.Write(toWrite)
	if err != nil {
		return err
	}
	for _, value := range output {
		toWrite = []string{value.Name, strconv.Itoa(value.LinesAmount), strconv.Itoa(value.CommitsAmount), strconv.Itoa(value.FilesAmount)}
		err = w.Write(toWrite)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

func PrintJSON(output []information.FameInfo) error {
	res, err := json.Marshal(output)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}

func PrintJSONLines(output []information.FameInfo) error {
	for _, value := range output {
		res, err := json.Marshal(value)
		if err != nil {
			return err
		}
		fmt.Println(string(res))
	}
	return nil
}

func PrintAnswer(answer []information.FameInfo, info information.InputInfo) error {
	switch *info.FlagFormat {
	case "tabular":
		return PrintTabular(answer)
	case "csv":
		return PrintCSV(answer)
	case "json":
		return PrintJSON(answer)
	case "json-lines":
		return PrintJSONLines(answer)
	default:
		return errors.New("incorrect format")
	}
}
