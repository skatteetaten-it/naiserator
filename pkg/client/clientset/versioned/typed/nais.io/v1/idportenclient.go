// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"time"

	v1 "github.com/nais/naiserator/pkg/apis/nais.io/v1"
	scheme "github.com/nais/naiserator/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// IDPortenClientsGetter has a method to return a IDPortenClientInterface.
// A group's client should implement this interface.
type IDPortenClientsGetter interface {
	IDPortenClients(namespace string) IDPortenClientInterface
}

// IDPortenClientInterface has methods to work with IDPortenClient resources.
type IDPortenClientInterface interface {
	Create(*v1.IDPortenClient) (*v1.IDPortenClient, error)
	Update(*v1.IDPortenClient) (*v1.IDPortenClient, error)
	UpdateStatus(*v1.IDPortenClient) (*v1.IDPortenClient, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.IDPortenClient, error)
	List(opts metav1.ListOptions) (*v1.IDPortenClientList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.IDPortenClient, err error)
	IDPortenClientExpansion
}

// iDPortenClients implements IDPortenClientInterface
type iDPortenClients struct {
	client rest.Interface
	ns     string
}

// newIDPortenClients returns a IDPortenClients
func newIDPortenClients(c *NaisV1Client, namespace string) *iDPortenClients {
	return &iDPortenClients{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the iDPortenClient, and returns the corresponding iDPortenClient object, and an error if there is any.
func (c *iDPortenClients) Get(name string, options metav1.GetOptions) (result *v1.IDPortenClient, err error) {
	result = &v1.IDPortenClient{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("idportenclients").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of IDPortenClients that match those selectors.
func (c *iDPortenClients) List(opts metav1.ListOptions) (result *v1.IDPortenClientList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.IDPortenClientList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("idportenclients").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested iDPortenClients.
func (c *iDPortenClients) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("idportenclients").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a iDPortenClient and creates it.  Returns the server's representation of the iDPortenClient, and an error, if there is any.
func (c *iDPortenClients) Create(iDPortenClient *v1.IDPortenClient) (result *v1.IDPortenClient, err error) {
	result = &v1.IDPortenClient{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("idportenclients").
		Body(iDPortenClient).
		Do().
		Into(result)
	return
}

// Update takes the representation of a iDPortenClient and updates it. Returns the server's representation of the iDPortenClient, and an error, if there is any.
func (c *iDPortenClients) Update(iDPortenClient *v1.IDPortenClient) (result *v1.IDPortenClient, err error) {
	result = &v1.IDPortenClient{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("idportenclients").
		Name(iDPortenClient.Name).
		Body(iDPortenClient).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *iDPortenClients) UpdateStatus(iDPortenClient *v1.IDPortenClient) (result *v1.IDPortenClient, err error) {
	result = &v1.IDPortenClient{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("idportenclients").
		Name(iDPortenClient.Name).
		SubResource("status").
		Body(iDPortenClient).
		Do().
		Into(result)
	return
}

// Delete takes name of the iDPortenClient and deletes it. Returns an error if one occurs.
func (c *iDPortenClients) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("idportenclients").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *iDPortenClients) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("idportenclients").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched iDPortenClient.
func (c *iDPortenClients) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.IDPortenClient, err error) {
	result = &v1.IDPortenClient{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("idportenclients").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}