package errorhandling

import (
	"fmt"
	"os"
)

func CheckError(err error) {
	if err != nil {
		fmt.Printf("%s", err.Error())
		os.Exit(1)
	}
}
