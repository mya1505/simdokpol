package dto

type LicenseRequest struct {
	Key string `json:"key" binding:"required"`
}