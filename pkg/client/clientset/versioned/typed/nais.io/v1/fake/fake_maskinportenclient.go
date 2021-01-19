// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	naisiov1 "github.com/nais/naiserator/pkg/apis/nais.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMaskinportenClients implements MaskinportenClientInterface
type FakeMaskinportenClients struct {
	Fake *FakeNaisV1
	ns   string
}

var maskinportenclientsResource = schema.GroupVersionResource{Group: "nais.io", Version: "v1", Resource: "maskinportenclients"}

var maskinportenclientsKind = schema.GroupVersionKind{Group: "nais.io", Version: "v1", Kind: "MaskinportenClient"}

// Get takes name of the maskinportenClient, and returns the corresponding maskinportenClient object, and an error if there is any.
func (c *FakeMaskinportenClients) Get(name string, options v1.GetOptions) (result *naisiov1.MaskinportenClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(maskinportenclientsResource, c.ns, name), &naisiov1.MaskinportenClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*naisiov1.MaskinportenClient), err
}

// List takes label and field selectors, and returns the list of MaskinportenClients that match those selectors.
func (c *FakeMaskinportenClients) List(opts v1.ListOptions) (result *naisiov1.MaskinportenClientList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(maskinportenclientsResource, maskinportenclientsKind, c.ns, opts), &naisiov1.MaskinportenClientList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &naisiov1.MaskinportenClientList{ListMeta: obj.(*naisiov1.MaskinportenClientList).ListMeta}
	for _, item := range obj.(*naisiov1.MaskinportenClientList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested maskinportenClients.
func (c *FakeMaskinportenClients) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(maskinportenclientsResource, c.ns, opts))

}

// Create takes the representation of a maskinportenClient and creates it.  Returns the server's representation of the maskinportenClient, and an error, if there is any.
func (c *FakeMaskinportenClients) Create(maskinportenClient *naisiov1.MaskinportenClient) (result *naisiov1.MaskinportenClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(maskinportenclientsResource, c.ns, maskinportenClient), &naisiov1.MaskinportenClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*naisiov1.MaskinportenClient), err
}

// Update takes the representation of a maskinportenClient and updates it. Returns the server's representation of the maskinportenClient, and an error, if there is any.
func (c *FakeMaskinportenClients) Update(maskinportenClient *naisiov1.MaskinportenClient) (result *naisiov1.MaskinportenClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(maskinportenclientsResource, c.ns, maskinportenClient), &naisiov1.MaskinportenClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*naisiov1.MaskinportenClient), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMaskinportenClients) UpdateStatus(maskinportenClient *naisiov1.MaskinportenClient) (*naisiov1.MaskinportenClient, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(maskinportenclientsResource, "status", c.ns, maskinportenClient), &naisiov1.MaskinportenClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*naisiov1.MaskinportenClient), err
}

// Delete takes name of the maskinportenClient and deletes it. Returns an error if one occurs.
func (c *FakeMaskinportenClients) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(maskinportenclientsResource, c.ns, name), &naisiov1.MaskinportenClient{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMaskinportenClients) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(maskinportenclientsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &naisiov1.MaskinportenClientList{})
	return err
}

// Patch applies the patch and returns the patched maskinportenClient.
func (c *FakeMaskinportenClients) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *naisiov1.MaskinportenClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(maskinportenclientsResource, c.ns, name, pt, data, subresources...), &naisiov1.MaskinportenClient{})

	if obj == nil {
		return nil, err
	}
	return obj.(*naisiov1.MaskinportenClient), err
}