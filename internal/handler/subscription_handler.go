package handler

import (
	"errors"
	"net/http"
	"strconv"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"
	"subscription-service/internal/utils"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" validate:"required"`
	Price       int     `json:"price" validate:"gte=0"`
	UserID      string  `json:"user_id" validate:"required,uuid"`
	StartDate   string  `json:"start_date" validate:"required,datetime=01-2006"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty,datetime=01-2006"`
}

// Create godoc
// @Summary Create subscription
// @Accept json
// @Produce json
// @Param request body CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} domain.Subscription
// @Failure 400 {object} map[string]string
// @Router /api/subscriptions [post]
func (h *SubscriptionHandler) Create(c echo.Context) error {
	var req CreateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	start, err := utils.ParseMonthYear(req.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date format, use MM-YYYY")
	}
	var end *time.Time
	if req.EndDate != nil {
		e, err := utils.ParseMonthYear(*req.EndDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date format, use MM-YYYY")
		}
		// set end date to last day of month
		e = utils.LastDayOfMonth(e)
		end = &e
	}
	userUUID, _ := uuid.Parse(req.UserID)

	sub, err := h.svc.Create(c.Request().Context(), req.ServiceName, req.Price, userUUID, &start, end)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, sub)
}

// TotalCost handler
type TotalCostRequest struct {
	From        string  `query:"from" validate:"required,datetime=01-2006"`
	To          string  `query:"to" validate:"required,datetime=01-2006"`
	UserID      *string `query:"user_id" validate:"omitempty,uuid"`
	ServiceName *string `query:"service_name"`
}

func (h *SubscriptionHandler) TotalCost(c echo.Context) error {
	var req TotalCostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query parameters")
	}
	from, _ := utils.ParseMonthYear(req.From)
	toRaw, _ := utils.ParseMonthYear(req.To)
	to := utils.LastDayOfMonth(toRaw)
	var userUUID *uuid.UUID
	if req.UserID != nil {
		u, _ := uuid.Parse(*req.UserID)
		userUUID = &u
	}
	sum, err := h.svc.TotalCost(c.Request().Context(), from, to, userUUID, req.ServiceName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]int{"total_cost": sum})
}
func (h *SubscriptionHandler) GetByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subscription id")
	}
	sub, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "subscription not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, sub)
}

func (h *SubscriptionHandler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subscription id")
	}
	err = h.svc.Delete(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "subscription not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *SubscriptionHandler) List(c echo.Context) error {
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}
	subs, err := h.svc.List(c.Request().Context(), limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, subs)
}
func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" validate:"omitempty"`
	Price       *int    `json:"price,omitempty" validate:"omitempty,gte=0"`
	StartDate   *string `json:"start_date,omitempty" validate:"omitempty,datetime=01-2006"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty,datetime=01-2006"`
}

func (h *SubscriptionHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subscription id")
	}

	var req UpdateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Convert dates if provided
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t, err := utils.ParseMonthYear(*req.StartDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date format")
		}
		startDate = &t
	}
	if req.EndDate != nil {
		t, err := utils.ParseMonthYear(*req.EndDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date format")
		}
		lastDay := utils.LastDayOfMonth(t)
		endDate = &lastDay
	}

	sub, err := h.svc.Update(c.Request().Context(), id, req.ServiceName, req.Price, startDate, endDate)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "subscription not found")
		}
		if errors.Is(err, service.ErrInvalidDateRange) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, sub)
}
