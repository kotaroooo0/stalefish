.PHONY: test
test:
	go test ./... -v

.PHONY: mock
mock:
	mockgen -source=storage.go -destination=mock_storage.go -package stalefish
	mockgen -source=morphology/morphology.go -destination=mock_morphology.go -package stalefish
