// Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://nunet.io/tos",
        "contact": {
            "name": "Support",
            "url": "https://devexchange.nunet.io/",
            "email": "support@nunet.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/dht": {
            "get": {
                "description": "Returns entire DHT content",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "p2p"
                ],
                "summary": "Return a dump of the dht",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/onboarding/address/new": {
            "get": {
                "description": "Create a payment address from public key. Return payment address and private key.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "onboarding"
                ],
                "summary": "Create a new payment address.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.BlockchainAddressPrivKey"
                        }
                    }
                }
            }
        },
        "/onboarding/metadata": {
            "get": {
                "description": "Responds with metadata of current provideer",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "onboarding"
                ],
                "summary": "Get current device info.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Metadata"
                        }
                    }
                }
            }
        },
        "/onboarding/offboard": {
            "delete": {
                "description": "Offboard runs the offboarding script to remove resources associated with a device.",
                "tags": [
                    "onboarding"
                ],
                "summary": "Runs the offboarding process.",
                "responses": {
                    "200": {
                        "description": "Successfully Onboarded"
                    }
                }
            }
        },
        "/onboarding/onboard": {
            "post": {
                "description": "Onboard runs onboarding script given the amount of resources to onboard.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "onboarding"
                ],
                "summary": "Runs the onboarding process.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Metadata"
                        }
                    }
                }
            }
        },
        "/onboarding/provisioned": {
            "get": {
                "description": "Get total memory capacity in MB and CPU capacity in MHz.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "onboarding"
                ],
                "summary": "Returns provisioned capacity on host.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Provisioned"
                        }
                    }
                }
            }
        },
        "/onboarding/resource-config": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "onboarding"
                ],
                "summary": "changes the amount of resources of onboarded device .",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Metadata"
                        }
                    }
                }
            }
        },
        "/onboarding/status": {
            "get": {
                "description": "Returns json with 5 parameters: onboarded, error, machine_uuid, metadata_path, database_path.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "onboarding"
                ],
                "summary": "Onboarding status and other metadata.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.OnboardingStatus"
                        }
                    }
                }
            }
        },
        "/peers": {
            "get": {
                "description": "Gets a list of peers the libp2p node can see within the network and return a list of peers",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "p2p"
                ],
                "summary": "Return list of peers currently connected to",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/peers/chat": {
            "get": {
                "description": "Get a list of chat requests from peers",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "chat"
                ],
                "summary": "List chat requests",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/peers/chat/clear": {
            "get": {
                "description": "Clear chat request streams from peers",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "chat"
                ],
                "summary": "Clear chat requests",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/peers/chat/join": {
            "get": {
                "description": "Join a chat session started by a peer",
                "tags": [
                    "chat"
                ],
                "summary": "Join chat with a peer",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/peers/chat/start": {
            "get": {
                "description": "Start chat session with a peer",
                "tags": [
                    "chat"
                ],
                "summary": "Start chat with a peer",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/peers/depreq": {
            "get": {
                "description": "Set peer as the default receipient of deployment requests by setting the peerID parameter on GET request.\nShow peer set as default deployment request receiver by sending a GET request without any parameters.\nRemove default deployment request receiver by sending a GET request with peerID parameter set to '0'.",
                "tags": [
                    "peers"
                ],
                "summary": "Manage default deplyment request receiver peer",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/peers/dht": {
            "get": {
                "description": "Gets a list of peers the libp2p node has received a dht update from",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "p2p"
                ],
                "summary": "Return list of peers which have sent a dht update",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/peers/kad-dht": {
            "get": {
                "description": "Gets a list of peers the libp2p node has received a dht update from",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "p2p"
                ],
                "summary": "Return list of peers which have sent a dht update",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/peers/self": {
            "get": {
                "description": "Gets self peer info of libp2p node",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "p2p"
                ],
                "summary": "Return self peer info",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/peers/ws": {
            "get": {
                "description": "Sends a command to specific node and prints back response.",
                "tags": [
                    "peers"
                ],
                "summary": "Sends a command to specific node and prints back response.",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/run/deploy": {
            "get": {
                "description": "Loads deployment request from the DB after a successful blockchain transaction has been made and passes it to compute provider.",
                "tags": [
                    "run"
                ],
                "summary": "Websocket endpoint responsible for sending deployment request and receiving deployment response.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/run/request-service": {
            "post": {
                "description": "HandleRequestService searches the DHT for non-busy, available devices with appropriate metadata. Then informs parameters related to blockchain to request to run a service on NuNet.",
                "tags": [
                    "run"
                ],
                "summary": "Informs parameters related to blockchain to request to run a service on NuNet",
                "parameters": [
                    {
                        "description": "Deployment Request",
                        "name": "deployment_request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.DeploymentRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/machines.fundingRespToSPD"
                        }
                    }
                }
            }
        },
        "/telemetry/free": {
            "get": {
                "description": "Checks and returns the amount of free resources available",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "telemetry"
                ],
                "summary": "Returns the amount of free resources available",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/transactions": {
            "get": {
                "description": "Get list of TxHashes along with the date and time of jobs done.",
                "tags": [
                    "run"
                ],
                "summary": "Get list of TxHashes for jobs done.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/tokenomics.TxHashResp"
                        }
                    }
                }
            }
        },
        "/transactions/request-reward": {
            "post": {
                "description": "HandleRequestReward takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.",
                "tags": [
                    "run"
                ],
                "summary": "Get NTX tokens for work done.",
                "parameters": [
                    {
                        "description": "Claim Reward Body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/tokenomics.ClaimCardanoTokenBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/tokenomics.rewardRespToCPD"
                        }
                    }
                }
            }
        },
        "/transactions/send-status": {
            "post": {
                "description": "HandleSendStatus is used by webapps to send status of blockchain activities. Such token withdrawl.",
                "tags": [
                    "run"
                ],
                "summary": "Sends blockchain status of contract creation.",
                "parameters": [
                    {
                        "description": "Blockchain Transaction Status Body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BlockchainTxStatus"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/vm/start-custom": {
            "post": {
                "description": "This endpoint is an abstraction of all primitive endpoints. When invokend, it calls all primitive endpoints in a sequence.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Start a VM with custom configuration.",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/vm/start-default": {
            "post": {
                "description": "Everything except kernel files and filesystem file will be set by DMS itself.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Start a VM with default configuration.",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        }
    },
    "definitions": {
        "machines.fundingRespToSPD": {
            "type": "object",
            "properties": {
                "compute_provider_addr": {
                    "type": "string"
                },
                "distribute_hash": {
                    "type": "string"
                },
                "estimated_price": {
                    "type": "number"
                },
                "metadata_hash": {
                    "type": "string"
                },
                "refund_hash": {
                    "type": "string"
                },
                "withdraw_hash": {
                    "type": "string"
                }
            }
        },
        "models.BlockchainAddressPrivKey": {
            "type": "object",
            "properties": {
                "address": {
                    "type": "string"
                },
                "mnemonic": {
                    "type": "string"
                },
                "private_key": {
                    "type": "string"
                }
            }
        },
        "models.BlockchainTxStatus": {
            "type": "object",
            "properties": {
                "transaction_status": {
                    "type": "string"
                },
                "transaction_type": {
                    "type": "string"
                },
                "tx_hash": {
                    "type": "string"
                }
            }
        },
        "models.DeploymentRequest": {
            "type": "object",
            "properties": {
                "address_user": {
                    "description": "service provider wallet address",
                    "type": "string"
                },
                "blockchain": {
                    "type": "string"
                },
                "constraints": {
                    "type": "object",
                    "properties": {
                        "complexity": {
                            "type": "string"
                        },
                        "cpu": {
                            "type": "integer"
                        },
                        "power": {
                            "type": "integer"
                        },
                        "ram": {
                            "type": "integer"
                        },
                        "time": {
                            "type": "integer"
                        },
                        "vram": {
                            "type": "integer"
                        }
                    }
                },
                "distribute_hash": {
                    "type": "string"
                },
                "estimated_ntx": {
                    "type": "number"
                },
                "max_ntx": {
                    "type": "integer"
                },
                "metadata_hash": {
                    "type": "string"
                },
                "params": {
                    "type": "object",
                    "properties": {
                        "image_id": {
                            "type": "string"
                        },
                        "local_node_id": {
                            "description": "NodeID of service provider (machine triggering the job)",
                            "type": "string"
                        },
                        "local_public_key": {
                            "description": "Public key of service provider",
                            "type": "string"
                        },
                        "machine_type": {
                            "type": "string"
                        },
                        "model_url": {
                            "type": "string"
                        },
                        "node_id": {
                            "description": "NodeID of compute provider (machine to deploy the job on)",
                            "type": "string"
                        },
                        "packages": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        },
                        "public_key": {
                            "description": "Public key of compute provider",
                            "type": "string"
                        }
                    }
                },
                "refund_hash": {
                    "type": "string"
                },
                "service_type": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "traceinfo": {
                    "type": "object",
                    "properties": {
                        "span_id": {
                            "type": "string"
                        },
                        "trace_flags": {
                            "type": "string"
                        },
                        "trace_id": {
                            "type": "string"
                        },
                        "trace_state": {
                            "type": "string"
                        }
                    }
                },
                "tx_hash": {
                    "type": "string"
                },
                "withdraw_hash": {
                    "type": "string"
                }
            }
        },
        "models.Metadata": {
            "type": "object",
            "properties": {
                "available": {
                    "type": "object",
                    "properties": {
                        "ram": {
                            "type": "integer"
                        },
                        "update_timestamp": {
                            "type": "integer"
                        }
                    }
                },
                "name": {
                    "type": "string"
                },
                "network": {
                    "type": "string"
                },
                "public_key": {
                    "type": "string"
                },
                "reserved": {
                    "type": "object",
                    "properties": {
                        "cpu": {
                            "type": "integer"
                        },
                        "memory": {
                            "type": "integer"
                        }
                    }
                },
                "resource": {
                    "type": "object",
                    "properties": {
                        "cpu_max": {
                            "type": "number"
                        },
                        "cpu_usage": {
                            "type": "number"
                        },
                        "ram_max": {
                            "type": "integer"
                        },
                        "total_core": {
                            "type": "integer"
                        },
                        "update_timestamp": {
                            "type": "integer"
                        }
                    }
                }
            }
        },
        "models.OnboardingStatus": {
            "type": "object",
            "properties": {
                "database_path": {
                    "type": "string"
                },
                "error": {
                    "type": "string"
                },
                "machine_uuid": {
                    "type": "string"
                },
                "metadata_path": {
                    "type": "string"
                },
                "onboarded": {
                    "type": "boolean"
                }
            }
        },
        "models.Provisioned": {
            "type": "object",
            "properties": {
                "cpu": {
                    "type": "number"
                },
                "memory": {
                    "type": "integer"
                },
                "total_cores": {
                    "type": "integer"
                }
            }
        },
        "tokenomics.ClaimCardanoTokenBody": {
            "type": "object",
            "properties": {
                "compute_provider_address": {
                    "type": "string"
                },
                "tx_hash": {
                    "type": "string"
                }
            }
        },
        "tokenomics.TxHashResp": {
            "type": "object",
            "properties": {
                "date_time": {
                    "type": "string"
                },
                "tx_hash": {
                    "type": "string"
                }
            }
        },
        "tokenomics.rewardRespToCPD": {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string"
                },
                "datum": {
                    "type": "string"
                },
                "message_hash_action": {
                    "type": "string"
                },
                "message_hash_datum": {
                    "type": "string"
                },
                "reward_type": {
                    "type": "string"
                },
                "signature_action": {
                    "type": "string"
                },
                "signature_datum": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.4.126",
	Host:             "localhost:9999",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Device Management Service",
	Description:      "A dashboard application for computing providers.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
