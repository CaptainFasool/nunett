// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
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
        "/free": {
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
                "description": "Gets a list of peers the adapter can see within the network and return a list of peer info",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "run"
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
                "description": "HandleDeploymentRequest searches the DHT for non-busy, available devices with appropriate metadata. Then sends a deployment request to the first machine",
                "summary": "Search devices on DHT with appropriate machines and sends a deployment request.",
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
	Version:          "0.4.23",
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
