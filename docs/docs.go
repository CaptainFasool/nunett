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
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Metadata"
                            }
                        }
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
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Metadata"
                            }
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
        "/run/request-reward": {
            "post": {
                "description": "HandleRequestReward takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.",
                "summary": "Get NTX tokens for work done.",
                "responses": {}
            }
        },
        "/run/request-service": {
            "post": {
                "description": "HandleRequestService searches the DHT for non-busy, available devices with appropriate metadata. Then informs parameters related to blockchain to request to run a service on NuNet.",
                "summary": "Informs parameters related to blockchain to request to run a service on NuNet",
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
        "/run/send-status": {
            "post": {
                "description": "HandleSendStatus is used by webapps to send status of blockchain activities. Such as if tokens have been put in escrow account and account creation.",
                "summary": "Sends blockchain status of contract creation.",
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
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.4.61",
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
