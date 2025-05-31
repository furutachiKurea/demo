package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

const (
	// namespace 强制使用 demo
	defaultNamespace = "demo"
)

var (
	// 子命令的 FlagSet
	createCmd *flag.FlagSet
	deleteCmd *flag.FlagSet
	scaleCmd  *flag.FlagSet

	// 子命令 flag
	createName    string // 需要创建的集群名称
	deleteName    string // 需要删除的集群名称
	scaleName     string // 需要扩缩容的集群名称
	scaleReplicas int    // 扩缩容集群时的目标副本数
)

func printUsage() {
	fmt.Printf(`%s - KubeBlocks Operator 命令行工具

使用方式:
  %s <command> [flags]

可用命令:
  create    创建一个单副本的 PostgreSQL 集群
  delete    删除指定的集群
  scale     调整集群副本数

示例:
  # 创建集群
  %s create --name op-test

  # 删除集群
  %s delete --name op-test

  # 扩缩容集群
  %s scale --name op-test --replicas 3
`,
		os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func init() {
	// 初始化 create 子命令
	createCmd = flag.NewFlagSet("create", flag.ExitOnError)
	createCmd.StringVarP(&createName, "name", "n", "", "要创建的集群名称 (必需)")

	// 初始化 delete 子命令
	deleteCmd = flag.NewFlagSet("delete", flag.ExitOnError)
	deleteCmd.StringVarP(&deleteName, "name", "n", "", "要删除的集群名称 (必需)")

	// 初始化 scale 子命令
	scaleCmd = flag.NewFlagSet("scale", flag.ExitOnError)
	scaleCmd.StringVarP(&scaleName, "name", "n", "", "要扩缩容的集群名称 (必需)")
	scaleCmd.IntVarP(&scaleReplicas, "replicas", "r", 0, "目标副本数 (必需且必须 > 0)")

	// 设置全局用法信息
	flag.Usage = printUsage
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subCmd := os.Args[1]

	switch subCmd {
	case "create":
		if err := createCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("解析参数失败: %v\n", err)
			os.Exit(1)
		}
		if createName == "" {
			fmt.Println("错误: --name 参数不能为空")
			createCmd.Usage()
			os.Exit(1)
		}
		if err := createCluster(createName); err != nil {
			fmt.Printf("创建集群失败: %v\n", err)
			os.Exit(1)
		}

	case "delete":
		if err := deleteCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("解析参数失败: %v\n", err)
			os.Exit(1)
		}
		if deleteName == "" {
			fmt.Println("错误: --name 参数不能为空")
			deleteCmd.Usage()
			os.Exit(1)
		}
		if err := deleteCluster(deleteName); err != nil {
			fmt.Printf("删除集群失败: %v\n", err)
			os.Exit(1)
		}

	case "scale":
		if err := scaleCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("解析参数失败: %v\n", err)
			os.Exit(1)
		}
		if scaleName == "" || scaleReplicas <= 0 {
			fmt.Println("错误: --name 不能为空且 --replicas 必须大于 0")
			scaleCmd.Usage()
			os.Exit(1)
		}
		if err := scaleCluster(scaleName, scaleReplicas); err != nil {
			fmt.Printf("扩缩容集群失败: %v\n", err)
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}
