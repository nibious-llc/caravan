package server

import (
	"github.com/nibious-llc/caravan/api/types/v1alpha1"
	client_v1alpha1 "github.com/nibious-llc/caravan/pkg/clientset/v1alpha1"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8s_auth "github.com/nibious-llc/caravan/internal/auth/k8s"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"time"
)

var (
	k8sCluster       *client_v1alpha1.V1Alpha1Client
	currentNamespace string
	clientStore      cache.Store
)

func ConnectToK8s() {
	// --- setup k8s connection ---

	// Get current cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Panic().Err(err).Msg("Could not get cluster config")
	}

	k8sCluster, err = client_v1alpha1.NewForConfig(config)
	if err != nil {
		log.Panic().Err(err).Msg("Could not create configuration to communicate with cluster")
	}

	currentNamespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Panic().Err(err).Msg("Could not determine current namespace from k8s mounts")
	}
	currentNamespace = string(currentNamespaceBytes)

	clientStore = WatchResources(k8sCluster)

	// Default setup the k8s auth provider
	ap := k8s_auth.K8sAuthProvider{
		ClientStore: clientStore,
	}
	AuthProvider = ap
}

func WatchResources(clientSet client_v1alpha1.V1Alpha1Interface) cache.Store {
	projectStore, projectController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return clientSet.Clients(currentNamespace).List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return clientSet.Clients(currentNamespace).Watch(lo)
			},
		},
		&v1alpha1.Client{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{},
	)

	go projectController.Run(wait.NeverStop)
	return projectStore
}
