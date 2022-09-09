package models

const ClientTemplate = `{
  "plugin": {
    "raw_exec": {
      "config": {
        "enabled": "true"
      }
    },
    "docker": {}
  },
  "log_level": "DEBUG",
  "data_dir": "/var/log/nomad",
  "name": "{{ .Name }}",
  "datacenter": "{{ .Network }}",
  "client": {
    "host_volume": {
      "machine-metadata": {
        "name": "machine-metadata",
        "path": "/etc/nunet"
      }
    },
    "enabled": "true",
    "servers": [
      "nomad-nunetio.ddns.net:4647"
    ],
    "reserved": {
      "cpu": {{ .ReservedCPU }},
      "memory": {{ .ReservedMemory }}
    }
  }
}
`

const AdapterTemplate = `{
  "Job": {
    "Region": null,
    "Namespace": null,
    "ID": "{{ .AdapterPrefix }}-{{ .ClientName }}",
    "Name": "{{ .AdapterPrefix }}-{{ .ClientName }}",
    "Type": null,
    "Priority": null,
    "AllAtOnce": null,
    "Datacenters": [
      "{{ .Datacenters }}"
    ],
    "Constraints": null,
    "Affinities": null,
    "TaskGroups": [
      {
        "Name": "{{ .AdapterPrefix }}-{{ .ClientName }}",
        "Count": 1,
        "Constraints": null,
        "Affinities": null,
        "Tasks": [
            {
              "Name": "{{ .AdapterPrefix }}-{{ .ClientName }}",
              "Driver": "docker",
              "User": "",
              "Lifecycle": null,
              "Config": {
                "args": [
                  "opentelemetry-instrument",
                  "python3",
                  "nunet_adapter.py",
                  " ${NOMAD_PORT_rpc}"
                ],
                  "image": "registry.gitlab.com/nunet/nunet-adapter:{{ .DockerTag }}",
                  "network_mode": "host"
                },
                "Constraints": [
                  {
                    "LTarget": "${node.unique.name}",
                    "RTarget": "{{ .ClientName }}",
                    "Operand": "="
                  }
                ],
                "Affinities": null,
                "Env": {
                  "LS_ACCESS_TOKEN": "ej4S2fSomE5V7uP/LYJLw9frBRnBRqHVoQiRYSCwoSld1MncIAHByOSkc8jcKiSAbhVEy2zI0Emjw38vVF8c87vFd6Q1wMCnI/DD4/Bi",
                  "LS_SERVICE_NAME": "{{ .AdapterPrefix }}-{{ .ClientName }}",
                  "OTEL_PYTHON_LOG_CORRELATION": "true",
                  "OTEL_PYTHON_TRACER_PROVIDER": "sdk_tracer_provider",
                  "deployment_type": "{{ .DeploymentType }}",
                  "tokenomics_api_name": "{{ .TokenomicsApiName }}",
      "cardano_passive": "{{ .Cardano }}"
                },
                "Services": [
                  {
                    "Id": "",
                    "Name": "{{ .AdapterPrefix }}-{{ .ClientName }}",
                    "Tags": [
                      "theNunetMachine",
                      "urlprefix-/{{ .AdapterPrefix }}-{{ .ClientName }} proto=grpc"
                    ],
                  "CanaryTags": null,
                  "EnableTagOverride": false,
                  "PortLabel": "rpc",
                  "AddressMode": "",
                  "CheckRestart": null,
                  "Connect": null,
                  "Meta": null,
                  "CanaryMeta": null,
                  "TaskName": "",
                  "OnUpdate": ""
                }
              ],
              "Resources": {
                "CPU": 2000,
                "Cores": null,
                "MemoryMB": 1000,
                "MemoryMaxMB": null,
                "DiskMB": null,
                "Networks": null,
                "Devices": null,
                "IOPS": null
              },
              "RestartPolicy": null,
              "Meta": null,
              "KillTimeout": null,
              "LogConfig": {
                "MaxFiles": 3,
                "MaxFileSizeMB": 5
              },
              "Artifacts": null,
              "Vault": null,
              "Templates": null,
              "DispatchPayload": null,
              "VolumeMounts": [
                {
                  "Volume": "metadata",
                  "Destination": "/etc/nunet",
                  "ReadOnly": null,
                  "PropagationMode": null
                }
              ],
            "Leader": false,
            "ShutdownDelay": 0,
            "KillSignal": "",
            "Kind": "",
            "ScalingPolicies": null
          }
        ],
        "Spreads": null,
        "Volumes": {
          "metadata":{
            "Name":"metadata",
            "Type":"host",
            "Source":"machine-metadata"
          }
        },
        "RestartPolicy": null,
        "ReschedulePolicy": null,
        "EphemeralDisk": null,
        "Update": null,
        "Migrate": null,
        "Networks": [
          {
            "Mode": "",
            "Device": "",
            "CIDR": "",
            "IP": "",
            "DNS": null,
            "ReservedPorts": [
              {
                "Label": "rpc",
                "Value": 60777,
                "To": 0,
                "HostNetwork": ""
              }
            ],
            "DynamicPorts": null,
            "MBits": null
          }
        ],
        "Meta": null,
        "Services": null,
        "ShutdownDelay": null,
        "StopAfterClientDisconnect": null,
        "Scaling": null,
        "Consul": null
      }
    ],
    "Update": null,
    "Multiregion": null,
    "Spreads": null,
    "Periodic": null,
    "ParameterizedJob": null,
    "Reschedule": null,
    "Migrate": null,
    "Meta": null,
    "ConsulToken": null,
    "VaultToken": null,
    "Stop": null,
    "ParentID": null,
    "Dispatched": false,
    "Payload": null,
    "ConsulNamespace": null,
    "VaultNamespace": null,
    "NomadTokenID": null,
    "Status": null,
    "StatusDescription": null,
    "Stable": null,
    "Version": null,
    "SubmitTime": null,
    "CreateIndex": null,
    "ModifyIndex": null,
    "JobModifyIndex": null
  }
}
`
