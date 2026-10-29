IMAGE_NAME ?= http-from-scratch:latest

.PHONY: all setup setup-ingress setup-vmstack image deploy-local help

all: setup

setup: setup-ingress setup-vmstack deploy-local
	@echo "=== Alles erfolgreich eingerichtet! ==="
	@echo "Server: http://localhost/test"
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
	docker build -t $(IMAGE_NAME) .

deploy-local: image
	kubectl apply -f k8s/
	kubectl rollout restart deployment/http-from-scratch
	kubectl rollout status deployment/http-from-scratch --timeout=60s
