package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCMD = &cobra.Command{
		Use:   "http-buffer-agent",
		Short: "is a http agent with buffer",
		Long:  "keep your http request in buffer and do request util success",
	}
)
