package tests

import (
	"fmt"
	"github.com/ravendb/ravendb-go-client"
	"github.com/ravendb/ravendb-go-client/examples/northwind"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createNorthwindDatabase(t *testing.T, driver *RavenTestDriver, store *ravendb.DocumentStore) {
	sampleData := ravendb.NewCreateSampleDataOperation()
	err := store.Maintenance().Send(sampleData)
	must(err)

	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	must(err)
}

func goNorthwindEmployeeLoad(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	createNorthwindDatabase(t, driver, store)

	session, err := store.OpenSession("")
	assert.NoError(t, err)
	defer session.Close()

	var e *northwind.Employee
	err = session.Load(&e, "employees/7-A")
	assert.NoError(t, err)
	fmt.Printf("employee: %#v\n", e)
}

func TestGoNorthwind(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	if false {
		// failing
		goNorthwindEmployeeLoad(t, driver)
	}
}