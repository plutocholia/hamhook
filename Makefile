.PHONY: build-image
build-image:
	docker build -t plutocholia/hamhook:latest . --no-cache

.PHONY: push-image
push-image:
	docker push plutocholia/hamhook:latest

.PHONY: kind-image-load
kind-image-load:
	kind load docker-image plutocholia/hamhook:latest --name c1

.PHONY: kind-image-rmi
kind-image-rmi:
	docker ps | grep c1 | awk '{print $$1}' | xargs -n 1 -P 1 sh -c 'docker exec $$0 crictl rmi docker.io/plutocholia/hamhook:latest'