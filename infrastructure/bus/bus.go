package bus

import (
	"time"
	"tracking-server/interfaces"
	"tracking-server/shared"
	"tracking-server/shared/dto"

	"tracking-server/shared/common"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"google.golang.org/api/option"

	"context"

	firebase "firebase.google.com/go"
)

type Controller struct {
	Interfaces interfaces.Holder
	Shared     shared.Holder
}

func (c *Controller) Routes(app *fiber.App) {
	bus := app.Group("/bus")
	bus.Post("/", c.create)
	bus.Post("/login", c.login)
	bus.Post("/loginAlt", c.loginAlt)
	bus.Delete("/:id", c.delete)
	bus.Put("/:id", c.edit)
	bus.Post("/info/:id", c.busInfo)

	bus.Use("/stream", func(ctx *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(ctx) {
			return ctx.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	bus.Get("/stream", websocket.New(c.trackBusLocation))
	bus.Get("/streamfirebase", websocket.New(c.trackBusLocationFirebase))
}

// All godoc
// @Tags Bus
// @Summary Create new bus entry
// @Description Put all mandatory parameter
// @Param CreateBusDto body dto.CreateBusDto true "CreateBus"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.CreateBusResponse
// @Failure 200 {object} dto.CreateBusResponse
// @Router /bus/ [post]
func (c *Controller) create(ctx *fiber.Ctx) error {
	var (
		body     dto.CreateBusDto
		response dto.CreateBusResponse
	)

	err := common.DoCommonRequest(ctx, &body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	c.Shared.Logger.Infof("create bus, data: %s", body)

	response, err = c.Interfaces.BusViewService.CreateBusEntry(body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	return common.DoCommonSuccessResponse(ctx, response)
}

// All godoc
// @Tags Bus
// @Summary Driver login
// @Description Put all mandatory parameter
// @Param DriverLoginDto body dto.DriverLoginDto true "DriverLoginDto"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.DriverLoginResponse
// @Failure 200 {object} dto.DriverLoginResponse
// @Router /bus/login [post]
func (c *Controller) login(ctx *fiber.Ctx) error {
	var (
		body     dto.DriverLoginDto
		response dto.DriverLoginResponse
	)

	err := common.DoCommonRequest(ctx, &body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	c.Shared.Logger.Infof("login driver, data: %s", body)

	response, err = c.Interfaces.BusViewService.LoginDriver(body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	return common.DoCommonSuccessResponse(ctx, response)
}

// All godoc
// @Tags Bus
// @Summary Alternative Driver login
// @Description Put all mandatory parameter
// @Param DriverLoginDto body dto.DriverLoginDto true "DriverLoginDto"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.DriverLoginResponse
// @Failure 200 {object} dto.DriverLoginResponse
// @Router /bus/loginAlt [post]
func (c *Controller) loginAlt(ctx *fiber.Ctx) error {
	var (
		body     dto.DriverLoginDto
		response dto.DriverLoginResponse
	)

	err := common.DoCommonRequest(ctx, &body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	c.Shared.Logger.Infof("login driver, data: %s", body)

	response, err = c.Interfaces.BusViewService.LoginDriverAlt(body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	return common.DoCommonSuccessResponse(ctx, response)
}

// All godoc
// @Tags Bus
// @Summary Delete bus
// @Description Put all mandatory parameter
// @Param id path string true "Bus ID"
// @Accept  json
// @Produce  json
// @Router /bus/{id} [delete]
func (c *Controller) delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	c.Shared.Logger.Infof("delete bus, data: %s", id)

	err := c.Interfaces.BusViewService.DeleteBus(id)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	return common.DoCommonSuccessResponse(ctx, nil)
}

// All godoc
// @Tags Bus
// @Summary Edit Bus
// @Description Put all mandatory parameter
// @Param id path string true "Bus ID"
// @Param auth header string true "token"
// @Param EditBusDto body dto.EditBusDto true "EditBusDto"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.EditBusResponse
// @Failure 200 {object} dto.EditBusResponse
// @Router /bus/{id} [put]
func (c *Controller) edit(ctx *fiber.Ctx) error {
	var (
		body     dto.EditBusDto
		response dto.EditBusResponse
	)

	err := common.DoCommonRequest(ctx, &body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	id := ctx.Params("id")

	auth := ctx.Get("auth")

	c.Shared.Logger.Infof("edit driver, data: %s, id: %s, token: %s", body, id, auth)

	response, err = c.Interfaces.BusViewService.EditBus(body, id, auth)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	return common.DoCommonSuccessResponse(ctx, response)
}

// All godoc
// @Tags Bus
// @Summary Get bus estimation
// @Description Put all mandatory parameter
// @Param id path string true "Terminal ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.BusInfoResponse
// @Failure 200 {object} dto.BusInfoResponse
// @Router /bus/info/{id} [post]
func (c *Controller) busInfo(ctx *fiber.Ctx) error {
	var (
		response dto.BusInfoResponse
	)

	id := ctx.Params("id")

	c.Shared.Logger.Infof("bus info, data: %s", id)

	response, err := c.Interfaces.BusViewService.BusInfo(id)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	return common.DoCommonSuccessResponse(ctx, response)
}

/**
 * Track bus location using websocket
 * @param type to differentiate between driver and client
 * @param token authentication token used only if type is driver
 * @param experimental toggler for experimnetal tracking using bot
 * @param expeerimentalId bus identifier for bot
 */
func (c *Controller) trackBusLocation(ctx *websocket.Conn) {
	defer func() {
		ctx.Close()
	}()

	query := dto.BusLocationQuery{
		Type:           ctx.Query("type", string(dto.CLIENT)),
		Token:          ctx.Query("token", ""),
		Experimental:   ctx.Query("experimental", c.Shared.Env.Experimental),
		ExperminetalID: ctx.Query("experimentalId", ""),
	}

	c.Shared.Logger.Infof("stream bus location, query: %s", query)

	for {
		if query.Type == string(dto.DRIVER) {
			data, err := c.Interfaces.BusViewService.TrackBusLocation(query, ctx)
			if err != nil {
				return
			}
			ctx.WriteJSON(data)
		} else {
			busLocation := c.Interfaces.BusViewService.StreamBusLocation(query)
			ctx.WriteJSON(busLocation)
			time.Sleep(1 * time.Second)
		}
	}
}

/**
 * Track bus location using websocket and firebase
 * @param type to differentiate between driver and client
 * @param token authentication token used only if type is driver
 * @param experimental toggler for experimnetal tracking using bot
 * @param experimentalId bus identifier for bot
 */
 func (c *Controller) trackBusLocationFirebase(ctx *websocket.Conn) {
	firebaseCtx := context.Background()
	sa := option.WithCredentialsFile(c.Shared.Env.GoogleApplicationCredentials)
	c.Shared.Logger.Infof("sasa, %s", sa)
	app, err := firebase.NewApp(firebaseCtx, nil, sa)
	if err != nil {
		c.Shared.Logger.Infof("error: %v", err)
	}

	client, err := app.Firestore(firebaseCtx)
	if err != nil {
		c.Shared.Logger.Infof("error: %v", err)
	}

	defer func() {
		client.Close()
		ctx.Close()
	}()

	query := dto.BusLocationQuery{
		Type:           ctx.Query("type", string(dto.CLIENT)),
		Token:          ctx.Query("token", ""),
		Experimental:   ctx.Query("experimental", c.Shared.Env.Experimental),
		ExperminetalID: ctx.Query("experimentalId", ""),
	}

	c.Shared.Logger.Infof("stream bus location firebase, query: %s", query)

	for {
		if query.Type == string(dto.DRIVER) {
			data, err := c.Interfaces.BusViewService.TrackBusLocationFirebase(query, ctx, client, firebaseCtx)
			if err != nil {
				return
			}
			ctx.WriteJSON(data)
		}
	}
}

func NewController(interfaces interfaces.Holder, shared shared.Holder) Controller {
	return Controller{
		Interfaces: interfaces,
		Shared:     shared,
	}
}
