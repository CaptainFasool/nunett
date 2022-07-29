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
        "/actions": {
            "put": {
                "description": "Start or stop the VM.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Start or stop the VM.",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/address/new": {
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
                            "$ref": "#/definitions/models.AddressPrivKey"
                        }
                    }
                }
            }
        },
        "/boot-source": {
            "put": {
                "description": "Configure kernel for the VM.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Configures kernel for the VM.",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/drives": {
            "put": {
                "description": "Configures filesystem for the VM.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Configures filesystem for the VM.",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/init": {
            "post": {
                "description": "Starts the firecracker server for the specific VM. Further configuration are required.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Starts the VM booting process.",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/machine-config": {
            "put": {
                "description": "Configures system spec for the VM like CPU and Memory.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Configures system spec for the VM.",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/metadata": {
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
        "/network-interface": {
            "put": {
                "description": "Configures network interface on the host.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Configures network interface on the host.",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/onboard": {
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
        "/provisioned": {
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
        "/start-default": {
            "post": {
                "description": "This endpoint is an abstraction of all other endpoints. When invokend, it calls all other endpoints in a sequence.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vm"
                ],
                "summary": "Start a VM with default configuration.",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        }
    },
    "definitions": {
        "models.AddressPrivKey": {
            "type": "object",
            "properties": {
                "address": {
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
	Version:          "0.1",
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
