package main

import (
	"flag"
	"fmt"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kc *string
	if home := homedir.HomeDir(); home != "" {
		kc = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "optional absolute path to the kubeconfig file")
	} else {
		kc = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kc)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	podsClient := clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	pods, err := podsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, po := range pods.Items {
		fmt.Printf("%s, %s, %s\n", po.GetName(), po.Spec.NodeName, po.Spec.Containers)
	}

}
