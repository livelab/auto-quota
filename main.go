package main

import (
	"flag"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	log "k8s.io/klog"
)

const NS = "auto-ns-test"

var clientset *kubernetes.Clientset

func main() {
	log.InitFlags(nil)
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

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	podsClient := clientset.CoreV1().Pods("test")
	pods, err := podsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	log.Infoln("Listing pods:")
	for _, po := range pods.Items {
		log.Infoln("%s, %s, %s\n", po.GetName(), po.Spec.NodeName, po.Status)
	}

	nsClient := clientset.CoreV1().Namespaces()
	nss, err := nsClient.List(metav1.ListOptions{})
	for _, ns := range nss.Items {
		log.Infoln(ns.GetName())
	}

	factory := informers.NewSharedInformerFactory(clientset, 1*time.Second)
	informer := factory.Core().V1().Namespaces().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nsCreated,
	})

	informer.Run(stopper)

	<-make(chan struct{})
}

func nsCreated(obj interface{}) {
	ns := obj.(*v1.Namespace)
	log.Infoln("Namespace created:", ns.GetName())
	if ns.GetName() == NS {
		log.Infoln("Create ResourceQuota for ", NS)
		rqClient := clientset.CoreV1().ResourceQuotas(NS)
		rqSpec := &v1.ResourceQuota{}
		rqSpec.SetName("compute-resource-quota")
		rqSpec.Spec = v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("100000"),
				v1.ResourceMemory: resource.MustParse("1Pi"),
				"limits.cpu":      resource.MustParse("100000"),
				"limits.memory":   resource.MustParse("1Pi"),
				"requests.cpu":    resource.MustParse("100000"),
				"requests.memory": resource.MustParse("1Pi"),
			},
		}
		rqClient.Create(rqSpec)
	}
}
