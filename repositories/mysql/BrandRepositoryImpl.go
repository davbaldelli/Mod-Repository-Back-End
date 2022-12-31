package mysql

import (
	"github.com/davide/ModRepository/models/entities"
	"github.com/davide/ModRepository/repositories/models"
	"gorm.io/gorm"
)

type BrandRepositoryImpl struct {
	Db *gorm.DB
}

func (b BrandRepositoryImpl) SelectAllBrands() ([]entities.CarBrand, error) {
	return b.selectBrandsWithQuery(func(brands *[]models.Manufacturer) *gorm.DB {
		return b.Db.Order("name ASC").Find(&brands)
	})
}

func (b BrandRepositoryImpl) selectBrandsWithQuery(query selectFromBrandsQuery) ([]entities.CarBrand, error) {
	var dbBrands []models.Manufacturer
	var brands []entities.CarBrand
	if result := query(&dbBrands); result.Error != nil {
		return nil, result.Error
	}
	for _, dbBrand := range dbBrands {
		nation := models.Nation{Id: dbBrand.IdNation}
		if res2 := b.Db.Find(&nation); res2.Error != nil {
			return nil, res2.Error
		} else {
			brands = append(brands, dbBrand.ToEntity(nation))
		}
	}
	return brands, nil
}
