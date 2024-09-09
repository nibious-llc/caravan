
build: build/server build/client

build/%: cmd/%/main.go internal/**/*.go pkg/**/**/*.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $@ $<

.PHONY: setup
setup:
	curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.22.0/kind-linux-amd64
	install -m 0755 kind ~/.local/bin/kind

	curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
	install -m 0755 skaffold ~/.local/bin/

	curl -Lo helm.tar.gz https://get.helm.sh/helm-v3.14.1-linux-amd64.tar.gz
	tar -zxvf helm.tar.gz
	install -m 0755 linux-amd64/helm ~/.local/bin/
	rm -r linux-amd64

	curl -LO https://storage.googleapis.com/container-structure-test/latest/container-structure-test-linux-amd64 
	chmod +x container-structure-test-linux-amd64 
	install -m 0755 container-structure-test-linux-amd64 ~/.local/bin/container-structure-test

	curl -LO https://dl.k8s.io/release/v1.29.1/bin/linux/amd64/kubectl
	install -m 0755 kubectl ~/.local/bin/

.PHONY: clean
clean: 
	rm -rf build/

.PHONY: test
test:
	# delete the cluster just in case it is still running
	-kind delete cluster  --name nibious-caravan-test
	kind create cluster   --name nibious-caravan-test
	k apply -f deploy/server/crds.yaml
	cd deploy/server && skaffold run


