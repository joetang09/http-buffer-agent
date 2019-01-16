package cmd

import (
	"github.com/joetang09/http-buffer-agent/config"
	"github.com/joetang09/http-buffer-agent/server"
	"github.com/spf13/cobra"
)

func init() {
	serveCMD.Flags().Uint("retrytimes", config.DefaultRetryTimes, "every request retrytimes")
	serveCMD.Flags().Uint("bufferlength", config.DefaultBufferLen, "buffer length")
	serveCMD.Flags().Uint("port", config.DefaultPort, "port")
	serveCMD.Flags().Uint("outparallel", config.DefaultOutParallel, "out parallel number")

	config.GetViper().BindPFlag("retrytimes", serveCMD.Flags().Lookup("retrytimes"))
	config.GetViper().BindPFlag("bufferlength", serveCMD.Flags().Lookup("bufferlength"))
	config.GetViper().BindPFlag("port", serveCMD.Flags().Lookup("port"))
	config.GetViper().BindPFlag("outparallel", serveCMD.Flags().Lookup("outparallel"))

	rootCMD.AddCommand(serveCMD)
}

var (
	serveCMD = &cobra.Command{
		Use:   "serve",
		Short: "start server",
		Long:  "start http buffer agent server",
		RunE: func(c *cobra.Command, args []string) error {
			server.Serve()
			return nil
		},
	}
)

/**

1. 请求服务的地址
2. 转发规则列表



*/
