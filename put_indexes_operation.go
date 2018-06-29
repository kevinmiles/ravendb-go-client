package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IMaintenanceOperation = &PutIndexesOperation{}
)

type PutIndexesOperation struct {
	_indexToAdd []*IndexDefinition

	Command *PutIndexesCommand
}

func NewPutIndexesOperation(indexToAdd ...*IndexDefinition) *PutIndexesOperation {
	return &PutIndexesOperation{
		_indexToAdd: indexToAdd,
	}
}

func (o *PutIndexesOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewPutIndexesCommand(conventions, o._indexToAdd)
	return o.Command
}

var _ RavenCommand = &PutIndexesCommand{}

type PutIndexesCommand struct {
	*RavenCommandBase

	_indexToAdd []ObjectNode

	Result []*PutIndexResult
}

func NewPutIndexesCommand(conventions *DocumentConventions, indexesToAdd []*IndexDefinition) *PutIndexesCommand {
	panicIf(conventions == nil, "conventions cannot be null")
	panicIf(indexesToAdd == nil, "indexesToAdd cannot be null")

	cmd := &PutIndexesCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}

	for _, indexToAdd := range indexesToAdd {
		panicIf(indexToAdd.getName() == "", "Index name cannot be null")
		objectNode := EntityToJson_convertEntityToJson(indexToAdd, nil)
		cmd._indexToAdd = append(cmd._indexToAdd, objectNode)
	}

	return cmd
}

func (c *PutIndexesCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/indexes"

	m := map[string]interface{}{
		"Indexes": c._indexToAdd,
	}
	d, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("\nPutIndexesCommand.createRequest:\n%s\n\n", string(d))
	return NewHttpPut(url, d)
}

func (c *PutIndexesCommand) setResponse(response []byte, fromCache bool) error {
	var res PutIndexesResponse
	err := json.Unmarshal(response, &res)
	if err != nil {
		dbg("PutIndexesCommand.setResponse: json.Unmarshal failed with %s. JSON:\n%s\n\n", err, string(response))
		return err
	}
	c.Result = res.Results
	return nil
}
