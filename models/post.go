package models

type Post struct {
	Base
	RestaurantName *string `json:"restaurantName" validate:"required"`
	Title          *string `json:"title" validate:"required"`
	Description    *string `json:"description" validate:"required"`
	Image          *string `json:"image" validate:"required"`
	StartTime      *string `json:"start_time" validate:"required"`
	EndTime        *string `json:"end_time" validate:"required"`
	UrlPost        *string `json:"urlPost" validate:"required"`
	Category       *string `json:"category" validate:"required"`
	Period         *string `json:"period" validate:"required"`
	MapUrl         *string `json:"mapUrl" validate:"required"`
	Location       *string `json:"location" validate:"required"`
	Participants   *int32  `json:"participants" validate:"required"`
}
