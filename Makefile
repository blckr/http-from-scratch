IMAGE_NAME ?= http-from-scratch:latest
CLUSTER_NAME ?= desktop

.PHONY: all setup setup-ingress setup-vmstack image deploy-local clean clean-all help

all: setup

setup: setup-ingress setup-vmstack deploy-local
	@echo "=== Alles erfolgreich eingerichtet! ==="
	@echo "Server: http://myapp.local/test"
	@echo "Grafana: 'kubectl port-forward -n monitoring svc/vmstack-grafana 3000:80' -> http://localhost:3000"

setup-ingress:
	@helm upgrade --install ingress-nginx ingress-nginx \
		--repo https://kubernetes.github.io/ingress-nginx \
		--namespace ingress-nginx --create-namespace
	@kubectl wait --namespace ingress-nginx \
		--for=condition=ready pod \
		--selector=app.kubernetes.io/component=controller \
		--timeout=90s

setup-vmstack:
	@helm repo add vm https://victoriametrics.github.io/helm-charts/ 2>/dev/null || true
	@helm repo update
	@helm upgrade --install vmstack vm/victoria-metrics-k8s-stack \
		--namespace monitoring --create-namespace \
		-f helm/vmstack-values.yaml

image:
	docker build --no-cache -t $(IMAGE_NAME) .
	kind load docker-image $(IMAGE_NAME) --name $(CLUSTER_NAME)

deploy-local: image
	kubectl apply -f k8s/
	kubectl delete pod -l app=http-from-scratch --wait=true
	kubectl wait --for=condition=ready pod -l app=http-from-scratch --timeout=60s

clean:
	kubectl delete -f k8s/ --ignore-not-found
	-docker rmi $(IMAGE_NAME) http-from-scratch:v1

clean-all: clean
	-helm uninstall vmstack -n monitoring
	-helm uninstall ingress-nginx -n ingress-nginx
	-kubectl delete namespace monitoring ingress-nginx
