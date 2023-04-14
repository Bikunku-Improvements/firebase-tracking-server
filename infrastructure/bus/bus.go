package bus

import (
	"time"
	"tracking-server/interfaces"
	"tracking-server/shared"
	"tracking-server/shared/dto"

	"tracking-server/shared/common"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Controller struct {
	Interfaces interfaces.Holder
	Shared     shared.Holder
}

func (c *Controller) Routes(app *fiber.App) {
	bus := app.Group("/bus")
	bus.Post("/", c.create)
	bus.Post("/login", c.login)
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
	bus.Post("/track", c.trackBusLocationFirebase)
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

// All godoc
// @Tags Bus
// @Summary Track bus location using firebase
// @Description Put all mandatory parameter
// @Param id path string true "Bus ID"
// @Param auth header string true "token"
// @Param BusLocationMessageFirebase body dto.BusLocationMessageFirebase true "BusLocationMessageFirebase"
// @Accept json
// @Produce json
// @Router /bus/track [post]
func (c *Controller) trackBusLocationFirebase(ctx *fiber.Ctx) error {
	var (
		body dto.BusLocationMessage
	)

	err := common.DoCommonRequest(ctx, &body)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	auth := ctx.Get("auth")

	c.Shared.Logger.Infof("track bus, data: %s, token: %s", body, auth)

	err = c.Interfaces.BusViewService.TrackBusLocationFirebase(body, auth)
	if err != nil {
		return common.DoCommonErrorResponse(ctx, err)
	}

	return common.DoCommonSuccessResponse(ctx, nil)
}

func NewController(interfaces interfaces.Holder, shared shared.Holder) Controller {
	return Controller{
		Interfaces: interfaces,
		Shared:     shared,
	}
}
