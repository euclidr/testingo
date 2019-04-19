package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/alicebob/miniredis"
)

var testRdsSrv *miniredis.Miniredis

func TestMain(m *testing.M) {
	// stub redis
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()
	testRdsSrv = s

	// stub mysql
	mysqlPort := os.Getenv("MYSQLTEST_PORT")
	mysqlUser := os.Getenv("MYSQLTEST_USER")
	mysqlPass := os.Getenv("MYSQLTEST_PASS")
	mysqlDsnFmt := "%s:%s@tcp(localhost:%s)/testingo?parseTime=true&charset=utf8mb4&loc=Local"
	mysqlDsn := fmt.Sprintf(mysqlDsnFmt, mysqlUser, mysqlPass, mysqlPort)
	_, err = InitDB(mysqlDsn)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
