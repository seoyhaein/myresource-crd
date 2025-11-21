package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	myv1alpha1 "github.com/seoyhaein/myresource-crd/pkg/apis/mygroup.example.com/v1alpha1"
	myclientset "github.com/seoyhaein/myresource-crd/pkg/clientset/clientset"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// 소스는 일단 잘 모름.

func main() {
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file (optional)")
	flag.Parse()

	// 1) kubeconfig 로드 (파라미터 없으면 ~/.kube/config 사용)
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		log.Fatalf("failed to load kubeconfig: %v", err)
	}

	// 2) 생성된 clientset 만들기
	cs, err := myclientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	// 3) ListWatch 정의 (MyResource 전체 네임스페이스)
	resyncPeriod := 30 * time.Second

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			// 처음 시작할 때 전체 리스트
			return cs.MygroupV1alpha1().MyResources(metav1.NamespaceAll).List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			// 이후 변경 사항 watch
			return cs.MygroupV1alpha1().MyResources(metav1.NamespaceAll).Watch(context.Background(), options)
		},
	}

	// 4) SharedIndexInformer 생성
	informer := cache.NewSharedIndexInformer(
		lw,
		&myv1alpha1.MyResource{}, // 우리가 정의한 타입
		resyncPeriod,
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		},
	)

	// 5) 이벤트 핸들러 등록 (Add / Update / Delete)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mr := obj.(*myv1alpha1.MyResource)
			log.Printf("[ADD]    %s/%s image=%v memory=%v\n",
				mr.Namespace,
				mr.Name,
				mr.Spec.Image,  // types.go 에 Spec 정의했다면 이렇게 접근 가능
				mr.Spec.Memory, // 없으면 그냥 mr.Spec 써도 됨
			)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldMr := oldObj.(*myv1alpha1.MyResource)
			newMr := newObj.(*myv1alpha1.MyResource)
			log.Printf("[UPDATE] %s/%s -> generation %d -> %d\n",
				newMr.Namespace,
				newMr.Name,
				oldMr.Generation,
				newMr.Generation,
			)
		},
		DeleteFunc: func(obj interface{}) {
			mr, ok := obj.(*myv1alpha1.MyResource)
			if !ok {
				// tombstone 타입으로 들어오는 경우도 있어서 방어 코드
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if ok {
					if mr2, ok2 := tombstone.Obj.(*myv1alpha1.MyResource); ok2 {
						mr = mr2
					}
				}
			}
			if mr != nil {
				log.Printf("[DELETE] %s/%s\n", mr.Namespace, mr.Name)
			} else {
				log.Printf("[DELETE] unknown object: %#v\n", obj)
			}
		},
	})

	// 6) 종료 시그널 처리용 stop 채널
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("Received shutdown signal")
		close(stopCh)
	}()

	log.Println("Starting MyResource controller (client-go + informer) ...")
	// 7) 인포머 실행 (이 안에서 List + Watch + 캐시 유지)
	informer.Run(stopCh)
	log.Println("MyResource controller stopped")
}
