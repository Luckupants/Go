package progressbar

import (
	"fmt"
	"os"
	"strings"
)

// Хотел воспользоваться библиотекой на гитхабе, но уверен проверяющая система ее не подгрузит...

var Red = "\033[31m"

type ProgressBar struct {
	progress int
	Delta    int
}

func (pb *ProgressBar) SendMessage(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message) // не могу в stdout писать...
}

func (pb *ProgressBar) UpdateProgress(progress int) {
	if progress > pb.progress+pb.Delta {
		pb.progress = (progress / pb.Delta) * pb.Delta
		_, _ = fmt.Fprintln(os.Stderr, "cur progress:") // не могу в stdout писать...
		_, _ = fmt.Fprintln(os.Stderr, Red+strings.Repeat("▆", pb.progress/pb.Delta)+strings.Repeat(" ", (100-pb.progress)/pb.Delta), pb.progress, "%")
	}
}
