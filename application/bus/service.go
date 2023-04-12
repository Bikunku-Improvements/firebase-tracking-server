package bus

import (
	"fmt"
	"strconv"
	"tracking-server/shared"
	"tracking-server/shared/dto"

	"context"
	"log"

	firebase "firebase.google.com/go"
)

type (
	Service interface {
		Create(data *dto.Bus) error
		FindByUsername(username string, bus *dto.Bus) error
		Delete(id string) error
		Save(data *dto.Bus) error
		FindById(id string, bus *dto.Bus) error
		InsertBusLocation(location *dto.BusLocation) error
		FindAllBus(bus *[]dto.Bus) error
		FindBusLatestLocation(id uint, location *dto.BusLocation) error
		InsertBusLocationFirebase(id string, location *dto.BusLocationFirebase) error
	}
	service struct {
		shared shared.Holder
	}
)

func (s *service) Create(data *dto.Bus) error {
	err := s.shared.DB.Create(data).Error
	return err
}

func (s *service) FindByUsername(username string, bus *dto.Bus) error {
	err := s.shared.DB.Where("username = ?", username).First(bus).Error
	return err
}

func (s *service) Delete(id string) error {
	err := s.shared.DB.Delete(&dto.Bus{}, id).Error
	return err
}

func (s *service) Save(data *dto.Bus) error {
	err := s.shared.DB.Save(data).Error
	return err
}

func (s *service) FindById(id string, bus *dto.Bus) error {
	err := s.shared.DB.Where("id = ?", id).First(bus).Error
	return err
}

func (s *service) InsertBusLocation(location *dto.BusLocation) error {
	err := s.shared.DB.Create(location).Error
	return err
}

func (s *service) FindAllBus(bus *[]dto.Bus) error {
	err := s.shared.DB.Find(bus).Error
	return err
}

func (s *service) FindBusLatestLocation(id uint, location *dto.BusLocation) error {
	err := s.shared.DB.Where("bus_id = ?", id).Order("timestamp DESC").First(location).Error
	return err
}

func (s *service) InsertBusLocationFirebase(id string, location *dto.BusLocationFirebase) error {
	// Connect Google Cloud
	// Use the application default credentials
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "ta-tracking-f43e5"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	defer client.Close()

	// Execution
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error during conversion")
		return err
	}

	// For adding multiple docs in locations
	ref := client.Collection("bus_locations").NewDoc()

	// For updating 1 doc in locations
	// ref := client.Collection("buses").Doc(id).Collection("locations").Doc(id)

	_, err = ref.Set(ctx, map[string]interface{}{
		"bus_id": idInt,
		"number": location.Number,
		"plate": location.Plate,
		"status": location.Status,
		"route": location.Route,
		"isActive": location.IsActive,
		"longitude": location.Long,
		"latitude": location.Lat,
		"timestamp": location.Timestamp,
		"speed": location.Speed,
		"heading": location.Speed,
	})
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred: %s", err)
	}

	// err := s.shared.DB.Delete(&dto.Bus{}, id).Error
	return err
}

func NewBusService(shared shared.Holder) Service {
	return &service{
		shared: shared,
	}
}
