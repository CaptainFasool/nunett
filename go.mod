module gitlab.com/nunet/device-management-service

go 1.19

require (
	github.com/KyleBanks/dockerstats v0.0.0-20180213183355-b5fec062e953
	github.com/docker/cli v20.10.23+incompatible
	github.com/docker/docker v20.10.18+incompatible
	github.com/ethereum/go-ethereum v1.10.18
	github.com/fivebinaries/go-cardano-serialization v0.0.0-20220907134105-ec9b85086588
	github.com/gin-contrib/cors v1.4.0
	github.com/gin-gonic/gin v1.8.2
	github.com/google/go-github v17.0.0+incompatible
	github.com/gorilla/websocket v1.5.0
	github.com/joho/godotenv v1.4.0
	github.com/libp2p/go-libp2p v0.24.2
	github.com/libp2p/go-libp2p-core v0.20.1
	github.com/libp2p/go-libp2p-kad-dht v0.20.0
	github.com/multiformats/go-multiaddr v0.8.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/spf13/afero v1.9.3
	github.com/stretchr/testify v1.8.2
	github.com/swaggo/files v0.0.0-20220728132757-551d4a08d97a
	github.com/swaggo/gin-swagger v1.3.2
	github.com/swaggo/swag v1.8.1
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/uptrace/opentelemetry-go-extra/otelgorm v0.1.21
	github.com/whyrusleeping/multiaddr-filter v0.0.0-20160516205228-e903e4adabd7
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.39.0
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/exporters/jaeger v1.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.13.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.13.0
	go.opentelemetry.io/otel/sdk v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
	go.uber.org/zap v1.24.0
	golang.org/x/oauth2 v0.4.0
	google.golang.org/grpc v1.52.3
	google.golang.org/protobuf v1.28.1
	gorm.io/driver/sqlite v1.3.6
	gorm.io/gorm v1.24.5
)

replace gorm.io/gorm => gorm.io/gorm v1.23.8

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.2 // indirect
	github.com/btcsuite/btcutil v1.0.2 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/elastic/gosigar v0.14.2 // indirect
	github.com/flynn/noise v1.0.0 // indirect
	github.com/francoispqt/gojay v1.2.13 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.7 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/pprof v0.0.0-20221203041831-ce31453925ec // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/huin/goupnp v1.0.3 // indirect
	github.com/ipfs/go-cid v0.3.2 // indirect
	github.com/ipfs/go-datastore v0.6.0 // indirect
	github.com/ipfs/go-ipfs-util v0.0.2 // indirect
	github.com/ipfs/go-ipns v0.2.0 // indirect
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/ipld/go-ipld-prime v0.9.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jbenet/go-temp-err-catcher v0.1.0 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.1 // indirect
	github.com/koron/go-ssdp v0.0.3 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-cidranger v1.1.0 // indirect
	github.com/libp2p/go-flow-metrics v0.1.0 // indirect
	github.com/libp2p/go-libp2p-asn-util v0.2.0 // indirect
	github.com/libp2p/go-libp2p-kbucket v0.5.0 // indirect
	github.com/libp2p/go-libp2p-record v0.2.0 // indirect
	github.com/libp2p/go-msgio v0.2.0 // indirect
	github.com/libp2p/go-nat v0.1.0 // indirect
	github.com/libp2p/go-netroute v0.2.1 // indirect
	github.com/libp2p/go-openssl v0.1.0 // indirect
	github.com/libp2p/go-reuseport v0.2.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.0 // indirect
	github.com/lucas-clemente/quic-go v0.31.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/marten-seemann/qtls-go1-18 v0.1.3 // indirect
	github.com/marten-seemann/qtls-go1-19 v0.1.1 // indirect
	github.com/marten-seemann/tcp v0.0.0-20210406111302-dfbc87cc63fd // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mikioh/tcpinfo v0.0.0-20190314235526-30a79bb1804b // indirect
	github.com/mikioh/tcpopt v0.0.0-20190314235656-172688c1accc // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr-dns v0.3.1 // indirect
	github.com/multiformats/go-multiaddr-fmt v0.1.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multicodec v0.7.0 // indirect
	github.com/multiformats/go-multihash v0.2.1 // indirect
	github.com/multiformats/go-multistream v0.3.3 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/onsi/ginkgo/v2 v2.5.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runtime-spec v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/polydawn/refmt v0.0.0-20190807091052-3d65705ee9f1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/raulk/go-watchdog v1.3.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.1.21 // indirect
	github.com/whyrusleeping/go-keyspace v0.0.0-20160322163242-5b898ac5add1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.13.0 // indirect
	go.opentelemetry.io/otel/metric v0.36.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/dig v1.15.0 // indirect
	go.uber.org/fx v1.18.2 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	golang.org/x/tools v0.4.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.3.0 // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
)

require github.com/NVIDIA/go-nvml v0.12.0-1
