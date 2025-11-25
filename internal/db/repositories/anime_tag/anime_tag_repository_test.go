package anime_tag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime_tag"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/tag"
)

func setupTestDB(t *testing.T) *db.DB {
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

	return database
}

func TestAnimeTagRepository_SetTagsForAnime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	animeRepo := anime.NewAnimeRepository(database)
	tagRepo := tag.NewTagRepository(database)
	animeTagRepo := anime_tag.NewAnimeTagRepository(database)

	// Clean up test data
	cleanup := func() {
		database.DB.Exec("DELETE FROM anime_tags WHERE anime_id LIKE ?", "test-anime-tag-%")
		database.DB.Exec("DELETE FROM tags WHERE name LIKE ?", "test-tag-%")
		database.DB.Where("id LIKE ?", "test-anime-tag-%").Delete(&anime.Anime{})
	}
	cleanup()
	defer cleanup()

	// Create test anime
	titleEn := "Test Anime for Tags"
	testAnime := &anime.Anime{
		ID:      "test-anime-tag-001",
		TitleEn: &titleEn,
	}
	err := animeRepo.Upsert(testAnime, nil)
	require.NoError(t, err)

	// Create test tags
	tag1, err := tagRepo.FindOrCreate("test-tag-action")
	require.NoError(t, err)
	tag2, err := tagRepo.FindOrCreate("test-tag-comedy")
	require.NoError(t, err)
	tag3, err := tagRepo.FindOrCreate("test-tag-drama")
	require.NoError(t, err)

	t.Run("SetTagsForAnime_AddsNewTags", func(t *testing.T) {
		tagIDs := []int64{tag1.ID, tag2.ID}
		err := animeTagRepo.SetTagsForAnime("test-anime-tag-001", tagIDs)
		require.NoError(t, err)

		// Verify tags were added
		savedTagIDs, err := animeTagRepo.GetTagIDsForAnime("test-anime-tag-001")
		require.NoError(t, err)
		assert.Len(t, savedTagIDs, 2)
		assert.Contains(t, savedTagIDs, tag1.ID)
		assert.Contains(t, savedTagIDs, tag2.ID)
	})

	t.Run("SetTagsForAnime_ReplacesExistingTags", func(t *testing.T) {
		// Set new tags that replace the old ones
		tagIDs := []int64{tag2.ID, tag3.ID}
		err := animeTagRepo.SetTagsForAnime("test-anime-tag-001", tagIDs)
		require.NoError(t, err)

		// Verify old tags were replaced
		savedTagIDs, err := animeTagRepo.GetTagIDsForAnime("test-anime-tag-001")
		require.NoError(t, err)
		assert.Len(t, savedTagIDs, 2)
		assert.NotContains(t, savedTagIDs, tag1.ID)
		assert.Contains(t, savedTagIDs, tag2.ID)
		assert.Contains(t, savedTagIDs, tag3.ID)
	})

	t.Run("SetTagsForAnime_ClearsAllTags", func(t *testing.T) {
		// Set empty tags
		err := animeTagRepo.SetTagsForAnime("test-anime-tag-001", []int64{})
		require.NoError(t, err)

		// Verify all tags were removed
		savedTagIDs, err := animeTagRepo.GetTagIDsForAnime("test-anime-tag-001")
		require.NoError(t, err)
		assert.Len(t, savedTagIDs, 0)
	})
}

func TestAnimeTagRepository_AddAndRemoveTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	animeRepo := anime.NewAnimeRepository(database)
	tagRepo := tag.NewTagRepository(database)
	animeTagRepo := anime_tag.NewAnimeTagRepository(database)

	// Clean up test data
	cleanup := func() {
		database.DB.Exec("DELETE FROM anime_tags WHERE anime_id LIKE ?", "test-anime-tag-%")
		database.DB.Exec("DELETE FROM tags WHERE name LIKE ?", "test-tag-%")
		database.DB.Where("id LIKE ?", "test-anime-tag-%").Delete(&anime.Anime{})
	}
	cleanup()
	defer cleanup()

	// Create test anime
	titleEn := "Test Anime for Add/Remove"
	testAnime := &anime.Anime{
		ID:      "test-anime-tag-002",
		TitleEn: &titleEn,
	}
	err := animeRepo.Upsert(testAnime, nil)
	require.NoError(t, err)

	// Create test tag
	testTag, err := tagRepo.FindOrCreate("test-tag-single")
	require.NoError(t, err)

	t.Run("AddTagToAnime", func(t *testing.T) {
		err := animeTagRepo.AddTagToAnime("test-anime-tag-002", testTag.ID)
		require.NoError(t, err)

		savedTagIDs, err := animeTagRepo.GetTagIDsForAnime("test-anime-tag-002")
		require.NoError(t, err)
		assert.Len(t, savedTagIDs, 1)
		assert.Contains(t, savedTagIDs, testTag.ID)
	})

	t.Run("RemoveTagFromAnime", func(t *testing.T) {
		err := animeTagRepo.RemoveTagFromAnime("test-anime-tag-002", testTag.ID)
		require.NoError(t, err)

		savedTagIDs, err := animeTagRepo.GetTagIDsForAnime("test-anime-tag-002")
		require.NoError(t, err)
		assert.Len(t, savedTagIDs, 0)
	})
}

func TestAnimeTagRepository_DeleteAllTagsForAnime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	animeRepo := anime.NewAnimeRepository(database)
	tagRepo := tag.NewTagRepository(database)
	animeTagRepo := anime_tag.NewAnimeTagRepository(database)

	// Clean up test data
	cleanup := func() {
		database.DB.Exec("DELETE FROM anime_tags WHERE anime_id LIKE ?", "test-anime-tag-%")
		database.DB.Exec("DELETE FROM tags WHERE name LIKE ?", "test-tag-%")
		database.DB.Where("id LIKE ?", "test-anime-tag-%").Delete(&anime.Anime{})
	}
	cleanup()
	defer cleanup()

	// Create test anime
	titleEn := "Test Anime for DeleteAll"
	testAnime := &anime.Anime{
		ID:      "test-anime-tag-003",
		TitleEn: &titleEn,
	}
	err := animeRepo.Upsert(testAnime, nil)
	require.NoError(t, err)

	// Create and add multiple tags
	tag1, err := tagRepo.FindOrCreate("test-tag-delete-1")
	require.NoError(t, err)
	tag2, err := tagRepo.FindOrCreate("test-tag-delete-2")
	require.NoError(t, err)

	err = animeTagRepo.SetTagsForAnime("test-anime-tag-003", []int64{tag1.ID, tag2.ID})
	require.NoError(t, err)

	// Verify tags exist
	savedTagIDs, err := animeTagRepo.GetTagIDsForAnime("test-anime-tag-003")
	require.NoError(t, err)
	assert.Len(t, savedTagIDs, 2)

	// Delete all tags
	err = animeTagRepo.DeleteAllTagsForAnime("test-anime-tag-003")
	require.NoError(t, err)

	// Verify all tags were removed
	savedTagIDs, err = animeTagRepo.GetTagIDsForAnime("test-anime-tag-003")
	require.NoError(t, err)
	assert.Len(t, savedTagIDs, 0)
}

func TestAnimeTagRepository_GetTagIDsForAnime_EmptyResult(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	animeTagRepo := anime_tag.NewAnimeTagRepository(database)

	// Get tags for non-existent anime
	tagIDs, err := animeTagRepo.GetTagIDsForAnime("non-existent-anime-id")
	require.NoError(t, err)
	assert.Len(t, tagIDs, 0)
}
