package messaging

import "gitlab.com/nunet/device-management-service/models"

// All message types are described below:
// - Deployment Request
// - Deployment Response
var MessageTypeToModel = make(map[string]interface{})

func init() {
	MessageTypeToModel["DeploymentRequest"] = models.DeploymentRequest{}
	MessageTypeToModel["DeploymentResponse"] = models.DeploymentResponse{}
}
