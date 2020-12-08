.PHONY: test
test:
	go test ./... -v

.PHONY: truncate
truncate:
	mysql -h127.0.0.1 -uroot -ppassword -D stalefish -e "truncate table tokens;truncate table documents;truncate table inverted_indexes;truncate table compressed_inverted_indexes"
