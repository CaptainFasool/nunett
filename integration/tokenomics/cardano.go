package tokenomics

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ClaimCardanoTokenBody struct {
	ComputeProviderAddress string `json:"compute_provider_address"`
}

// HandleClaimCardanoTokens  godoc
// @Summary      Get NTX tokens for work done.
// @Description  HandleClaimCardanoTokens takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.
// @Router       /run/claim [get]
func HandleClaimCardanoTokens(c *gin.Context) {
	// TODO: This is a stub function. Replace the logic to talk with Oracle.
	rand.Seed(time.Now().Unix())

	body := ClaimCardanoTokenBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// body.ComputeProviderAddress

	rewardTypes := []string{"withdraw", "refund", "distribute"}
	randomRewardType := rewardTypes[rand.Intn(len(rewardTypes))]

	c.JSON(200, gin.H{"hash": "some hash data", "reward_type": randomRewardType})
}
