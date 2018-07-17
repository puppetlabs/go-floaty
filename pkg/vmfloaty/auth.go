package vmfloaty

import (
	"fmt"
)

type VMStatus struct {
	Running []string
}

type Detail struct {
	User    string
	Created string
	Last    string
	VMs     VMStatus
}

type UserStatus struct {
	Ok     bool
	Detail Detail
}

func TokenStatus(client PoolerClient) UserStatus {
	var data map[string]interface{}
	get(&client.client, fmt.Sprintf("%s/token/%s", client.config.URL, client.config.Token), &data)

	status := UserStatus{}
	detail := &Detail{}
	if processResponse(data, client.config.Token, detail) {
		status.Ok = true
		status.Detail = *detail
	}

	return status
}
