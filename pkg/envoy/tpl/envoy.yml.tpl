admin:
  access_log_path: "/dev/null"
  address:
    socket_address:
      address: {{ .AdminAddress }}
      port_value: {{ .AdminPort }}
node:
  cluster: {{ .ServiceName }}
  id: {{ .ServiceID }}
  metadata:
    namespace: default
    partition: default
layered_runtime:
  layers:
  - name: base
    static_layer:
      re2.max_program_size.error_level: 1048576
static_resources:
  clusters:
  - name: local_agent
    ignore_health_on_host_removal: false
    connect_timeout: 1s
    type: STATIC
    http2_protocol_options: {}
    loadAssignment:
      clusterName: local_agent
      endpoints:
      - lbEndpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ .AgentAddress }}
                port_value: {{ .AgentPort }}
    {{- if .AgentTLS }}
    transport_socket:
      name: tls
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        common_tls_context:
          validation_context:
            trusted_ca:
              inline_string: "{{ .AgentCAPEM }}"
    {{- end }}
dynamic_resources:
  lds_config:
    ads: {}
    resource_api_version: V3
  cds_config:
    ads: {}
    resource_api_version: V3
  ads_config:
    api_type: DELTA_GRPC
    transport_api_version: V3
    grpc_services:
      initial_metadata:
      - key: x-consul-token
        value: "{{ .ConsulToken }}"
      envoy_grpc:
        cluster_name: local_agent
stats_config:
  stats_tags:
  - regex: "^cluster\\.(?:passthrough~)?((?:([^.]+)~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.custom_hash
  - regex: "^cluster\\.(?:passthrough~)?((?:[^.]+~)?(?:([^.]+)\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.service_subset
  - regex: "^cluster\\.(?:passthrough~)?((?:[^.]+~)?(?:[^.]+\\.)?([^.]+)\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.service
  - regex: "^cluster\\.(?:passthrough~)?((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.([^.]+)\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.namespace
  - regex: "^cluster\\.(?:passthrough~)?((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:([^.]+)\\.)?[^.]+\\.internal[^.]*\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.partition
  - regex: "^cluster\\.(?:passthrough~)?((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?([^.]+)\\.internal[^.]*\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.datacenter
  - regex: "^cluster\\.([^.]+\\.(?:[^.]+\\.)?([^.]+)\\.external\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.peer
  - regex: "^cluster\\.(?:passthrough~)?((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.([^.]+)\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.routing_type
  - regex: "^cluster\\.(?:passthrough~)?((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.([^.]+)\\.consul\\.)"
    tag_name: consul.destination.trust_domain
  - regex: "^cluster\\.(?:passthrough~)?(((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+)\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.destination.target
  - regex: "^cluster\\.(?:passthrough~)?(((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+)\\.consul\\.)"
    tag_name: consul.destination.full_target
  - regex: "^(?:tcp|http)\\.upstream(?:_peered)?\\.(([^.]+)(?:\\.[^.]+)?(?:\\.[^.]+)?\\.[^.]+\\.)"
    tag_name: consul.upstream.service
  - regex: "^(?:tcp|http)\\.upstream\\.([^.]+(?:\\.[^.]+)?(?:\\.[^.]+)?\\.([^.]+)\\.)"
    tag_name: consul.upstream.datacenter
  - regex: "^(?:tcp|http)\\.upstream_peered\\.([^.]+(?:\\.[^.]+)?\\.([^.]+)\\.)"
    tag_name: consul.upstream.peer
  - regex: "^(?:tcp|http)\\.upstream(?:_peered)?\\.([^.]+(?:\\.([^.]+))?(?:\\.[^.]+)?\\.[^.]+\\.)"
    tag_name: consul.upstream.namespace
  - regex: "^(?:tcp|http)\\.upstream\\.([^.]+(?:\\.[^.]+)?(?:\\.([^.]+))?\\.[^.]+\\.)"
    tag_name: consul.upstream.partition
  - regex: "^cluster\\.((?:([^.]+)~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.custom_hash
  - regex: "^cluster\\.((?:[^.]+~)?(?:([^.]+)\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.service_subset
  - regex: "^cluster\\.((?:[^.]+~)?(?:[^.]+\\.)?([^.]+)\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.service
  - regex: "^cluster\\.((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.([^.]+)\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.namespace
  - regex: "^cluster\\.((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?([^.]+)\\.internal[^.]*\\.[^.]+\\.consul\\.)"
    tag_name: consul.datacenter
  - regex: "^cluster\\.((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.([^.]+)\\.[^.]+\\.consul\\.)"
    tag_name: consul.routing_type
  - regex: "^cluster\\.((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.([^.]+)\\.consul\\.)"
    tag_name: consul.trust_domain
  - regex: "^cluster\\.(((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+)\\.[^.]+\\.[^.]+\\.consul\\.)"
    tag_name: consul.target
  - regex: "^cluster\\.(((?:[^.]+~)?(?:[^.]+\\.)?[^.]+\\.[^.]+\\.(?:[^.]+\\.)?[^.]+\\.[^.]+\\.[^.]+)\\.consul\\.)"
    tag_name: consul.full_target
  - tag_name: local_cluster
    fixed_value: {{.ServiceName}}
  - tag_name: consul.source.service
    fixed_value: {{.ServiceName}}
  - tag_name: consul.source.namespace
    fixed_value: default
  - tag_name: consul.source.partition
    fixed_value: default
  - tag_name: consul.source.datacenter
    fixed_value: dc1
  use_all_default_tags: true
