package main

import (
	"context"
	"fmt"
	"log"

	myclientset "github.com/seoyhaein/myresource-crd/pkg/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		log.Fatalf("failed to load kubeconfig: %v", err)
	}

	cs, err := myclientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	ctx := context.Background()
	
	list, err := cs.MygroupV1alpha1().MyResources("default").List(ctx, metav1.ListOptions{})

	if err != nil {
		log.Fatalf("failed to list MyResources: %v", err)
	}

	if len(list.Items) == 0 {
		fmt.Println("no MyResources found in default namespace")
		return
	}

	fmt.Println("MyResources in default namespace:")
	for _, res := range list.Items {
		fmt.Println(" -", res.GetName())
	}
}
