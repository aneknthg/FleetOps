.PHONY: up down test scan

LOCAL_ENVS := local-mgmt local-dev local-prod

up:
	@set -eu; for env in $(LOCAL_ENVS); do \
		echo "Applying terraform/envs/$$env"; \
		terraform -chdir=terraform/envs/$$env init -input=false; \
		terraform -chdir=terraform/envs/$$env apply -input=false -auto-approve; \
	done

down:
	@set -eu; for env in local-prod local-dev local-mgmt; do \
		echo "Destroying terraform/envs/$$env"; \
		terraform -chdir=terraform/envs/$$env destroy -input=false -auto-approve; \
	done

test:
	cd sample-app && go test ./... && cd ../tools/fleetctl && go test ./...

scan:
	trivy config terraform/ gitops/
