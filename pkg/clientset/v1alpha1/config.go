package v1alpha1

import (
	"github.com/nibious-llc/caravan/api/types/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type V1Alpha1Interface interface {
	Clients(namespace string) ClientInterface
}

type V1Alpha1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*V1Alpha1Client, error) {

	v1alpha1.AddToScheme(scheme.Scheme)

	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupName, Version: v1alpha1.GroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &V1Alpha1Client{restClient: client}, nil
}

func (c *V1Alpha1Client) Clients(namespace string) ClientInterface {
	return &clientClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
