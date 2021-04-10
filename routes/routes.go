package routes

import (
	"github.com/davide/ModRepository/routes/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
)

type Route interface {
	Listen(string)
}

type Web struct {
	CarHandler    handlers.CarsHandler
	TracksHandler handlers.TracksHandler
	NationHandler handlers.NationsHandler
	BrandsHandler handlers.BrandsHandler
}

func (w Web) Listen(port string) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/car/new", w.CarHandler.POSTNewCar).Methods("POST")
	router.HandleFunc("/car/all", w.CarHandler.GETAllCars).Methods("GET")
	router.HandleFunc("/car/nation/{nation}", w.CarHandler.GETCarsByNation).Methods("GET")
	router.HandleFunc("/car/model/{model}", w.CarHandler.GETCarByModel).Methods("GET")
	router.HandleFunc("/car/brand/{brand}", w.CarHandler.GETCarsByBrand).Methods("GET")
	router.HandleFunc("/car/category/{category}", w.CarHandler.GETCarsByType).Methods("GET")
	router.HandleFunc("/car/type/all", w.CarHandler.GETAllCarCategories).Methods("GET")

	router.HandleFunc("/track/new", w.TracksHandler.POSTNewTrack).Methods("POST")
	router.HandleFunc("/track/all", w.TracksHandler.GETAllTracks).Methods("GET")
	router.HandleFunc("/track/nation/{nation}", w.TracksHandler.GETTracksByNation).Methods("GET")
	router.HandleFunc("/track/layout/type/{layoutType}", w.TracksHandler.GETTracksByLayoutType).Methods("GET")
	router.HandleFunc("/track/name/{name}", w.TracksHandler.GETTrackByName).Methods("GET")
	router.HandleFunc("/layout/type/all", w.TracksHandler.GETAllLayoutTypes).Methods("GET")

	router.HandleFunc("/nation/brand/all", w.NationHandler.GETAllBrandsNations).Methods("GET")
	router.HandleFunc("/nation/track/all", w.NationHandler.GETAllTracksNations).Methods("GET")

	router.HandleFunc("/brand/all", w.BrandsHandler.GETAllBrands).Methods("GET")
	router.HandleFunc("/brand/name/{name}", w.BrandsHandler.GETBrandByName).Methods("GET")
	router.HandleFunc("/brand/nation/{nation}", w.BrandsHandler.GETBrandByNation).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
