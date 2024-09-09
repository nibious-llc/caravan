package v1alpha1

import (
	"context"
	"github.com/nibious-llc/caravan/api/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type ClientInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.ClientList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.Client, error)
	Create(*v1alpha1.Client) (*v1alpha1.Client, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	UpdateStatus(client *v1alpha1.Client, opts metav1.UpdateOptions) (*v1alpha1.Client, error)
	// ...
}

type clientClient struct {
	restClient rest.Interface
	ns         string
}

func (c *clientClient) List(opts metav1.ListOptions) (*v1alpha1.ClientList, error) {
	result := v1alpha1.ClientList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("clients").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *clientClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.Client, error) {
	result := v1alpha1.Client{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("clients").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *clientClient) Create(client *v1alpha1.Client) (*v1alpha1.Client, error) {
	result := v1alpha1.Client{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("clients").
		Body(client).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *clientClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("clients").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.TODO())
}

func (c *clientClient) UpdateStatus(client *v1alpha1.Client, opts metav1.UpdateOptions) (*v1alpha1.Client, error) {
	result := v1alpha1.Client{}
	err := c.restClient.Put().
		Namespace(c.ns).
		Resource("clients").
		Name(client.GetName()).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(client).
		Do(context.TODO()).
		Into(&result)
	return &result, err
}
