
check-binary: 
	@if [ ! -f "./marketWatch" ]; then \
		echo "❌ 'marketWatch' binary missing!"; \
		echo "👉 Suggestion: \n\t Run 'make build' \n"; \
		exit 1; \
	fi

check-config:
	@if [ ! -f "./config.yaml" ]; then \
		echo "❌ 'config.yaml' missing!"; \
		echo "👉 Suggestion: \n\t Run 'cp .config_sample.yaml config.yaml'"; \
		echo "\t modify the config file as per your environment\n"; \
		exit 1; \
	fi

# -----------------------------------------------

run-serve: check-config
	go run main.go serve -p 8080 -c ./config.yaml

build:
	go build -o marketWatch main.go

lint:
	golangci-lint run --fix

serve: check-binary check-config
	./marketWatch serve -p 8080 -c ./config.yaml

refresh-trends: check-binary check-config
	./marketWatch refresh-trends -c ./config.yaml
