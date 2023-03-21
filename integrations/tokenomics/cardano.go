package tokenomics

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/integrations/oracle"
)

type ClaimCardanoTokenBody struct {
	ComputeProviderAddress string `json:"compute_provider_address"`
}

// HandleRequestReward  godoc
// @Summary      Get NTX tokens for work done.
// @Description  HandleRequestReward takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.
// @Router       /run/request-reward [post]
func HandleRequestReward(c *gin.Context) {
	rand.Seed(time.Now().Unix())

	body := ClaimCardanoTokenBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// TODO: Fetch this data and pass to the function below
	// JobStatus:            "finished without errors",
	// JobDuration:          5,
	// EstimatedJobDuration: 10,
	// LogPath:              "https://gist.github.com/santosh/42e86f264c89be54e3351e2373c92edf",

	resp, err := oracle.WithdrawTokenRequest()

	if err != nil {
		c.JSON(500, gin.H{"message": "some error has occured"})
	}

	c.JSON(200, resp)
}
