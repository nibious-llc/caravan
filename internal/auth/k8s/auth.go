package k8s

import (
	"github.com/google/uuid"
	"github.com/nibious-llc/caravan/api/types/v1alpha1"
	c "github.com/nibious-llc/caravan/internal/common"
	"k8s.io/client-go/tools/cache"
)

type K8sAuthProvider struct {
	ClientStore cache.Store
}

func (ap K8sAuthProvider) IsLoginValid(clientID uuid.UUID, secret string) (c.IunctioClient, bool) {

	if ap.ClientStore == nil {
		panic("Authentication client store not set. Bad programming")
	}

	var db_login_data c.IunctioClient

	// Search the watcher for all client IDs and secretID matches
	for _, element := range ap.ClientStore.List() {

		var client = element.(*v1alpha1.Client)

		clientID_parsed, err := uuid.Parse(client.Spec.ClientID)
		if err != nil {
			return db_login_data, false
		}

		if clientID_parsed == clientID && client.Spec.SecretID == secret {
			db_login_data.Namespace = client.Spec.Owner
			db_login_data.Hostname = client.Spec.Hostname
			db_login_data.ClientID = clientID

			return db_login_data, true

		}

	}

	return db_login_data, false

}
