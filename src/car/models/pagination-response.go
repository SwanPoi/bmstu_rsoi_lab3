package models

type PaginationResponse struct {
    Page          int             `json:"page"`
    PageSize      int             `json:"pageSize"`
    TotalElements int             `json:"totalElements"`
    Items         []CarResponse   `json:"items"`
}