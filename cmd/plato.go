package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	ConfigPath string
)

func init() {
	cobra.OnInitialize(initConfig)
	// 命令行没有提供--config参数时，使用默认值 ./plato.yaml作为config文件，后面viper会根据这个文件读取配置
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "./plato.yaml", "config file (default is ./plato.yaml)")
}

var rootCmd = &cobra.Command{
	Use:   "plato",
	Short: "IM系统实现连接层",
	Run:   Plato,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Plato(cmd *cobra.Command, args []string) {

}

func initConfig() {

}
