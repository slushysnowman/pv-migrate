package pvc

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Info interface {
	KubeClient() kubernetes.Interface
	Claim() *corev1.PersistentVolumeClaim
	MountedNode() string
	SupportsRWO() bool
	SupportsROX() bool
	SupportsRWX() bool
}

type info struct {
	kubeClient  kubernetes.Interface
	claim       *corev1.PersistentVolumeClaim
	mountedNode string
	supportsRWO bool
	supportsROX bool
	supportsRWX bool
}

func (p *info) KubeClient() kubernetes.Interface {
	return p.kubeClient
}

func (p *info) Claim() *corev1.PersistentVolumeClaim {
	return p.claim
}

func (p *info) MountedNode() string {
	return p.mountedNode
}

func (p *info) SupportsRWO() bool {
	return p.supportsRWO
}

func (p *info) SupportsROX() bool {
	return p.supportsROX
}

func (p *info) SupportsRWX() bool {
	return p.supportsRWX
}

func New(kubeClient kubernetes.Interface, namespace string, name string) (Info, error) {
	claim, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	mountedNode, err := findMountedNode(kubeClient, claim)
	if err != nil {
		return nil, err
	}

	supportsRWO := false
	supportsROX := false
	supportsRWX := false

	for _, accessMode := range claim.Spec.AccessModes {
		switch accessMode {
		case corev1.ReadWriteOnce:
			supportsRWO = true
		case corev1.ReadOnlyMany:
			supportsROX = true
		case corev1.ReadWriteMany:
			supportsRWX = true
		}
	}

	return &info{
		kubeClient:  kubeClient,
		claim:       claim,
		mountedNode: mountedNode,
		supportsRWO: supportsRWO,
		supportsROX: supportsROX,
		supportsRWX: supportsRWX,
	}, nil
}

func findMountedNode(kubeClient kubernetes.Interface, pvc *corev1.PersistentVolumeClaim) (string, error) {
	podList, err := kubeClient.CoreV1().Pods(pvc.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, pod := range podList.Items {
		for _, volume := range pod.Spec.Volumes {
			persistentVolumeClaim := volume.PersistentVolumeClaim
			if persistentVolumeClaim != nil && persistentVolumeClaim.ClaimName == pvc.Name {
				return pod.Spec.NodeName, nil
			}
		}
	}

	return "", errors.New("couldn't find the node that the pvc is mounted to")
}