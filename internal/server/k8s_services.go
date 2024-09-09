package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	c "github.com/nibious-llc/caravan/internal/common"
	"github.com/rs/zerolog/log"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func create_service(client *c.ClientHub) {

	log.Info().Msg(fmt.Sprintf("[create_service][%s] Creating Service for client", client.LoginID.String()))

	// Create our service ports
	var ports []apiv1.ServicePort

	for _, tunnel := range client.Tunnels {

		port := apiv1.ServicePort{
			Name:       tunnel.SessionID,
			Port:       int32(tunnel.Port),                  //The port the service on the client is at
			TargetPort: intstr.FromInt(tunnel.ListenerPort), //The port on this server instance
		}
		ports = append(ports, port)
	}

	// Add our prometheus port
	ports = append(ports, apiv1.ServicePort{
		Name:       "metrics",
		Port:       int32(client.MetricsPort),
		TargetPort: intstr.FromInt(client.MetricsPort),
	})

	// Create our service description
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", strings.ToLower(client.Namespace), strings.ToLower(client.Hostname)), //  Needs to be the name of the
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   strconv.Itoa(client.MetricsPort),
			},
			Labels: map[string]string{
				"nibious.com/app": "caravan",
				"client":          client.Namespace,
				"hostname":        client.Hostname,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"nibious.com/app": "caravan",
			},
			Ports: ports,
		},
	}

	// Now ask for our service to be created
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error().Err(err).Msg("Could not find config")
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error().Err(err).Msg("Could not create config object")
		return
	}

	_, err = clientset.CoreV1().Services(currentNamespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Could not create service -- trying again")
		delete_service(client)
		time.Sleep(3 * time.Second)
		create_service(client)
		return
	}

	update_client_status_connected(client, true)

}

func update_client_status_connected(client *c.ClientHub, status bool) {

	// Update the status of the item (connected and date)
	clientRecord, err := k8sCluster.Clients(currentNamespace).Get(fmt.Sprintf("%s-%s", strings.ToLower(client.Namespace), strings.ToLower(client.Hostname)), metav1.GetOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Could not retrieve client record from server. Likely it is not named properly")
		return
	}

	clientRecord.Status.Connected = status

	_, err = k8sCluster.Clients(currentNamespace).UpdateStatus(clientRecord, metav1.UpdateOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Could not update client status")
		return
	}
}

func delete_service(client *c.ClientHub) {

	// Now ask for our service to be created
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error().Err(err).Msg("Could not find config")
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error().Err(err).Msg("Could not create config object")
	}

	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.CoreV1().Services(currentNamespace).Delete(context.TODO(), fmt.Sprintf("%s-%s", strings.ToLower(client.Namespace), strings.ToLower(client.Hostname)), metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Error().Err(err).Msg("Could not delete service")
	}

	update_client_status_connected(client, false)
}
