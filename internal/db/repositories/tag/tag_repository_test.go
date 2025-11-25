package tag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
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

func TestTagRepository_FindOrCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	tagRepo := tag.NewTagRepository(database)

	// Clean up test data
	cleanup := func() {
		database.DB.Exec("DELETE FROM tags WHERE name LIKE ?", "test-findorcreate-%")
	}
	cleanup()
	defer cleanup()

	t.Run("CreatesNewTag", func(t *testing.T) {
		tagName := "test-findorcreate-action"
		createdTag, err := tagRepo.FindOrCreate(tagName)
		require.NoError(t, err)
		assert.NotNil(t, createdTag)
		assert.Equal(t, tagName, createdTag.Name)
		assert.NotZero(t, createdTag.ID)
	})

	t.Run("FindsExistingTag", func(t *testing.T) {
		tagName := "test-findorcreate-existing"

		// Create tag first
		firstTag, err := tagRepo.FindOrCreate(tagName)
		require.NoError(t, err)

		// FindOrCreate again should return the same tag
		secondTag, err := tagRepo.FindOrCreate(tagName)
		require.NoError(t, err)
		assert.Equal(t, firstTag.ID, secondTag.ID)
		assert.Equal(t, firstTag.Name, secondTag.Name)
	})
}

func TestTagRepository_FindByName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	tagRepo := tag.NewTagRepository(database)

	// Clean up test data
	cleanup := func() {
		database.DB.Exec("DELETE FROM tags WHERE name LIKE ?", "test-findbyname-%")
	}
	cleanup()
	defer cleanup()

	t.Run("FindsExistingTag", func(t *testing.T) {
		tagName := "test-findbyname-comedy"

		// Create tag first
		_, err := tagRepo.FindOrCreate(tagName)
		require.NoError(t, err)

		// Find by name
		foundTag, err := tagRepo.FindByName(tagName)
		require.NoError(t, err)
		assert.Equal(t, tagName, foundTag.Name)
	})

	t.Run("ReturnsErrorForNonExistentTag", func(t *testing.T) {
		_, err := tagRepo.FindByName("test-findbyname-nonexistent")
		assert.Error(t, err)
	})
}

func TestTagRepository_FindByNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	tagRepo := tag.NewTagRepository(database)

	// Clean up test data
	cleanup := func() {
		database.DB.Exec("DELETE FROM tags WHERE name LIKE ?", "test-findbynames-%")
	}
	cleanup()
	defer cleanup()

	// Create test tags
	tagNames := []string{"test-findbynames-action", "test-findbynames-comedy", "test-findbynames-drama"}
	for _, name := range tagNames {
		_, err := tagRepo.FindOrCreate(name)
		require.NoError(t, err)
	}

	t.Run("FindsMultipleTags", func(t *testing.T) {
		foundTags, err := tagRepo.FindByNames(tagNames)
		require.NoError(t, err)
		assert.Len(t, foundTags, 3)

		foundNames := make([]string, len(foundTags))
		for i, t := range foundTags {
			foundNames[i] = t.Name
		}
		for _, name := range tagNames {
			assert.Contains(t, foundNames, name)
		}
	})

	t.Run("FindsSubsetOfTags", func(t *testing.T) {
		foundTags, err := tagRepo.FindByNames([]string{"test-findbynames-action", "test-findbynames-drama"})
		require.NoError(t, err)
		assert.Len(t, foundTags, 2)
	})

	t.Run("ReturnsEmptyForNoMatches", func(t *testing.T) {
		foundTags, err := tagRepo.FindByNames([]string{"nonexistent-1", "nonexistent-2"})
		require.NoError(t, err)
		assert.Len(t, foundTags, 0)
	})

	t.Run("HandlesEmptyInput", func(t *testing.T) {
		foundTags, err := tagRepo.FindByNames([]string{})
		require.NoError(t, err)
		assert.Len(t, foundTags, 0)
	})
}

func TestTagRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupTestDB(t)
	tagRepo := tag.NewTagRepository(database)

	// Clean up test data
	cleanup := func() {
		database.DB.Exec("DELETE FROM tags WHERE name LIKE ?", "test-create-%")
	}
	cleanup()
	defer cleanup()

	t.Run("CreatesNewTag", func(t *testing.T) {
		newTag := &tag.Tag{
			Name: "test-create-supernatural",
		}
		err := tagRepo.Create(newTag)
		require.NoError(t, err)
		assert.NotZero(t, newTag.ID)

		// Verify it can be found
		foundTag, err := tagRepo.FindByName("test-create-supernatural")
		require.NoError(t, err)
		assert.Equal(t, newTag.ID, foundTag.ID)
	})

	t.Run("FailsOnDuplicateName", func(t *testing.T) {
		tagName := "test-create-duplicate"

		// Create first tag
		firstTag := &tag.Tag{Name: tagName}
		err := tagRepo.Create(firstTag)
		require.NoError(t, err)

		// Attempt to create duplicate
		secondTag := &tag.Tag{Name: tagName}
		err = tagRepo.Create(secondTag)
		assert.Error(t, err)
	})
}
