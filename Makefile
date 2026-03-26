# Numele proiectului
PROJECT_NAME=licenta-iot
NAMESPACE=licenta-iot

.PHONY: deploy undeploy simulate status logs-alert

# Aplică toată infrastructura în Kubernetes
deploy:
	@echo "🚀 Aplicăm manifestele K8s..."
	kubectl apply -f k8s/00-base.yaml
	kubectl apply -f k8s/infrastructure/
	kubectl apply -f k8s/monitoring/
	kubectl apply -f k8s/apps/

# Șterge tot din Kubernetes
undeploy:
	@echo "🛑 Ștergem resursele K8s..."
	kubectl delete -f k8s/apps/
	kubectl delete -f k8s/monitoring/
	kubectl delete -f k8s/infrastructure/

# Rulează simulatorul (CALE ACTUALIZATĂ)
simulate:
	@echo "📡 Pornim simulatorul de senzori (Smart Factory)..."
	go run ./cmd/simulator/main.go

# Verifică starea pod-urilor
status:
	kubectl get pods -n $(NAMESPACE)

# Vezi log-urile Alertmanager pentru depanare rapidă
logs-alert:
	kubectl logs -l app=alertmanager -n $(NAMESPACE)
