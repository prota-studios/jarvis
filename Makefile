.PHONY: gen


gen:
		swagger generate server

values.decrypted.yaml:
		sops --decrypt values.yaml > values.decrypted.yaml

deploy: values.decrypted.yaml
		helm upgrade --install jarvis charts/jarvis \
		--values values.decrypted.yaml
dev: values.decrypted.yaml
		skaffold dev
creds:
		kubectl create secret docker-registry gcr-json-key \
		--docker-server=gcr.io \
		--docker-username=_json_key \
		--docker-password="`cat creds.json`" \
		--docker-email=any@valid.email