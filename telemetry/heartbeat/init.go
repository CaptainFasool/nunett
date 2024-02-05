package heartbeat

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/telemetry/logger"
)

var (
	zlog            otelzap.Logger
	elastictoken    models.ElasticToken
	esClient        *elasticsearch.Client
	esClientHealthy bool
)

func init() {
	zlog = logger.OtelZapLogger("telemetry/heartbeat")
}
