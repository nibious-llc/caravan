package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ActivePlugin struct {
	Name   string `json:"name"`
	Config string `json:"config"`
}

type ClientSpec struct {
	//type everything out there!
	Hostname      string         `json:"hostname"`
	ClientID      string         `json:"clientID"`
	SecretID      string         `json:"secretID"`
	ForwardPorts  []int          `json:"forwardPorts"`
	Owner         string         `json:"owner"`
	ServiceName   string         `json:"serviceName"`
	ActivePlugins []ActivePlugin `json:"activePlugins"`
}

type ClientStatus struct {
	Connected   bool        `json:"connected"`
	LastContact metav1.Time `json:"lastContact"`
}

type Client struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClientSpec   `json:"spec"`
	Status ClientStatus `json:"status"`
}

type ClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Client `json:"items"`
}
