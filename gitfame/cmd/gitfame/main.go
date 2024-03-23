//go:build !solution

package main

import (
	"gitlab.com/slon/shad-go/gitfame/internal/core"
	"gitlab.com/slon/shad-go/gitfame/internal/inputreader"
	"gitlab.com/slon/shad-go/gitfame/pkg/errorhandling"
	"gitlab.com/slon/shad-go/gitfame/pkg/progressbar"
)

func main() {
	pb := progressbar.ProgressBar{Delta: 5}
	pb.SendMessage("Reading input...")
	info := inputreader.ParseFlags()
	err := inputreader.CheckInput(info)
	errorhandling.CheckError(err)
	err = core.Execute(info, &pb)
	errorhandling.CheckError(err)
}
