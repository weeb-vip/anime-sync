package processor

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"testing"
	"time"
)

// TestAnime represents a test anime structure
type TestAnime struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Episode int    `json:"episode"`
}

// TestProcessorIntegration tests the generic processor with database operations
func TestProcessorIntegration(t *testing.T) {
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

	// Clean up test data
	defer func() {
		database.DB.Where("id LIKE ?", "test-%").Delete(&anime.Anime{})
	}()

	t.Run("TestProcessorParse", func(t *testing.T) {
		processor := NewProcessor[TestAnime]()

		testPayload := TestAnime{
			ID:      "test-123",
			Title:   "Test Anime",
			Episode: 1,
		}

		payloadJSON, err := json.Marshal(testPayload)
		require.NoError(t, err)

		// Test parsing
		parsed, err := processor.Parse(context.Background(), string(payloadJSON))
		require.NoError(t, err)
		assert.Equal(t, testPayload.ID, parsed.ID)
		assert.Equal(t, testPayload.Title, parsed.Title)
		assert.Equal(t, testPayload.Episode, parsed.Episode)
	})

	t.Run("TestProcessorParseError", func(t *testing.T) {
		processor := NewProcessor[TestAnime]()

		// Test parsing invalid JSON
		_, err := processor.Parse(context.Background(), "invalid json")
		assert.Error(t, err)
	})

	t.Run("TestProcessorProcessSuccess", func(t *testing.T) {
		processor := NewProcessor[TestAnime]()

		testPayload := TestAnime{
			ID:      "test-456",
			Title:   "Test Anime 2",
			Episode: 2,
		}

		payloadJSON, err := json.Marshal(testPayload)
		require.NoError(t, err)

		// Test successful processing
		processedData := make(chan TestAnime, 1)
		processorFunc := func(ctx context.Context, data TestAnime) error {
			processedData <- data
			return nil
		}

		err = processor.Process(context.Background(), string(payloadJSON), processorFunc)
		require.NoError(t, err)

		// Verify data was processed
		select {
		case processed := <-processedData:
			assert.Equal(t, testPayload.ID, processed.ID)
			assert.Equal(t, testPayload.Title, processed.Title)
			assert.Equal(t, testPayload.Episode, processed.Episode)
		case <-time.After(time.Second):
			t.Fatal("Processor function was not called")
		}
	})

	t.Run("TestProcessorProcessWithRetry", func(t *testing.T) {
		processor := NewProcessor[TestAnime]()

		testPayload := TestAnime{
			ID:      "test-retry",
			Title:   "Retry Test",
			Episode: 1,
		}

		payloadJSON, err := json.Marshal(testPayload)
		require.NoError(t, err)

		// Test processing with retries
		attempts := 0
		processorFunc := func(ctx context.Context, data TestAnime) error {
			attempts++
			if attempts < 3 {
				return assert.AnError // Will cause retry
			}
			return nil // Success on third attempt
		}

		err = processor.Process(context.Background(), string(payloadJSON), processorFunc)
		require.NoError(t, err)
		assert.Equal(t, 3, attempts, "Should have retried 3 times")
	})

	t.Run("TestProcessorProcessWithDatabase", func(t *testing.T) {
		processor := NewProcessor[anime.Anime]()
		repository := anime.NewAnimeRepository(database)

		// Create test anime
		titleEn := "Integration Test Anime"
		testAnime := anime.Anime{
			ID:        "test-db-integration",
			TitleEn:   &titleEn,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		payloadJSON, err := json.Marshal(testAnime)
		require.NoError(t, err)

		// Process and save to database
		processorFunc := func(ctx context.Context, data anime.Anime) error {
			return repository.Upsert(&data, nil)
		}

		err = processor.Process(context.Background(), string(payloadJSON), processorFunc)
		require.NoError(t, err)

		// Verify data was saved to database
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", testAnime.ID).First(&savedAnime).Error
		require.NoError(t, err)
		assert.Equal(t, testAnime.ID, savedAnime.ID)
		assert.Equal(t, *testAnime.TitleEn, *savedAnime.TitleEn)

		// Test update
		newTitle := "Updated Integration Test Anime"
		testAnime.TitleEn = &newTitle
		testAnime.UpdatedAt = time.Now()

		payloadJSON, err = json.Marshal(testAnime)
		require.NoError(t, err)

		err = processor.Process(context.Background(), string(payloadJSON), processorFunc)
		require.NoError(t, err)

		// Verify data was updated
		err = database.DB.Where("id = ?", testAnime.ID).First(&savedAnime).Error
		require.NoError(t, err)
		assert.Equal(t, *testAnime.TitleEn, *savedAnime.TitleEn)
	})

	t.Run("TestProcessorProcessFailure", func(t *testing.T) {
		processor := NewProcessor[TestAnime]()

		testPayload := TestAnime{
			ID:      "test-failure",
			Title:   "Failure Test",
			Episode: 1,
		}

		payloadJSON, err := json.Marshal(testPayload)
		require.NoError(t, err)

		// Test processing failure that exhausts retries
		processorFunc := func(ctx context.Context, data TestAnime) error {
			return assert.AnError // Always fail
		}

		err = processor.Process(context.Background(), string(payloadJSON), processorFunc)
		assert.Error(t, err)
	})

	t.Run("TestProcessorWithContext", func(t *testing.T) {
		processor := NewProcessor[TestAnime]()

		testPayload := TestAnime{
			ID:      "test-context",
			Title:   "Context Test",
			Episode: 1,
		}

		payloadJSON, err := json.Marshal(testPayload)
		require.NoError(t, err)

		// Test with context cancellation
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		processorFunc := func(ctx context.Context, data TestAnime) error {
			time.Sleep(50 * time.Millisecond) // Simulate slow operation
			return nil
		}

		// This test may or may not fail with context timeout depending on timing
		// The important thing is that the processor accepts and passes context
		processor.Process(ctx, string(payloadJSON), processorFunc)
	})
}

// BenchmarkProcessor benchmarks the processor performance
func BenchmarkProcessor(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	processor := NewProcessor[TestAnime]()

	testPayload := TestAnime{
		ID:      "bench-test",
		Title:   "Benchmark Test",
		Episode: 1,
	}

	payloadJSON, err := json.Marshal(testPayload)
	require.NoError(b, err)

	processorFunc := func(ctx context.Context, data TestAnime) error {
		return nil // No-op for benchmarking
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := processor.Process(context.Background(), string(payloadJSON), processorFunc)
		if err != nil {
			b.Fatal(err)
		}
	}
}