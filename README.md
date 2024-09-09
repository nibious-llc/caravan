# Caravan
A Kubernetes-native remote tunneling solution for system administrations 

## Table of Contents

- [Why](#why)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)


## Why

Having a Kubernetes-native remote tunneling solution enables other services, like Ansible AWX, Netbox, and more to be better integrated and managed together. When embarking on this project, I did not see any other solution that simply forwarded connections to remote hosts. Now, AWX can connect to hosts via k8s services which enables hosts to be located on any other, internet connected network.

I made a few design decisions you may find odd, so let me explain:

- I do not want to deal with running commands on a remote host. There are many tools, like Ansible, that do this well and are supported. Instead, I wanted a way to be able to reach the remote host even through a proxy and public networks.

- I wanted it to be Kubernetes-native so that it can be self-contained. This enables the usage of other tools inside the cluster more easily. 

- I wanted to have the option to integrate with Prometheus and Netbox. This is still a work-in-progress, but the idea that it can provide more accurate updates to monitoring and documentation services is a nice one.

## Installation

This requires skaffold to build and run currently. 

Navigate to `deploy/server` and run `skaffold run`. This will deploy to the cluster you are currently connected with inside the `caravan` namespace.

## Usage

Right now there isn't a simple tutorial. The essentials are:

1. Install on a K8s server
2. Setup external access via ingress or gateway api
3. Build the client
4. Put the client on the remote host you wish to have access to with the following systemd service file:

```
[Unit]
Description=Nibious LLC - Caravan - Remote access tool
After=network.target

[Service]
ExecStart=<YOUR_LOCATION_FOR_THE_CLIENT_BINARY>
Restart=always
RestartSec=5
# RuntimeMaxUSec=1d # if your device is a laptop or something moving around this can be helpful
Environment="CLIENTID=SOME_UUID_YOU_GENERATED_THAT_IS_UNIQUE_AMONG_CLIENTS" 
Environment="SECRET=A_RANDOM_64_CHARACTER_STRING" 
Environment="ADDRESS=YOUR_DNS_ADDRESS_WITHOUT_THE_SCHEME"


[Install]
WantedBy=default.target
```


To generate the Client ID and secret, you can use the following command:

```bash

``echo;echo -n "client_target: CLIENTID="; uuidgen; echo -n "client_target: SECRET="; < /dev/urandom tr -dc _A-Z-a-z-0-9- | head -c${1:-64};echo;echo;echo;
```

5. After deploying, check the logs. On the k8s server, you should be able to type: `kubectl get clients` and see a `true` if everything worked well.

## Contributing

1. Fork the repository.
2. Create a new branch: `git checkout -b feature-name`.
3. Make your changes.
4. Push your branch: `git push origin feature-name`.
5. Create a pull request.

## License

This project is licensed under the [MIT License](LICENSE).


## FAQs

- Why call it "Caravan"?

One definition of caravan is: 

> A company of travelers journeying together, as across a desert or through hostile territory. 

Whereas this is a remote tunneling solution that enables remote access to 
devices across the internet. I'm horrible at naming things, so it was my
best attempt. If you see the name `Iunctio` around the code base, that 
was the original name. It is latin for connection/joining. I'm still
working to remove and rename some of that.
