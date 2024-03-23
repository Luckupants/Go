//go:build !solution

package main

import (
	"gitlab.com/slon/shad-go/gitfame/internal/core"
	"gitlab.com/slon/shad-go/gitfame/internal/input_reader"
	"gitlab.com/slon/shad-go/gitfame/pkg/error_handling"
	"gitlab.com/slon/shad-go/gitfame/pkg/progress_bar"
)

func main() {
	pb := progress_bar.ProgressBar{Delta: 5}
	pb.SendMessage("Reading input...")
	info := input_reader.ParseFlags()
	err := input_reader.CheckInput(info)
	error_handling.CheckError(err)
	err = core.Execute(info, &pb)
	error_handling.CheckError(err)
}
