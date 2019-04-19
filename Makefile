MYSQLTEST_PORT = 33061
MYSQLTEST_USER = root
MYSQLTEST_PASS = root

test:
	@docker-compose up -d --renew-anon-volumes mysqltest
	@docker-compose up waitformysqltest
	@docker-compose up testdatafiller
	env MYSQLTEST_PORT=$(MYSQLTEST_PORT) MYSQLTEST_USER=$(MYSQLTEST_USER) MYSQLTEST_PASS=$(MYSQLTEST_PASS) go test
	@docker-compose down