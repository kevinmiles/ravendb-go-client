package ravendb

import (
	"encoding/json"
	"net/http"
)

func GetClusterTopologyCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	url := node.getUrl() + "/cluster/topology"
	return NewHttpGet(url), url
}

func GetClusterTopologyCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res ClusterTopologyResponse
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}

func NewGetClusterTopologyCommand() *RavenCommand {
	res := NewRavenCommand()
	res.IsReadRequest = true
	res.createRequestFunc = GetClusterTopologyCommand_createRequest
	res.setResponseFunc = GetClusterTopologyCommand_setResponse
	return res
}
