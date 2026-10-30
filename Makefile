IMAGE_NAME ?= http-from-scratch:v2
CLUSTER_NAME ?= desktop

.PHONY: all setup setup-vmstack image deploy-local clean clean-all help forward-app forward-grafana

all: setup

setup: setup-vmstack deploy-local
	@echo "=== Alles erfolgreich eingerichtet! ==="
	@echo "App testen:      make forward-app"
	@echo "Metriken prüfen: make forward-grafana"

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
	kubectl delete pod -l app=http-from-scratch --ignore-not-found --wait=true
	kubectl wait --for=condition=ready pod -l app=http-from-scratch --timeout=60s

forward-app:
	@echo "App direkt erreichbar unter http://localhost:8080"
	kubectl port-forward svc/http-from-scratch 8080:8080

forward-grafana:
	@echo "Grafana Login: admin / admin -> http://localhost:3000"
	kubectl port-forward -n monitoring svc/vmstack-grafana 3000:80

clean:
	kubectl delete -f k8s/ --ignore-not-found
	-docker rmi $(IMAGE_NAME)

clean-all: clean
	-helm uninstall vmstack -n monitoring
	-kubectl delete namespace monitoring
