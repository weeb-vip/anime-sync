package anime_processor

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
	"testing"
	"time"
)

// TestTheTVDBIDIntegration focuses specifically on testing TheTVDBID functionality
func TestTheTVDBIDIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup database connection
	cfg := &config.DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "weeb",
		Password: "mysecretpassword",
		DataBase: "weeb",
		SSLMode:  "false",
	}

	database := db.NewDB(*cfg)
	require.NotNil(t, database)
	require.NotNil(t, database.DB)

	// Test database connection
	sqlDB, err := database.DB.DB()
	require.NoError(t, err)
	err = sqlDB.Ping()
	require.NoError(t, err, "Database should be accessible")

	// Clean up test data before and after
	cleanup := func() {
		database.DB.Where("id LIKE ?", "thetvdb-test-%").Delete(&anime.Anime{})
	}
	cleanup()
	defer cleanup()

	// Setup context with logger
	log := zap.NewNop()
	ctx := logger.WithCtx(context.Background(), log)

	// Create anime processor components
	repository := anime.NewAnimeRepository(database)
	processor := &AnimeProcessorImpl{Repository: repository}

	t.Run("TestParseToEntityWithTheTVDBID", func(t *testing.T) {
		thetvdbid := "987654"
		titleEn := "Test Anime with TheTVDBID"
		episodes := 12
		rating := "PG-13"
		startDate := "2024-01-15T09:00:00Z"
		endDate := "2024-03-30T21:00:00Z"

		// Test parsing schema to entity with TheTVDBID
		schema := Schema{
			ID:        "thetvdb-test-parse",
			TheTVDBID: &thetvdbid,
			TitleEn:   &titleEn,
			Episodes:  &episodes,
			Rating:    &rating,
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		animeEntity, err := processor.ParseToEntity(ctx, schema)
		require.NoError(t, err)
		require.NotNil(t, animeEntity)

		// Verify TheTVDBID was correctly parsed
		require.NotNil(t, animeEntity.TheTVDBID)
		assert.Equal(t, thetvdbid, *animeEntity.TheTVDBID)
		assert.Equal(t, "thetvdb-test-parse", animeEntity.ID)
		assert.Equal(t, titleEn, *animeEntity.TitleEn)
		assert.Equal(t, episodes, *animeEntity.Episodes)
		assert.Equal(t, rating, *animeEntity.Rating)

		// Verify date parsing
		assert.Contains(t, *animeEntity.StartDate, "2024-01-15")
		assert.Contains(t, *animeEntity.EndDate, "2024-03-30")
	})

	t.Run("TestParseToEntityWithoutTheTVDBID", func(t *testing.T) {
		titleEn := "Test Anime without TheTVDBID"

		// Test parsing schema to entity without TheTVDBID
		schema := Schema{
			ID:      "thetvdb-test-no-id",
			TitleEn: &titleEn,
		}

		animeEntity, err := processor.ParseToEntity(ctx, schema)
		require.NoError(t, err)
		require.NotNil(t, animeEntity)

		// Verify TheTVDBID is nil
		assert.Nil(t, animeEntity.TheTVDBID)
		assert.Equal(t, "thetvdb-test-no-id", animeEntity.ID)
		assert.Equal(t, titleEn, *animeEntity.TitleEn)
	})

	t.Run("TestDatabaseOperationsWithTheTVDBID", func(t *testing.T) {
		thetvdbid := "123456789"
		titleEn := "Database Test Anime"

		// Create anime entity with TheTVDBID
		animeEntity := &anime.Anime{
			ID:        "thetvdb-test-db-1",
			TheTVDBID: &thetvdbid,
			TitleEn:   &titleEn,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Test Upsert (create)
		err := repository.Upsert(animeEntity, nil)
		require.NoError(t, err)

		// Verify anime was saved with TheTVDBID
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "thetvdb-test-db-1").First(&savedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, thetvdbid, *savedAnime.TheTVDBID)
		assert.Equal(t, titleEn, *savedAnime.TitleEn)

		// Test Upsert (update) - change TheTVDBID
		newTheTVDBID := "987654321"
		animeEntity.TheTVDBID = &newTheTVDBID
		animeEntity.UpdatedAt = time.Now()

		err = repository.Upsert(animeEntity, nil)
		require.NoError(t, err)

		// Verify TheTVDBID was updated
		err = database.DB.Where("id = ?", "thetvdb-test-db-1").First(&savedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, newTheTVDBID, *savedAnime.TheTVDBID)

		// Test Delete
		err = repository.Delete(animeEntity)
		require.NoError(t, err)

		// Verify anime was deleted
		err = database.DB.Where("id = ?", "thetvdb-test-db-1").First(&savedAnime).Error
		assert.Error(t, err, "Anime should be deleted from database")
	})

	t.Run("TestTheTVDBIDQueryOperations", func(t *testing.T) {
		// Create multiple anime with different TheTVDBID scenarios
		testCases := []struct {
			id        string
			thetvdbid *string
			title     string
		}{
			{"thetvdb-test-query-1", stringPtr("111111"), "Anime 1 with TheTVDBID"},
			{"thetvdb-test-query-2", stringPtr("222222"), "Anime 2 with TheTVDBID"},
			{"thetvdb-test-query-3", nil, "Anime 3 without TheTVDBID"},
		}

		// Insert test data
		for _, tc := range testCases {
			animeEntity := &anime.Anime{
				ID:        tc.id,
				TheTVDBID: tc.thetvdbid,
				TitleEn:   &tc.title,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := repository.Upsert(animeEntity, nil)
			require.NoError(t, err)
		}

		// Test querying by TheTVDBID
		var animeWithTheTVDBID []anime.Anime
		err = database.DB.Where("thetvdbid IS NOT NULL AND id LIKE ?", "thetvdb-test-query-%").Find(&animeWithTheTVDBID).Error
		require.NoError(t, err)
		assert.Len(t, animeWithTheTVDBID, 2, "Should find 2 anime with TheTVDBID")

		// Test querying without TheTVDBID
		var animeWithoutTheTVDBID []anime.Anime
		err = database.DB.Where("thetvdbid IS NULL AND id LIKE ?", "thetvdb-test-query-%").Find(&animeWithoutTheTVDBID).Error
		require.NoError(t, err)
		assert.Len(t, animeWithoutTheTVDBID, 1, "Should find 1 anime without TheTVDBID")

		// Test querying by specific TheTVDBID
		var specificAnime anime.Anime
		err = database.DB.Where("thetvdbid = ?", "111111").First(&specificAnime).Error
		require.NoError(t, err)
		assert.Equal(t, "thetvdb-test-query-1", specificAnime.ID)
		assert.Equal(t, "Anime 1 with TheTVDBID", *specificAnime.TitleEn)
	})

	t.Run("TestTheTVDBIDValidation", func(t *testing.T) {
		// Test with various TheTVDBID values
		testCases := []struct {
			name      string
			thetvdbid *string
			shouldErr bool
		}{
			{"Valid TheTVDBID", stringPtr("123456"), false},
			{"Valid Long TheTVDBID", stringPtr("987654321012345"), false},
			{"Empty TheTVDBID", stringPtr(""), false}, // Empty string is valid
			{"Nil TheTVDBID", nil, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				titleEn := "Validation Test Anime"
				// Use shorter IDs to avoid MySQL column length issues
				shortName := map[string]string{
					"Valid TheTVDBID":      "val1",
					"Valid Long TheTVDBID": "val2",
					"Empty TheTVDBID":      "val3",
					"Nil TheTVDBID":        "val4",
				}[tc.name]
				animeEntity := &anime.Anime{
					ID:        "thetvdb-val-" + shortName,
					TheTVDBID: tc.thetvdbid,
					TitleEn:   &titleEn,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				err := repository.Upsert(animeEntity, nil)
				if tc.shouldErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)

					// Verify the value was stored correctly
					var savedAnime anime.Anime
					err = database.DB.Where("id = ?", animeEntity.ID).First(&savedAnime).Error
					require.NoError(t, err)

					if tc.thetvdbid == nil {
						assert.Nil(t, savedAnime.TheTVDBID)
					} else {
						require.NotNil(t, savedAnime.TheTVDBID)
						assert.Equal(t, *tc.thetvdbid, *savedAnime.TheTVDBID)
					}
				}
			})
		}
	})

	t.Run("TestTheTVDBIDUpdateScenarios", func(t *testing.T) {
		baseID := "thetvdb-test-update-scenarios"

		// Scenario 1: Add TheTVDBID to existing anime without one
		titleEn := "Update Scenario Test"
		animeWithoutTheTVDBID := &anime.Anime{
			ID:        baseID,
			TitleEn:   &titleEn,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repository.Upsert(animeWithoutTheTVDBID, nil)
		require.NoError(t, err)

		// Verify no TheTVDBID initially
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", baseID).First(&savedAnime).Error
		require.NoError(t, err)
		assert.Nil(t, savedAnime.TheTVDBID)

		// Add TheTVDBID
		thetvdbid := "new-thetvdbid-123"
		animeWithoutTheTVDBID.TheTVDBID = &thetvdbid
		animeWithoutTheTVDBID.UpdatedAt = time.Now()

		err = repository.Upsert(animeWithoutTheTVDBID, nil)
		require.NoError(t, err)

		// Verify TheTVDBID was added
		err = database.DB.Where("id = ?", baseID).First(&savedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, thetvdbid, *savedAnime.TheTVDBID)

		// Scenario 2: Remove TheTVDBID (set to nil)
		animeWithoutTheTVDBID.TheTVDBID = nil
		animeWithoutTheTVDBID.UpdatedAt = time.Now()

		err = repository.Upsert(animeWithoutTheTVDBID, nil)
		require.NoError(t, err)

		// Verify TheTVDBID was removed
		err = database.DB.Where("id = ?", baseID).First(&savedAnime).Error
		require.NoError(t, err)
		assert.Nil(t, savedAnime.TheTVDBID)
	})
}

// TestCompleteProcessorWorkflowWithTheTVDBID tests the full processor workflow through Process() method
func TestCompleteProcessorWorkflowWithTheTVDBID(t *testing.T) {
	// Commented out due to Flagsmith type casting complexity
	// Use TestAnimeProcessorWorkflowWithTheTVDBID in processor_workflow_test.go which bypasses Flagsmith
	t.Skip("Skipping - replaced by TestAnimeProcessorWorkflowWithTheTVDBID with working test processor")
}

// TestProcessorCoreLogicWithTheTVDBID tests processor logic without Flagsmith dependency
func TestProcessorCoreLogicWithTheTVDBID(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup database connection
	cfg := &config.DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "weeb",
		Password: "mysecretpassword",
		DataBase: "weeb",
		SSLMode:  "false",
	}

	database := db.NewDB(*cfg)
	require.NotNil(t, database)

	// Clean up test data
	cleanup := func() {
		database.DB.Where("id LIKE ?", "core-test-%").Delete(&anime.Anime{})
	}
	cleanup()
	defer cleanup()

	// Setup context with logger
	log := zap.NewNop()
	ctx := logger.WithCtx(context.Background(), log)

	// Create processor implementation directly
	repository := anime.NewAnimeRepository(database)
	processor := &AnimeProcessorImpl{
		Repository: repository,
		Options:    Options{NoErrorOnDelete: false},
	}

	t.Run("TestCreateLogicWithTheTVDBID", func(t *testing.T) {
		thetvdbid := "core-create-789"
		titleEn := "Core Logic Test Anime"
		episodes := 12

		// Test the core create logic by calling parseToEntity and repository directly
		schema := Schema{
			ID:        "core-test-create",
			TheTVDBID: &thetvdbid,
			TitleEn:   &titleEn,
			Episodes:  &episodes,
		}

		// Test parsing
		animeEntity, err := processor.ParseToEntity(ctx, schema)
		require.NoError(t, err)
		require.NotNil(t, animeEntity.TheTVDBID)
		assert.Equal(t, thetvdbid, *animeEntity.TheTVDBID)

		// Test repository upsert
		err = processor.Repository.Upsert(animeEntity, nil)
		require.NoError(t, err)

		// Verify in database
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "core-test-create").First(&savedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, thetvdbid, *savedAnime.TheTVDBID)
		assert.Equal(t, titleEn, *savedAnime.TitleEn)
		assert.Equal(t, episodes, *savedAnime.Episodes)
	})

	t.Run("TestUpdateLogicWithTheTVDBID", func(t *testing.T) {
		// First create an anime
		initialTheTVDBID := "core-update-111"
		titleEn := "Update Logic Test"

		initialEntity := &anime.Anime{
			ID:        "core-test-update",
			TheTVDBID: &initialTheTVDBID,
			TitleEn:   &titleEn,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := processor.Repository.Upsert(initialEntity, nil)
		require.NoError(t, err)

		// Now test update with new TheTVDBID
		newTheTVDBID := "core-update-222"
		updatedTitle := "Updated Core Logic Test"

		updateSchema := Schema{
			ID:        "core-test-update",
			TheTVDBID: &newTheTVDBID,
			TitleEn:   &updatedTitle,
		}

		// Test parsing update
		updatedEntity, err := processor.ParseToEntity(ctx, updateSchema)
		require.NoError(t, err)
		require.NotNil(t, updatedEntity.TheTVDBID)
		assert.Equal(t, newTheTVDBID, *updatedEntity.TheTVDBID)

		// Test repository update
		err = processor.Repository.Upsert(updatedEntity, &titleEn) // oldTitle
		require.NoError(t, err)

		// Verify update in database
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "core-test-update").First(&savedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, newTheTVDBID, *savedAnime.TheTVDBID)
		assert.Equal(t, updatedTitle, *savedAnime.TitleEn)
	})

	t.Run("TestDeleteLogicWithTheTVDBID", func(t *testing.T) {
		// Create anime to delete
		thetvdbid := "core-delete-333"
		titleEn := "Delete Logic Test"

		deleteEntity := &anime.Anime{
			ID:        "core-test-delete",
			TheTVDBID: &thetvdbid,
			TitleEn:   &titleEn,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := processor.Repository.Upsert(deleteEntity, nil)
		require.NoError(t, err)

		// Verify it exists
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "core-test-delete").First(&savedAnime).Error
		require.NoError(t, err)

		// Test delete
		err = processor.Repository.Delete(deleteEntity)
		require.NoError(t, err)

		// Verify deletion
		err = database.DB.Where("id = ?", "core-test-delete").First(&savedAnime).Error
		assert.Error(t, err, "Anime should be deleted from database")
	})

	t.Run("TestComplexTheTVDBIDScenario", func(t *testing.T) {
		// Test a complex scenario with full anime data including TheTVDBID
		thetvdbid := "core-complex-456789"
		titleEn := "Complex Core Test Anime"
		titleJp := "コンプレックスコアテスト"
		titleRomaji := "Complex Core Test"
		synopsis := "A comprehensive test anime with TheTVDBID for core logic testing"
		episodes := 26
		rating := "R - 17+"
		genres := "Action, Drama, Sci-Fi"
		startDate := "2024-04-01T00:00:00Z"
		endDate := "2024-09-30T23:59:59Z"

		complexSchema := Schema{
			ID:          "core-test-complex",
			TheTVDBID:   &thetvdbid,
			TitleEn:     &titleEn,
			TitleJp:     &titleJp,
			TitleRomaji: &titleRomaji,
			Synopsis:    &synopsis,
			Episodes:    &episodes,
			Rating:      &rating,
			Genres:      &genres,
			StartDate:   &startDate,
			EndDate:     &endDate,
		}

		// Test parsing complex schema
		complexEntity, err := processor.ParseToEntity(ctx, complexSchema)
		require.NoError(t, err)

		// Verify TheTVDBID and all other fields
		require.NotNil(t, complexEntity.TheTVDBID)
		assert.Equal(t, thetvdbid, *complexEntity.TheTVDBID)
		assert.Equal(t, titleEn, *complexEntity.TitleEn)
		assert.Equal(t, titleJp, *complexEntity.TitleJp)
		assert.Equal(t, titleRomaji, *complexEntity.TitleRomaji)
		assert.Equal(t, synopsis, *complexEntity.Synopsis)
		assert.Equal(t, episodes, *complexEntity.Episodes)
		assert.Equal(t, rating, *complexEntity.Rating)
		assert.Equal(t, genres, *complexEntity.Genres)

		// Verify date parsing
		assert.Contains(t, *complexEntity.StartDate, "2024-04-01")
		assert.Contains(t, *complexEntity.EndDate, "2024-09-30")

		// Test saving complex entity
		err = processor.Repository.Upsert(complexEntity, nil)
		require.NoError(t, err)

		// Verify everything was saved correctly
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "core-test-complex").First(&savedAnime).Error
		require.NoError(t, err)

		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, thetvdbid, *savedAnime.TheTVDBID)
		assert.Equal(t, titleEn, *savedAnime.TitleEn)
		assert.Equal(t, titleJp, *savedAnime.TitleJp)
		assert.Equal(t, titleRomaji, *savedAnime.TitleRomaji)
		assert.Equal(t, synopsis, *savedAnime.Synopsis)
		assert.Equal(t, episodes, *savedAnime.Episodes)
		assert.Equal(t, rating, *savedAnime.Rating)
		assert.Equal(t, genres, *savedAnime.Genres)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
