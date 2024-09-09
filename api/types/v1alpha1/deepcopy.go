package v1alpha1

import "k8s.io/apimachinery/pkg/runtime"

func (in *ActivePlugin) DeepCopyInto(out *ActivePlugin) {
	out.Name = in.Name
	out.Config = in.Config
}

func (in *Client) DeepCopyInto(out *Client) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = ClientSpec{
		Hostname:      in.Spec.Hostname,
		ClientID:      in.Spec.ClientID,
		SecretID:      in.Spec.SecretID,
		ForwardPorts:  in.Spec.ForwardPorts,
		Owner:         in.Spec.Owner,
		ServiceName:   in.Spec.ServiceName,
		ActivePlugins: make([]ActivePlugin, len(in.Spec.ActivePlugins)),
	}
	out.Status = ClientStatus{
		Connected:   in.Status.Connected,
		LastContact: in.Status.LastContact,
	}

	for i := range in.Spec.ActivePlugins {
		out.Spec.ActivePlugins[i].DeepCopyInto(&in.Spec.ActivePlugins[i])

	}
}

func (in *Client) DeepCopyObject() runtime.Object {
	out := Client{}
	in.DeepCopyInto(&out)

	return &out
}

func (in *ClientList) DeepCopyObject() runtime.Object {
	out := ClientList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]Client, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
