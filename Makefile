.PHONY: loadtest loadtest-remote test test-short

# Full /ws subscriber load test: DynamoDB Local + paced mock + app + k6.
# Requires docker, aws CLI, k6, go. Tunable via env, e.g. MAX_VUS=500 HOLD=90s.
loadtest:
	./loadtest/run.sh

# Load test an already-running REMOTE /ws endpoint (default wss://pushlap.co/ws).
# Only runs k6; optionally samples the server container via SSH_TARGET=ec2-user@<ip>.
# Tunable via env, e.g. WS_URL=wss://pushlap.co/ws MAX_VUS=50.
loadtest-remote:
	./loadtest/remote.sh

# Integration + unit tests (needs Docker for the LocalStack replay test).
test:
	go test ./...

# Skip the heavy LocalStack replay test.
test-short:
	go test -short ./...
