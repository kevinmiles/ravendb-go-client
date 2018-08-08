package ravendb

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTcpInfoTest_canGetTcpInfo(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := NewGetTcpInfoCommand("test")
	err := store.GetRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)
	assert.Nil(t, result.getCertificate())
	// Note: in Java this tests for non-nil but Port is not sent
	// in Json, so don't quite understand that. Unless Java check
	// is bogus
	assert.Equal(t, 0, result.getPort())
	assert.NotEmpty(t, result.getUrl())
}

func TestGetTcpInfo(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			fmt.Printf("Panic: '%v'\n", r)
			debug.PrintStack()
			t.Fail()
		}
	}()

	getTcpInfoTest_canGetTcpInfo(t)
}
