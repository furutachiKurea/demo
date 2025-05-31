package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	opsv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
)

// KubeBlocksClient 封装了 KubeBlocks 相关的 K8s 客户端功能
type KubeBlocksClient struct {
	client.Client
	namespace string
}

// NewKubeBlocksClient 创建并返回一个新的 KubeBlocks 客户端，如果命名空间不存在则创建
func NewKubeBlocksClient(namespace string) (*KubeBlocksClient, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("获取 kubeconfig 失败: %w", err)
	}

	scheme := runtime.NewScheme()
	// 添加 API 到 scheme
	if err := kbappsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("添加 KubeBlocks scheme 失败: %w", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("添加 corev1 scheme 失败: %w", err)
	}
	if err := opsv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("添加 OpsRequest scheme 失败: %w", err)
	}

	cl, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %w", err)
	}

	// 创建命名空间（如果不存在）
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	if err := cl.Create(context.Background(), ns); err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("创建命名空间 %s 失败: %w", namespace, err)
	}

	return &KubeBlocksClient{
		Client:    cl,
		namespace: namespace,
	}, nil
}

// GetNamespace 返回客户端配置的命名空间
func (c *KubeBlocksClient) GetNamespace() string {
	return c.namespace
}

// Create 创建资源
func (c *KubeBlocksClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return c.Client.Create(ctx, obj, opts...)
}

// Get 获取资源
func (c *KubeBlocksClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return c.Client.Get(ctx, key, obj)
}

// Update 更新资源
func (c *KubeBlocksClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return c.Client.Update(ctx, obj, opts...)
}

// Delete 删除资源
func (c *KubeBlocksClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return c.Client.Delete(ctx, obj, opts...)
}

// List 列出资源
func (c *KubeBlocksClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.Client.List(ctx, list, opts...)
}
