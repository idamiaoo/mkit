package new

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Command .
func Command() *cobra.Command {
	return initCmd
}

var (
	httpPort  = 9900
	grpcPort  = 9901
	modPrefix = "github.com/pescaria/"
	directory = "."

	initCmd = &cobra.Command{
		Use:   "new 项目名称",
		Short: "初始化一个 golang 项目",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 || strings.HasPrefix(args[0], "-") {
				_ = cmd.Help()
				os.Exit(1)
			}
			absPath, err := initProject(args)
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
			fmt.Printf("项目初始化完成 [%s]", absPath)
		},
	}
)

func init() {
	initCmd.Flags().StringVar(&modPrefix, "module-prefix", modPrefix, "go module 包名前缀")
	initCmd.Flags().IntVar(&httpPort, "http", httpPort, "指定项目http监听端口")
	initCmd.Flags().IntVar(&grpcPort, "grpc", grpcPort, "指定项目grpc监听端口")
	initCmd.Flags().StringVar(&directory, "d", directory, "指定项目所在目录")
}

func initProject(args []string) (string, error) {
	p := &project{
		Name:         args[0],
		ModPrefix:    modPrefix,
		HttpPort:     httpPort,
		GrpcPort:     grpcPort,
		AbsolutePath: directory,
	}

	if p.AbsolutePath == "." {
		pwd, _ := os.Getwd()
		p.AbsolutePath = filepath.Join(pwd, p.Name)
	} else {
		absPath, err := filepath.Abs(p.AbsolutePath)
		if err != nil {
			return "", err
		}
		p.AbsolutePath = filepath.Join(absPath, p.Name)
	}

	if err := p.create(); err != nil {
		return "", err
	}

	return p.AbsolutePath, nil
}
