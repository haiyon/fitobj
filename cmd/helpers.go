package cmd

import (
	"github.com/haiyon/fitobj/fitter"
	"github.com/haiyon/fitobj/processor"
	"github.com/spf13/viper"
)

func buildProcessorOptions() processor.Options {
	return processor.Options{
		Workers:       getWorkers(),
		FlattenOpts:   buildFlattenOptions(),
		UnflattenOpts: buildUnflattenOptions(),
	}
}

func buildFlattenOptions() fitter.FlattenOptions {
	opts := fitter.DefaultFlattenOptions()
	opts.Separator = getSeparator()
	opts.ArrayFormatting = getArrayFormat()
	opts.BufferSize = getBufferSize()
	return opts
}

func buildUnflattenOptions() fitter.UnflattenOptions {
	opts := fitter.DefaultUnflattenOptions()
	opts.Separator = getSeparator()
	opts.SupportBracketNotation = getArrayFormat() == "bracket"
	opts.BufferSize = getBufferSize()
	return opts
}

func getSeparator() string {
	return viper.GetString("separator")
}

func getArrayFormat() string {
	format := viper.GetString("array-format")
	if format != "index" && format != "bracket" {
		return "index" // default fallback
	}
	return format
}

func getWorkers() int {
	workers := viper.GetInt("workers")
	if workers <= 0 {
		workers = 4 // fallback default
	}
	return workers
}

func getBufferSize() int {
	size := viper.GetInt("buffer")
	if size <= 0 {
		size = 16 // fallback default
	}
	return size
}
