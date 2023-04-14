package dto

import (
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

var (
	ExperimentalBusLocation sync.Map
)

const (
	// Crowded Status
	EMPTY    BusStatus = "EMPTY"
	MODERATE BusStatus = "MODERATE"
	FULL     BusStatus = "FULL"

	// Route
	RED  Route = "RED"
	BLUE Route = "BLUE"

	CLIENT WSType = "client"
	DRIVER WSType = "driver"

	DEFAULTBUSSPEED = 1.0
)

type (
	BusStatus string

	Route string

	WSType string

	Bus struct {
		ID       uint          `gorm:"primaryKey;autoIncrement"`
		Number   int           `gorm:"column:number;unique"`
		Plate    string        `gorm:"column:plate;unique"`
		Status   BusStatus     `gorm:"column:status;default:EMPTY"`
		Route    Route         `gorm:"column:route"`
		IsActive bool          `gorm:"column:is_active;default:false"`
		Username string        `gorm:"column:username;unique"`
		Password string        `gorm:"password"`
		Location []BusLocation `gorm:"foreignKey:bus_id"`
	}

	BusLocation struct {
		ID        uint      `gorm:"primaryKey;autoIncrement"`
		BusID     uint      `gorm:"column:bus_id"`
		Long      float64   `gorm:"column:longitude"`
		Lat       float64   `gorm:"column:latitude"`
		Timestamp time.Time `gorm:"column:timestamp"`
		Speed     float64   `gorm:"column:speed"`
		Heading   float64   `gorm:"column:heading"`
	}

	// CreateBusDto CreateBusDto
	CreateBusDto struct {
		Number   int    `json:"number" validate:"required"`
		Plate    string `json:"plate" validate:"required"`
		Route    Route  `json:"route" validate:"required,oneof=RED BLUE"`
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	// CreateBusResponse CreateBusResponse
	CreateBusResponse struct {
		ID       uint      `json:"id"`
		Number   int       `json:"number"`
		Plate    string    `json:"plate"`
		Status   BusStatus `json:"status"`
		Route    Route     `json:"route"`
		IsActive bool      `json:"isActive"`
		Username string    `json:"username"`
	}

	// DriverLoginDto DriverLoginDto
	DriverLoginDto struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	// DriverLoginResponse DriverLoginResponse
	DriverLoginResponse struct {
		ID       uint      `json:"id"`
		Number   int       `json:"number"`
		Plate    string    `json:"plate"`
		Status   BusStatus `json:"status"`
		Route    Route     `json:"route"`
		IsActive bool      `json:"isActive"`
		Token    string    `json:"token"`
	}

	// EditBusDto EditBusDto
	EditBusDto struct {
		Number   int       `json:"number" validate:"omitempty"`
		Plate    string    `json:"plate" validate:"omitempty"`
		Status   BusStatus `json:"status" validate:"omitempty,oneof=EMPTY MODERATE FULL"`
		Route    Route     `json:"route" validate:"omitempty,oneof=RED BLUE"`
		IsActive bool      `json:"isActive" validate:"omitempty"`
	}

	// EditBusResponse EditBusResponse
	EditBusResponse struct {
		ID       uint      `json:"id"`
		Number   int       `json:"number"`
		Plate    string    `json:"plate"`
		Status   BusStatus `json:"status"`
		Route    Route     `json:"route"`
		IsActive bool      `json:"isActive"`
	}

	BusLocationQuery struct {
		Type           string
		Token          string
		Experimental   string
		ExperminetalID string
	}

	BusLocationMessage struct {
		Long    float64 `json:"long"`
		Lat     float64 `json:"lat"`
		Speed   float64 `json:"speed"`
		Heading float64 `json:"heading"`
	}

	TrackLocationResponse struct {
		ID       uint      `json:"id"`
		Number   int       `json:"number"`
		Plate    string    `json:"plate"`
		Status   BusStatus `json:"status"`
		Route    Route     `json:"route"`
		IsActive bool      `json:"isActive"`
		Long     float64   `json:"long"`
		Lat      float64   `json:"lat"`
		Speed    float64   `json:"speed"`
		Heading  float64   `json:"heading"`
	}
	BusInfo struct {
		ID       uint      `json:"id"`
		Number   int       `json:"number"`
		Plate    string    `json:"plate"`
		Status   BusStatus `json:"status"`
		Route    Route     `json:"route"`
		Estimate int       `json:"estimate"`
	}

	// BusInfoResponse BusInfoResponse
	BusInfoResponse struct {
		Bus []BusInfo `json:"bus"`
	}

	Connection struct {
		Socket *websocket.Conn
		Mu     sync.Mutex
	}
)

func (c *Connection) Send(data []TrackLocationResponse) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.Socket.WriteJSON(data)
}

func (b *Bus) ToCreateBusResponse() CreateBusResponse {
	return CreateBusResponse{
		ID:       b.ID,
		Number:   b.Number,
		Plate:    b.Plate,
		Status:   b.Status,
		Route:    b.Route,
		IsActive: b.IsActive,
		Username: b.Username,
	}
}

func (b *Bus) ToDriverLoginResponse(token string) DriverLoginResponse {
	return DriverLoginResponse{
		ID:       b.ID,
		Number:   b.Number,
		Plate:    b.Plate,
		Status:   b.Status,
		Route:    b.Route,
		IsActive: b.IsActive,
		Token:    token,
	}
}

func (b *Bus) FillBusEdit(data EditBusDto) {
	if data.Number != 0 {
		b.Number = data.Number
	}

	if data.Plate != "" {
		b.Plate = data.Plate
	}

	if data.Status != "" {
		b.Status = data.Status
	}

	if data.Route != "" {
		b.Route = data.Route
	}

	if b.IsActive != data.IsActive {
		b.IsActive = data.IsActive
	}
}

func (b *Bus) ToEditBusResponnse() EditBusResponse {
	return EditBusResponse{
		ID:       b.ID,
		Number:   b.Number,
		Plate:    b.Plate,
		Status:   b.Status,
		Route:    b.Route,
		IsActive: b.IsActive,
	}
}

func (t *TrackLocationResponse) GetBusSpeed() float64 {
	if t.Speed <= 0.0 {
		return DEFAULTBUSSPEED
	}
	return t.Speed
}
