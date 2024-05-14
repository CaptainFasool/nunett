package repositories_clover

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/nunet/device-management-service/models"
)

// TestLogbinAuthRepository is a test suite for the LogbinAuthRepository.
// It includes test cases that cover the basic CRUD operations and custom repository functions if there are any.
// This test suite ensures that the repository functions for the LogbinAuth model behave as expected.
func TestLogBinAuthRepository(t *testing.T) {
	// Setup your database connection for testing
	db, path := setup()
	defer teardown(db, path)

	// Initialize the repository
	logBinAuthRepo := NewLogBinAuthRepository(db)

	// Test Save method
	createdLogBinAuth, err := logBinAuthRepo.Save(
		context.Background(),
		models.LogBinAuth{Token: uuid.New().String()},
	)
	assert.NoError(t, err)
	assert.NotZero(t, createdLogBinAuth.Token)

	// Test Get method
	retrievedLogBinAuth, err := logBinAuthRepo.Get(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, createdLogBinAuth.Token, retrievedLogBinAuth.Token)

	// Test Save (update) method
	updatedLogBinAuth := retrievedLogBinAuth
	updatedLogBinAuth.Token = uuid.New().String()

	_, err = logBinAuthRepo.Save(context.Background(), updatedLogBinAuth)
	assert.NoError(t, err)
	retrievedLogBinAuth, err = logBinAuthRepo.Get(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, updatedLogBinAuth.Token, retrievedLogBinAuth.Token)

	// Test History method
	query := logBinAuthRepo.GetQuery()
	query.Limit = 3
	history, err := logBinAuthRepo.History(context.Background(), query)
	assert.NoError(t, err)
	assert.Len(t, history, 2)

	// Test Clear method
	err = logBinAuthRepo.Clear(context.Background())
	assert.NoError(t, err)
	history, err = logBinAuthRepo.History(context.Background(), query)
	assert.Len(t, history, 0)
}
