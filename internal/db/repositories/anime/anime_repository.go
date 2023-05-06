package anime

import "github.com/weeb-vip/anime-sync/internal/db"

type RECORD_TYPE string

type AnimeRepository interface {
	FindAll() ([]*Anime, error)
	FindById(id int) (*Anime, error)
	FindByName(name string) ([]*Anime, error)
	FindByType(recordType RECORD_TYPE) ([]*Anime, error)
	FindByStatus(status string) ([]*Anime, error)
	FindBySource(source string) ([]*Anime, error)
	FindByGenre(genre string) ([]*Anime, error)
	FindByStudio(studio string) ([]*Anime, error)
	FindByLicensors(licensors string) ([]*Anime, error)
	FindByRating(rating string) ([]*Anime, error)
	FindByYear(year int) ([]*Anime, error)
	FindBySeason(season string) ([]*Anime, error)
	FindByYearAndSeason(year int, season string) ([]*Anime, error)
	FindByYearAndSeasonAndType(year int, season string, recordType RECORD_TYPE) ([]*Anime, error)
	FindByYearAndSeasonAndTypeAndStatus(year int, season string, recordType RECORD_TYPE, status string) ([]*Anime, error)
	FindByYearAndSeasonAndTypeAndStatusAndSource(year int, season string, recordType RECORD_TYPE, status string, source string) ([]*Anime, error)
	FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenre(year int, season string, recordType RECORD_TYPE, status string, source string, genre string) ([]*Anime, error)
	FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudio(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string) ([]*Anime, error)
	FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudioAndLicensors(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string, licensors string) ([]*Anime, error)
	FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudioAndLicensorsAndRating(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string, licensors string, rating string) ([]*Anime, error)
	FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudioAndLicensorsAndRatingAndName(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string, licensors string, rating string, name string) ([]*Anime, error)
}

type AnimeRepositoryImpl struct {
	db *db.DB
}

func NewAnimeRepository(db *db.DB) AnimeRepository {
	return &AnimeRepositoryImpl{db: db}
}

func (a *AnimeRepositoryImpl) FindAll() ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindById(id int) (*Anime, error) {
	var anime Anime
	err := a.db.DB.Where("id = ?", id).First(&anime).Error
	if err != nil {
		return nil, err
	}
	return &anime, nil
}

func (a *AnimeRepositoryImpl) FindByName(name string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("name = ?", name).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByType(recordType RECORD_TYPE) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("type = ?", recordType).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByStatus(status string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("status = ?", status).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindBySource(source string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("source = ?", source).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByGenre(genre string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("genre = ?", genre).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByStudio(studio string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("studio = ?", studio).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByLicensors(licensors string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("licensors = ?", licensors).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByRating(rating string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("rating = ?", rating).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYear(year int) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ?", year).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindBySeason(season string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("season = ?", season).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeason(year int, season string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ?", year, season).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndType(year int, season string, recordType RECORD_TYPE) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ?", year, season, recordType).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndTypeAndStatus(year int, season string, recordType RECORD_TYPE, status string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ? AND status = ?", year, season, recordType, status).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndTypeAndStatusAndSource(year int, season string, recordType RECORD_TYPE, status string, source string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ? AND status = ? AND source = ?", year, season, recordType, status, source).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenre(year int, season string, recordType RECORD_TYPE, status string, source string, genre string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ? AND status = ? AND source = ? AND genre = ?", year, season, recordType, status, source, genre).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudio(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ? AND status = ? AND source = ? AND genre = ? AND studio = ?", year, season, recordType, status, source, genre, studio).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudioAndLicensors(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string, licensors string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ? AND status = ? AND source = ? AND genre = ? AND studio = ? AND licensors = ?", year, season, recordType, status, source, genre, studio, licensors).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudioAndLicensorsAndRating(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string, licensors string, rating string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ? AND status = ? AND source = ? AND genre = ? AND studio = ? AND licensors = ? AND rating = ?", year, season, recordType, status, source, genre, studio, licensors, rating).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}

func (a *AnimeRepositoryImpl) FindByYearAndSeasonAndTypeAndStatusAndSourceAndGenreAndStudioAndLicensorsAndRatingAndName(year int, season string, recordType RECORD_TYPE, status string, source string, genre string, studio string, licensors string, rating string, name string) ([]*Anime, error) {
	var animes []*Anime
	err := a.db.DB.Where("year = ? AND season = ? AND type = ? AND status = ? AND source = ? AND genre = ? AND studio = ? AND licensors = ? AND rating = ? AND name = ?", year, season, recordType, status, source, genre, studio, licensors, rating, name).Find(&animes).Error
	if err != nil {
		return nil, err
	}
	return animes, nil
}
