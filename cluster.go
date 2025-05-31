package main

import (
	"context"
	"fmt"
	"time"

	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	opsv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	clusterDef      = "postgresql"
	postgresqlCompName = "postgresql"
)

// createCluster 创建一个新的 PostgreSQL 集群，集群名为 name
func createCluster(name string) error {
	// 30 分钟超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	k8sClient, err := NewKubeBlocksClient(defaultNamespace)
	if err != nil {
		return fmt.Errorf("初始化 K8s 客户端失败: %w", err)
	}

	// TODO 由用户自定义
	// 设置终止策略为 Delete
	terminationPolicy := appsv1.WipeOut

	// TODO 由用户自定义
	disableExporter := false

	/*
		对应 yaml
		apiVersion: apps.kubeblocks.io/v1
		kind: Cluster
		metadata:
		  name: ${name}
		  namespace: demo
		spec:
		  terminationPolicy: WipeOut
		  clusterDef: postgresql
		  topology: replication
		  componentSpecs:
		    - name: postgresql
		      serviceVersion: "14.7.2"
		      disableExporter: false
		      labels:
		        apps.kubeblocks.postgres.patroni/scope: ${name}-postgresql
		      replicas: 1
		      resources:
		        requests:
		          cpu: "0.5"
		          memory: "1Gi"
		        limits:
		          cpu: "1"
		          memory: "1Gi"
		      volumeClaimTemplates:
		        - name: data
		          spec:
		            accessModes:
		              - ReadWriteOnce
		            resources:
		              requests:
		                storage: 10Gi
	*/
	cluster := &appsv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.kubeblocks.io/v1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNamespace,
		},
		Spec: appsv1.ClusterSpec{
			TerminationPolicy: terminationPolicy,
			ClusterDef:        clusterDef,
			Topology:          "replication",
			ComponentSpecs: []appsv1.ClusterComponentSpec{
				{
					Name:            postgresqlCompName,
					ServiceVersion:  "14.7.2",
					DisableExporter: &disableExporter,
					Labels: map[string]string{
						"apps.kubeblocks.postgres.patroni/scope": fmt.Sprintf("%s-postgresql", name),
					},
					Replicas: 1,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("0.5"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
					VolumeClaimTemplates: []appsv1.ClusterComponentVolumeClaimTemplate{
						{
							Name: "data",
							Spec: appsv1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									corev1.ReadWriteOnce,
								},
								Resources: corev1.VolumeResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("10Gi"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := k8sClient.Create(ctx, cluster); err != nil {
		return fmt.Errorf("创建集群失败: %w", err)
	}

	fmt.Printf("集群 %s 创建请求已提交，等待就绪...\n", name)

	// 等待集群就绪
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("等待集群就绪超时: %w", ctx.Err())
		case <-ticker.C:
			// 获取集群最新状态
			currentCluster := &appsv1.Cluster{}
			if err := k8sClient.Get(ctx, client.ObjectKey{
				Name:      name,
				Namespace: defaultNamespace,
			}, currentCluster); err != nil {
				return fmt.Errorf("获取集群状态失败: %w", err)
			}

			// 检查集群状态
			if currentCluster.Status.Phase == "Running" {
				fmt.Printf("集群 %s 已成功启动并运行\n", name)
				return nil
			}

			// 如果出现问题状态，提前返回错误
			if currentCluster.Status.Phase == "Failed" || currentCluster.Status.Phase == "Abnormal" {
				return fmt.Errorf("集群 %s 启动失败，当前状态: %s", name, currentCluster.Status.Phase)
			}

			fmt.Printf("等待集群 %s 就绪，当前状态: %s\n", name, currentCluster.Status.Phase)
		}
	}
}

// deleteCluster 删除指定的集群
func deleteCluster(name string) error {
	// 创建一个带超时的 context（5分钟）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	k8sClient, err := NewKubeBlocksClient(defaultNamespace)
	if err != nil {
		return fmt.Errorf("初始化 K8s 客户端失败: %w", err)
	}

	cluster := &appsv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNamespace,
		},
	}

	// 发送删除请求
	if err := k8sClient.Delete(ctx, cluster); err != nil {
		return fmt.Errorf("删除集群失败: %w", err)
	}

	fmt.Printf("集群 %s 删除中...\n", name)

	// 等待集群完全删除
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("集群删除超时: %w", ctx.Err())
		case <-ticker.C:
			err := k8sClient.Get(ctx, client.ObjectKey{
				Name:      name,
				Namespace: defaultNamespace,
			}, cluster)

			if err != nil {
				if k8serrors.IsNotFound(err) {
					fmt.Printf("集群 %s 已成功删除\n", name)
					return nil
				}
				return fmt.Errorf("检查集群状态失败: %w", err)
			}
		}
	}
}

// scaleCluster 调整集群的副本数
func scaleCluster(name string, replicas int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	k8sClient, err := NewKubeBlocksClient(defaultNamespace)
	if err != nil {
		return fmt.Errorf("初始化 K8s 客户端失败: %w", err)
	}

	// 获取当前集群
	cluster := &appsv1.Cluster{}
	if err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: defaultNamespace,
	}, cluster); err != nil {
		return fmt.Errorf("获取集群失败: %w", err)
	}

	// 获取当前副本数
	currentReplicas := int32(0)
	for _, comp := range cluster.Spec.ComponentSpecs {
		if comp.Name == postgresqlCompName {
			currentReplicas = comp.Replicas
			break
		}
	}

	targetReplicas := int32(replicas)
	if targetReplicas == currentReplicas {
		fmt.Printf("集群 %s 的副本数已经是 %d，无需调整\n", name, targetReplicas)
		return nil
	}

	// 计算需要调整的副本数
	replicaChanges := targetReplicas - currentReplicas

	// 创建 OpsRequest 进行扩缩容
	opsRequest := &opsv1alpha1.OpsRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operations.kubeblocks.io/v1",
			Kind:       "OpsRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-scale-", name),
			Namespace:    defaultNamespace,
		},
		Spec: opsv1alpha1.OpsRequestSpec{
			ClusterName: name,
		},
	}

	// 设置操作类型为水平扩缩容
	opsRequest.Spec.Type = opsv1alpha1.HorizontalScalingType

	if replicaChanges > 0 {
		// 扩容操作
		opsRequest.Spec.HorizontalScalingList = []opsv1alpha1.HorizontalScaling{
			{
				ComponentOps: opsv1alpha1.ComponentOps{
					ComponentName: postgresqlCompName,
				},
				ScaleOut: &opsv1alpha1.ScaleOut{
					ReplicaChanger: opsv1alpha1.ReplicaChanger{
						ReplicaChanges: &replicaChanges,
					},
				},
			},
		}
	} else {
		// 缩容操作
		replicaChanges = -replicaChanges 
		opsRequest.Spec.HorizontalScalingList = []opsv1alpha1.HorizontalScaling{
			{
				ComponentOps: opsv1alpha1.ComponentOps{
					ComponentName: postgresqlCompName,
				},
				ScaleIn: &opsv1alpha1.ScaleIn{
					ReplicaChanger: opsv1alpha1.ReplicaChanger{
						ReplicaChanges: &replicaChanges,
					},
				},
			},
		}
	}

	// 创建 OpsRequest
	if err := k8sClient.Create(ctx, opsRequest); err != nil {
		return fmt.Errorf("创建扩缩容请求失败: %w", err)
	}

	fmt.Printf("已提交扩缩容请求，将 %s 集群的 %s 组件副本数从 %d 调整为 %d\n", 
		name, postgresqlCompName, currentReplicas, replicas)
	
	// 等待操作完成
	fmt.Println("正在等待操作完成...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	opsRequestName := opsRequest.Name
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("等待扩缩容操作超时: %w", ctx.Err())
		case <-ticker.C:
			currentOps := &opsv1alpha1.OpsRequest{}
			if err := k8sClient.Get(ctx, client.ObjectKey{
				Name:      opsRequestName,
				Namespace: defaultNamespace,
			}, currentOps); err != nil {
				return fmt.Errorf("获取操作状态失败: %w", err)
			}

			// 检查操作状态
			switch currentOps.Status.Phase {
			case "Succeed":
				fmt.Println("扩缩容操作成功完成")
				return nil
			case "Failed", "Cancelled", "Aborted":
				return fmt.Errorf("扩缩容操作失败或已取消，状态: %s", currentOps.Status.Phase)
			default:
				fmt.Printf("操作正在执行中，当前状态: %s\n", currentOps.Status.Phase)
			}
		}
	}
}
