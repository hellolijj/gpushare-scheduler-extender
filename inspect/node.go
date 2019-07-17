package main

import (
	"log"
	"os"
	"path"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
)

var (
	clientConfig clientcmd.ClientConfig
	clientset    *kubernetes.Clientset
	restConfig   *rest.Config
	retries      = 5
)

func kubeInit() {

	kubeconfigFile := os.Getenv("KUBECONFIG")
	if kubeconfigFile == "" {
		kubeconfigFile = path.Join(os.Getenv("HOME"), "/.kube/config")
	}
	if _, err := os.Stat(kubeconfigFile); err != nil {
		log.Fatalf("kubeconfig %s failed to find due to %v, please set KUBECONFIG env", kubeconfigFile, err)
	}

	var err error
	restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		log.Fatalf("Failed due to %v", err)
	}
	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed due to %v", err)
	}
}

func getAllGPUNode() ([]v1.Node, error) {
	nodes := []v1.Node{}
	allNodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nodes, err
	}
	for _, node := range allNodes.Items {
		if utils.IsGPUTopologyNode(&node) {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func getNodes(nodeName string) ([]v1.Node, error) {
	node, err := clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	return []v1.Node{*node}, err
}