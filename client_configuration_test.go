package ravendb

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func clientConfiguration_canHandleNoConfiguration(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	operation := NewGetClientConfigurationOperation()
	err := store.maintenance().send(operation)
	assert.NoError(t, err)
	result := operation.Command.Result
	assert.Nil(t, result.getConfiguration())
	//TODO: java checks that result.getEtag() is not nil, which does not apply
}

func clientConfiguration_canSaveAndReadClientConfiguration(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	configurationToSave := NewClientConfiguration()
	configurationToSave.setEtag(123)
	configurationToSave.setMaxNumberOfRequestsPerSession(80)
	configurationToSave.setReadBalanceBehavior(ReadBalanceBehavior_FASTEST_NODE)
	configurationToSave.setDisabled(true)

	saveOperation, err := NewPutClientConfigurationOperation(configurationToSave)
	assert.NoError(t, err)
	store.maintenance().send(saveOperation)
	operation := NewGetClientConfigurationOperation()
	err = store.maintenance().send(operation)
	assert.NoError(t, err)
	result := operation.Command.Result
	assert.True(t, result.getEtag() > 0)
	newConfiguration := result.getConfiguration()
	assert.NotNil(t, newConfiguration)
	assert.True(t, newConfiguration.getEtag() > configurationToSave.getEtag())
	assert.True(t, newConfiguration.isDisabled())
	assert.Equal(t, newConfiguration.getMaxNumberOfRequestsPerSession(), 80)
	assert.Equal(t, newConfiguration.getReadBalanceBehavior(), ReadBalanceBehavior_FASTEST_NODE)
}

func TestClientConfiguration(t *testing.T) {
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

	// matches order of Java tests
	clientConfiguration_canHandleNoConfiguration(t)
	clientConfiguration_canSaveAndReadClientConfiguration(t)
}
