package anime_processor_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ThatCatDev/ep/v2/event"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/services/anime_processor"
	"github.com/weeb-vip/anime-sync/internal/services/anime_processor/mocks"
)

// TestRealProcessorWorkflowWithMocks tests the actual processor with proper mocks
func TestRealProcessorWorkflowWithMocks(t *testing.T) {
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

	// Test database connection
	sqlDB, err := database.DB.DB()
	require.NoError(t, err)
	err = sqlDB.Ping()
	require.NoError(t, err, "Database should be accessible")

	// Clean up test data before and after
	cleanup := func() {
		database.DB.Where("id LIKE ?", "real-proc-%").Delete(&anime.Anime{})
	}
	cleanup()
	defer cleanup()

	// Setup mock producers to capture calls
	var algoliaMessages []*kafka.Message
	var kafkaMessages []*kafka.Message

	algoliaProducer := func(ctx context.Context, message *kafka.Message) error {
		algoliaMessages = append(algoliaMessages, message)
		return nil
	}

	kafkaProducer := func(ctx context.Context, message *kafka.Message) error {
		kafkaMessages = append(kafkaMessages, message)
		return nil
	}

	// Create real processor
	options := anime_processor.Options{NoErrorOnDelete: false}
	processor := anime_processor.NewAnimeProcessor(options, database, algoliaProducer, kafkaProducer)

	// Setup context with logger and mock flagsmith client
	log := zap.NewNop()
	ctx := logger.WithCtx(context.Background(), log)

	// Create mock Flagsmith client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlagsmith := mocks.NewMockFlagSmithClient(ctrl)
	mockFlags := mocks.NewMockFlags(ctrl)

	// Setup expectations - allow any number of calls
	mockFlagsmith.EXPECT().GetEnvironmentFlags().Return(mockFlags, nil).AnyTimes()
	mockFlags.EXPECT().IsFeatureEnabled("enable_kafka").Return(true, nil).AnyTimes()

	ctx = context.WithValue(ctx, internal.FFClient{}, mockFlagsmith)

	t.Run("TestRealProcessorCreateWithTheTVDBID", func(t *testing.T) {
		// Reset message collectors
		algoliaMessages = nil
		kafkaMessages = nil

		thetvdbid := "real-proc-create-123"
		titleEn := "Real Processor Create Test"
		imageUrl := "https://example.com/real-create.jpg"
		episodes := 12
		synopsis := "Test anime creation with real processor"

		// Create payload for CREATE operation
		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:        "real-proc-create",
				TheTVDBID: &thetvdbid,
				TitleEn:   &titleEn,
				ImageUrl:  &imageUrl,
				Episodes:  &episodes,
				Synopsis:  &synopsis,
			},
			Source: anime_processor.Source{
				Version: "1.0",
				TsMs:    time.Now().UnixMilli(),
			},
		}

		eventData := event.Event[*kafka.Message, anime_processor.Payload]{
			Payload: payload,
		}

		// Process through real processor
		result, err := processor.Process(ctx, eventData)
		require.NoError(t, err)
		assert.Equal(t, payload, result.Payload)

		// Verify anime was saved to database with TheTVDBID
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "real-proc-create").First(&savedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, thetvdbid, *savedAnime.TheTVDBID)
		assert.Equal(t, titleEn, *savedAnime.TitleEn)
		assert.Equal(t, episodes, *savedAnime.Episodes)
		assert.Equal(t, synopsis, *savedAnime.Synopsis)

		// Verify producers were called
		assert.Len(t, algoliaMessages, 1, "Algolia producer should be called once")
		assert.Len(t, kafkaMessages, 1, "Kafka producer should be called once for image")

		// Verify producer message content
		var producerPayload anime_processor.ProducerPayload
		err = json.Unmarshal(algoliaMessages[0].Value, &producerPayload)
		require.NoError(t, err)
		assert.Equal(t, anime_processor.CreateAction, producerPayload.Action)
		require.NotNil(t, producerPayload.Data.TheTVDBID)
		assert.Equal(t, thetvdbid, *producerPayload.Data.TheTVDBID)
	})

	t.Run("TestRealProcessorUpdateWithTheTVDBID", func(t *testing.T) {
		// Reset message collectors
		algoliaMessages = nil
		kafkaMessages = nil

		// First create an anime
		initialTheTVDBID := "real-proc-initial-456"
		initialTitle := "Real Processor Initial"
		createPayload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:        "real-proc-update",
				TheTVDBID: &initialTheTVDBID,
				TitleEn:   &initialTitle,
			},
		}
		createEvent := event.Event[*kafka.Message, anime_processor.Payload]{Payload: createPayload}
		_, err := processor.Process(ctx, createEvent)
		require.NoError(t, err)

		// Reset collectors after create
		algoliaMessages = nil
		kafkaMessages = nil

		// Now update with new TheTVDBID
		newTheTVDBID := "real-proc-updated-789"
		newTitle := "Real Processor Updated"
		updatePayload := anime_processor.Payload{
			Before: &anime_processor.Schema{
				ID:        "real-proc-update",
				TheTVDBID: &initialTheTVDBID,
				TitleEn:   &initialTitle,
			},
			After: &anime_processor.Schema{
				ID:        "real-proc-update",
				TheTVDBID: &newTheTVDBID,
				TitleEn:   &newTitle,
			},
		}
		updateEvent := event.Event[*kafka.Message, anime_processor.Payload]{Payload: updatePayload}

		// Process the update through real processor
		result, err := processor.Process(ctx, updateEvent)
		require.NoError(t, err)
		assert.Equal(t, updatePayload, result.Payload)

		// Verify anime was updated with new TheTVDBID
		var updatedAnime anime.Anime
		err = database.DB.Where("id = ?", "real-proc-update").First(&updatedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, updatedAnime.TheTVDBID)
		assert.Equal(t, newTheTVDBID, *updatedAnime.TheTVDBID)
		assert.Equal(t, newTitle, *updatedAnime.TitleEn)

		// Verify producers were called for update
		assert.Len(t, algoliaMessages, 1, "Algolia producer should be called once for update")
	})

	t.Run("TestRealProcessorDeleteWithTheTVDBID", func(t *testing.T) {
		// Reset message collectors
		algoliaMessages = nil
		kafkaMessages = nil

		// First create an anime to delete
		thetvdbid := "real-proc-delete-999"
		titleEn := "Real Processor Delete Test"
		createPayload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:        "real-proc-delete",
				TheTVDBID: &thetvdbid,
				TitleEn:   &titleEn,
			},
		}
		createEvent := event.Event[*kafka.Message, anime_processor.Payload]{Payload: createPayload}
		_, err := processor.Process(ctx, createEvent)
		require.NoError(t, err)

		// Verify anime exists
		var createdAnime anime.Anime
		err = database.DB.Where("id = ?", "real-proc-delete").First(&createdAnime).Error
		require.NoError(t, err)
		require.NotNil(t, createdAnime.TheTVDBID)

		// Reset collectors after create
		algoliaMessages = nil
		kafkaMessages = nil

		// Now delete the anime
		deletePayload := anime_processor.Payload{
			Before: &anime_processor.Schema{
				ID:        "real-proc-delete",
				TheTVDBID: &thetvdbid,
				TitleEn:   &titleEn,
			},
			After: nil, // Delete operation
		}
		deleteEvent := event.Event[*kafka.Message, anime_processor.Payload]{Payload: deletePayload}

		// Process the delete through real processor
		result, err := processor.Process(ctx, deleteEvent)
		require.NoError(t, err)
		assert.Equal(t, deletePayload, result.Payload)

		// Verify anime was deleted
		var deletedAnime anime.Anime
		err = database.DB.Where("id = ?", "real-proc-delete").First(&deletedAnime).Error
		assert.Error(t, err, "Anime should be deleted from database")
	})

	t.Run("TestRealProcessorWithoutImageURL", func(t *testing.T) {
		// Reset message collectors
		algoliaMessages = nil
		kafkaMessages = nil

		thetvdbid := "real-proc-no-image-111"
		titleEn := "Real Processor No Image"

		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:        "real-proc-no-image",
				TheTVDBID: &thetvdbid,
				TitleEn:   &titleEn,
				// ImageUrl is nil
			},
		}
		eventData := event.Event[*kafka.Message, anime_processor.Payload]{Payload: payload}

		// Process through real processor
		result, err := processor.Process(ctx, eventData)
		require.NoError(t, err)
		assert.Equal(t, payload, result.Payload)

		// Verify anime was saved with TheTVDBID
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "real-proc-no-image").First(&savedAnime).Error
		require.NoError(t, err)
		require.NotNil(t, savedAnime.TheTVDBID)
		assert.Equal(t, thetvdbid, *savedAnime.TheTVDBID)

		// Verify producer behavior - Algolia should be called but not Kafka for image
		assert.Len(t, algoliaMessages, 1, "Algolia producer should be called once")
		assert.Len(t, kafkaMessages, 0, "Kafka producer should not be called when no image")
	})

	t.Run("TestRealProcessorWithoutTheTVDBID", func(t *testing.T) {
		// Reset message collectors
		algoliaMessages = nil
		kafkaMessages = nil

		titleEn := "Real Processor No TheTVDBID"
		imageUrl := "https://example.com/no-thetvdbid.jpg"
		episodes := 24
		synopsis := "Test anime creation without TheTVDBID"

		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:       "real-proc-no-thetvdbid",
				TitleEn:  &titleEn,
				ImageUrl: &imageUrl,
				Episodes: &episodes,
				Synopsis: &synopsis,
				// TheTVDBID is nil
			},
			Source: anime_processor.Source{
				Version: "1.0",
				TsMs:    time.Now().UnixMilli(),
			},
		}
		eventData := event.Event[*kafka.Message, anime_processor.Payload]{Payload: payload}

		// Process through real processor
		result, err := processor.Process(ctx, eventData)
		require.NoError(t, err)
		assert.Equal(t, payload, result.Payload)

		// Verify anime was saved without TheTVDBID
		var savedAnime anime.Anime
		err = database.DB.Where("id = ?", "real-proc-no-thetvdbid").First(&savedAnime).Error
		require.NoError(t, err)
		assert.Nil(t, savedAnime.TheTVDBID, "TheTVDBID should be nil when not provided")
		assert.Equal(t, titleEn, *savedAnime.TitleEn)
		assert.Equal(t, episodes, *savedAnime.Episodes)
		assert.Equal(t, synopsis, *savedAnime.Synopsis)

		// Verify producers were called (both Algolia and Kafka for image)
		assert.Len(t, algoliaMessages, 1, "Algolia producer should be called once")
		assert.Len(t, kafkaMessages, 1, "Kafka producer should be called once for image")

		// Verify producer message content - TheTVDBID should be nil in payload
		var producerPayload anime_processor.ProducerPayload
		err = json.Unmarshal(algoliaMessages[0].Value, &producerPayload)
		require.NoError(t, err)
		assert.Equal(t, anime_processor.CreateAction, producerPayload.Action)
		assert.Nil(t, producerPayload.Data.TheTVDBID, "TheTVDBID should be nil in producer payload")
	})
}

// TestSyncTagsWithJSONGenres tests the syncTags functionality with JSON array genres
func TestSyncTagsWithJSONGenres(t *testing.T) {
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

	sqlDB, err := database.DB.DB()
	require.NoError(t, err)
	err = sqlDB.Ping()
	require.NoError(t, err, "Database should be accessible")

	// Clean up test data before and after
	cleanup := func() {
		database.DB.Exec("DELETE FROM anime_tags WHERE anime_id LIKE ?", "sync-tags-test-%")
		database.DB.Exec("DELETE FROM tags WHERE name IN (?, ?, ?, ?, ?)", "Drama", "Supernatural", "Suspense", "Action", "Comedy")
		database.DB.Where("id LIKE ?", "sync-tags-test-%").Delete(&anime.Anime{})
	}
	cleanup()
	defer cleanup()

	// Setup mock producers
	algoliaProducer := func(ctx context.Context, message *kafka.Message) error {
		return nil
	}
	kafkaProducer := func(ctx context.Context, message *kafka.Message) error {
		return nil
	}

	// Create processor
	options := anime_processor.Options{NoErrorOnDelete: false}
	processor := anime_processor.NewAnimeProcessor(options, database, algoliaProducer, kafkaProducer)

	// Setup context with mock flagsmith client
	log := zap.NewNop()
	ctx := logger.WithCtx(context.Background(), log)

	// Create mock Flagsmith client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlagsmith := mocks.NewMockFlagSmithClient(ctrl)
	mockFlags := mocks.NewMockFlags(ctrl)

	// Setup expectations - allow any number of calls
	mockFlagsmith.EXPECT().GetEnvironmentFlags().Return(mockFlags, nil).AnyTimes()
	mockFlags.EXPECT().IsFeatureEnabled("enable_kafka").Return(true, nil).AnyTimes()

	ctx = context.WithValue(ctx, internal.FFClient{}, mockFlagsmith)

	t.Run("CreateAnimeWithJSONGenres", func(t *testing.T) {
		titleEn := "Sync Tags Test Anime"
		genres := `["Drama","Supernatural","Suspense"]`

		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:      "sync-tags-test-001",
				TitleEn: &titleEn,
				Genres:  &genres,
			},
		}
		eventData := event.Event[*kafka.Message, anime_processor.Payload]{Payload: payload}

		_, err := processor.Process(ctx, eventData)
		require.NoError(t, err)

		// Verify tags were created and associated
		var animeTags []struct {
			AnimeID string
			TagID   int64
		}
		err = database.DB.Table("anime_tags").Where("anime_id = ?", "sync-tags-test-001").Find(&animeTags).Error
		require.NoError(t, err)
		assert.Len(t, animeTags, 3, "Should have 3 tags associated")

		// Verify tag names
		var tagNames []string
		err = database.DB.Table("tags").
			Joins("INNER JOIN anime_tags ON tags.id = anime_tags.tag_id").
			Where("anime_tags.anime_id = ?", "sync-tags-test-001").
			Pluck("name", &tagNames).Error
		require.NoError(t, err)
		assert.Contains(t, tagNames, "Drama")
		assert.Contains(t, tagNames, "Supernatural")
		assert.Contains(t, tagNames, "Suspense")
	})

	t.Run("UpdateAnimeGenres", func(t *testing.T) {
		titleEn := "Sync Tags Update Test"
		oldGenres := `["Action","Comedy"]`
		newGenres := `["Drama","Action"]`

		// Create anime with initial genres
		createPayload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:      "sync-tags-test-002",
				TitleEn: &titleEn,
				Genres:  &oldGenres,
			},
		}
		createEvent := event.Event[*kafka.Message, anime_processor.Payload]{Payload: createPayload}
		_, err := processor.Process(ctx, createEvent)
		require.NoError(t, err)

		// Update with new genres
		updatePayload := anime_processor.Payload{
			Before: &anime_processor.Schema{
				ID:      "sync-tags-test-002",
				TitleEn: &titleEn,
				Genres:  &oldGenres,
			},
			After: &anime_processor.Schema{
				ID:      "sync-tags-test-002",
				TitleEn: &titleEn,
				Genres:  &newGenres,
			},
		}
		updateEvent := event.Event[*kafka.Message, anime_processor.Payload]{Payload: updatePayload}
		_, err = processor.Process(ctx, updateEvent)
		require.NoError(t, err)

		// Verify updated tags
		var tagNames []string
		err = database.DB.Table("tags").
			Joins("INNER JOIN anime_tags ON tags.id = anime_tags.tag_id").
			Where("anime_tags.anime_id = ?", "sync-tags-test-002").
			Pluck("name", &tagNames).Error
		require.NoError(t, err)
		assert.Len(t, tagNames, 2)
		assert.Contains(t, tagNames, "Drama")
		assert.Contains(t, tagNames, "Action")
		assert.NotContains(t, tagNames, "Comedy", "Comedy should be removed")
	})

	t.Run("CreateAnimeWithEmptyGenres", func(t *testing.T) {
		titleEn := "Sync Tags Empty Test"
		emptyGenres := `[]`

		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:      "sync-tags-test-003",
				TitleEn: &titleEn,
				Genres:  &emptyGenres,
			},
		}
		eventData := event.Event[*kafka.Message, anime_processor.Payload]{Payload: payload}

		_, err := processor.Process(ctx, eventData)
		require.NoError(t, err)

		// Verify no tags were associated
		var count int64
		err = database.DB.Table("anime_tags").Where("anime_id = ?", "sync-tags-test-003").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "Should have no tags")
	})

	t.Run("CreateAnimeWithNilGenres", func(t *testing.T) {
		titleEn := "Sync Tags Nil Test"

		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:      "sync-tags-test-004",
				TitleEn: &titleEn,
				Genres:  nil,
			},
		}
		eventData := event.Event[*kafka.Message, anime_processor.Payload]{Payload: payload}

		_, err := processor.Process(ctx, eventData)
		require.NoError(t, err)

		// Verify no tags were associated
		var count int64
		err = database.DB.Table("anime_tags").Where("anime_id = ?", "sync-tags-test-004").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "Should have no tags")
	})

	t.Run("CreateAnimeWithInvalidJSONGenres", func(t *testing.T) {
		titleEn := "Sync Tags Invalid JSON Test"
		invalidGenres := `not valid json`

		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:      "sync-tags-test-005",
				TitleEn: &titleEn,
				Genres:  &invalidGenres,
			},
		}
		eventData := event.Event[*kafka.Message, anime_processor.Payload]{Payload: payload}

		// Should not error, just log warning and clear tags
		_, err := processor.Process(ctx, eventData)
		require.NoError(t, err)

		// Verify no tags were associated (graceful handling)
		var count int64
		err = database.DB.Table("anime_tags").Where("anime_id = ?", "sync-tags-test-005").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "Should have no tags for invalid JSON")
	})

	t.Run("GenresWithWhitespace", func(t *testing.T) {
		titleEn := "Sync Tags Whitespace Test"
		genres := `["  Action  ", "Comedy ", " Drama"]`

		payload := anime_processor.Payload{
			Before: nil,
			After: &anime_processor.Schema{
				ID:      "sync-tags-test-006",
				TitleEn: &titleEn,
				Genres:  &genres,
			},
		}
		eventData := event.Event[*kafka.Message, anime_processor.Payload]{Payload: payload}

		_, err := processor.Process(ctx, eventData)
		require.NoError(t, err)

		// Verify tags were trimmed
		var tagNames []string
		err = database.DB.Table("tags").
			Joins("INNER JOIN anime_tags ON tags.id = anime_tags.tag_id").
			Where("anime_tags.anime_id = ?", "sync-tags-test-006").
			Pluck("name", &tagNames).Error
		require.NoError(t, err)
		assert.Len(t, tagNames, 3)
		assert.Contains(t, tagNames, "Action")
		assert.Contains(t, tagNames, "Comedy")
		assert.Contains(t, tagNames, "Drama")
	})
}

