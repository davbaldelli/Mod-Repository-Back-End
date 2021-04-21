package postgresrepo

import (
	"fmt"
	"github.com/davide/ModRepository/models/db"
	"github.com/davide/ModRepository/models/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CarRepositoryImpl struct {
	Db *gorm.DB
}

type selectFromCarsQuery func(*[]db.Car) *gorm.DB
type selectFromBrandsQuery func(*[]db.CarBrand) *gorm.DB

func dbCarToEntity(dbCar db.Car, nation string, author db.Author)entities.Car{
	return entities.Car{
		Mod: entities.Mod{
			DownloadLink: dbCar.DownloadLink,
			Premium: dbCar.Premium,
			Image: dbCar.Image,
			Author: entities.Author{
				Name: author.Name,
				Link: author.Link,
			},
		},
		Brand: entities.CarBrand{
			Name:   dbCar.Brand,
			Nation: entities.Nation{Name: nation},
		},
		ModelName:  dbCar.ModelName,
		Categories: allCategoriesToEntity(dbCar.Categories),
		Drivetrain: entities.Drivetrain(dbCar.Drivetrain),
		Transmission: entities.Transmission(dbCar.Transmission),
		Year: dbCar.Year,
		Torque: dbCar.Torque,
		TopSpeed: dbCar.TopSpeed,
		Weight: dbCar.Weight,
		BHP: dbCar.BHP,

	}
}

func selectCarsWithQuery(carsQuery selectFromCarsQuery, brandsQuery selectFromBrandsQuery, authorsQuery selectFromAuthorsQuery) ([]entities.Car, error){
	var cars []entities.Car
	var dbCars []db.Car
	var dbAuthors []db.Author

	if result := carsQuery(&dbCars); result.Error != nil{
		return nil,result.Error
	}
	var dbBrands []db.CarBrand
	if result := brandsQuery(&dbBrands); result.Error != nil {
		return nil,result.Error
	}
	
	var carNames []string
	for _, car := range dbCars {
		carNames = append(carNames, car.ModelName)
	}
	authorsMap := make(map[string]db.Author)
	if result := authorsQuery(&dbAuthors, carNames); result.Error != nil {
		return nil, result.Error
	}
	for _, author := range dbAuthors {
		authorsMap[author.Name] = author
	}
	brandsNation := make(map[string]string)
	for _, brand := range dbBrands {
		brandsNation[brand.Name] = brand.Nation
	}
	for _, dbCar := range dbCars {
		cars = append(cars, dbCarToEntity(dbCar, brandsNation[dbCar.Brand], authorsMap[dbCar.Author]))
	}
	return cars,nil
}

func (c CarRepositoryImpl) SelectAllCarCategories() ([]entities.CarCategory, error) {
	var categories []db.CarCategory
	if result := c.Db.Find(&categories) ; result.Error != nil{
		return  nil, result.Error
	}
	return allCategoriesToEntity(categories), nil
}

func (c CarRepositoryImpl) InsertCar(car entities.Car) error {
	dbCar := db.CarFromEntity(car)
	dbNation := db.NationFromEntity(car.Brand.Nation)
	dbBrand := db.BrandFromEntity(car.Brand)

	if res := c.Db.Clauses(clause.OnConflict{DoNothing: true}).Create(&dbNation); res.Error != nil {
		return res.Error
	}

	if res := c.Db.Clauses(clause.OnConflict{DoNothing: true}).Create(&dbBrand); res.Error != nil {
		return res.Error
	}

	if res := c.Db.Create(&dbCar); res.Error != nil {
		return res.Error
	}
	return nil
}

func (c CarRepositoryImpl) SelectAllCars() ([]entities.Car,error) {
	return selectCarsWithQuery(func(cars *[]db.Car) *gorm.DB {
		return c.Db.Order("concat(brand,' ',model_name) ASC").Preload("Categories").Find(&cars)
	}, func(brands *[]db.CarBrand) *gorm.DB {
		return c.Db.Find(&brands)
	}, func(authors *[]db.Author, carsNames []string) *gorm.DB {
		return c.Db.Joins("join cars on author = cars.author").Find(authors, "cars.model_name IN ?",carsNames)
	})
}

func allCategoriesToEntity(dbCategories []db.CarCategory) []entities.CarCategory{
	var cats []entities.CarCategory
	for _,dbCat := range  dbCategories {
		cats = append(cats, entities.CarCategory{Name: dbCat.Name})
	}
	return cats
}


func (c CarRepositoryImpl) SelectCarsByNation(nation string) ([]entities.Car,error) {
	return selectCarsWithQuery(func(cars *[]db.Car) *gorm.DB {
		return c.Db.Order("concat(brand,' ',model_name) ASC").Preload("Categories").Joins("join car_brands on cars.brand = car_brands.name").Where("car_brands.nation = ?",nation).Find(&cars)
	}, func(brands *[]db.CarBrand) *gorm.DB {
		return c.Db.Find(&brands,"nation = ?",nation)
	}, func(authors *[]db.Author, carsNames []string) *gorm.DB {
		return c.Db.Joins("join cars on author = cars.author").Find(authors, "cars.model_name IN ?",carsNames)
	})

}

func (c CarRepositoryImpl) SelectCarsByModelName(model string) ([]entities.Car,error) {
	return selectCarsWithQuery(func(cars *[]db.Car) *gorm.DB {
		return c.Db.Order("concat(brand,' ',model_name) ASC").Preload("Categories").Find(&cars,"LOWER(concat(brand,' ',model_name)) LIKE LOWER(?)", "%"+model+"%").Find(&cars)
	}, func(brands *[]db.CarBrand) *gorm.DB {
		return c.Db.Joins("right join cars on cars.brand = car_brands.name").Find(&brands,"cars.model_name = ?",model)
	}, func(authors *[]db.Author, carsNames []string) *gorm.DB {
		return c.Db.Joins("join cars on author = cars.author").Find(authors, "cars.model_name IN ?",carsNames)
	})
}

func (c CarRepositoryImpl) SelectCarsByBrand(brandName string) ([]entities.Car,error) {
	return selectCarsWithQuery(func(cars *[]db.Car) *gorm.DB {
		return c.Db.Order("concat(brand,' ',model_name) ASC").Preload("Categories").Find(&cars,"brand = ?",brandName)
	}, func(brands *[]db.CarBrand) *gorm.DB {
		return  c.Db.Find(&brands,"name = ?", brandName)
	}, func(authors *[]db.Author, carsNames []string) *gorm.DB {
		return c.Db.Joins("join cars on author = cars.author").Find(authors, "cars.model_name IN ?",carsNames)
	})
}

func (c CarRepositoryImpl) SelectCarsByType(category string) ([]entities.Car,error) {
	var cars []entities.Car
	var dbCars []db.Car
	var dbAuthors []db.Author

	if result := c.Db.Order("concat(brand,' ',model_name) ASC").Preload("Categories").Joins("join cars_categories_ass on cars_categories_ass.car_model_name = model_name").Where("car_category_name = ?", category).Find(&dbCars); result.Error != nil{
		return nil,result.Error
	}
	var brandsNames []string
	for _, car := range dbCars {
		brandsNames = append(brandsNames, car.Brand)
	}

	var carNames []string
	for _, car := range dbCars {
		carNames = append(carNames, car.ModelName)
	}
	authorsMap := make(map[string]db.Author)
	if result := c.Db.Joins("join cars on author = cars.author").Find(&dbAuthors, "cars.model_name IN ?",carNames); result.Error != nil {
		return nil, result.Error
	}
	for _, author := range dbAuthors {
		authorsMap[author.Name] = author
	}

	var dbBrands []db.CarBrand
	if result := c.Db.Find(&dbBrands, "name IN ?", brandsNames); result.Error != nil {
		return nil,result.Error
	}
	brandsNation := make(map[string]string)
	for _, brand := range dbBrands {
		brandsNation[brand.Name] = brand.Nation
	}
	for _, dbCar := range dbCars {
		cars = append(cars, dbCarToEntity(dbCar, brandsNation[dbCar.Brand],authorsMap[dbCar.Author]))
	}
	fmt.Println(cars)
	return cars, nil
}
