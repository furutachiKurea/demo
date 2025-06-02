使用 controller-runtime 实现的简单命令行工具，在本地的 k8s 集群中实现 Kubeblocks cluster 的创建、扩缩容和删除

作为 rainbond 集成 kubeblocks 的部分技术验证

# 开发环境

- Ubuntu 24.04 LTS
- kind version 0.24.0
- Docker version 28.2.1
- Kubernetes: v1.31.0
- KubeBlocks: 1.0.0
- kbcli: 1.0.0
- kubectl: v1.33.1

(It works on my machine XD)

# Kubeblocks 安装

## kind

创建一个三节点的集群

```shell
kind create cluster --name multinode-cluster --config kind_config.yaml
```

安装 Snapshot Controller, KubeBlocks DataProtection 控制器使用 Snapshot Controller 来为数据库创建快照备份

```shell
helm repo add piraeus-charts https://piraeus.io/helm-charts/
helm repo update

helm install snapshot-controller piraeus-charts/snapshot-controller -n kb-system --create-namespace 

kubectl get pods -n kb-system | grep snapshot-controller # 验证安装状态
```

使用 kbcli 安装 Kubeblocks，其他安装方法参考 [Kubeblocks 官方文档](https://cn.kubeblocks.io/docs/preview/user-docs/installation/install-kubeblocks/)

```shell
kbcli kubeblocks install 

kbcli kubeblocks status # 验证 Kubeblocks 安装状态
```

# 使用

使用 make 构建

```shell
make
```

create 命令将会创建一个单副本的 PostgreSQL 集群

```shell
kbop create --name test-cluster
```

scale 命令用来修改指定集群的副本数

```shell
kbop scale --name test-cluster --replicas 2
```

delete 命令用来删除指定的集群

```shell
kbop delete --name test-cluster
```