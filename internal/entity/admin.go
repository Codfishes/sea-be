package entity

import "time"

type AdminUser struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	Password  string    `db:"password" json:"-"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type AdminResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *AdminUser) ToResponse() AdminResponse {
	return AdminResponse{
		ID:        a.ID,
		Email:     a.Email,
		Name:      a.Name,
		Role:      a.Role,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func (a *AdminUser) IsSuperAdmin() bool {
	return a.Role == "super_admin"
}

func (a *AdminUser) CanManageUsers() bool {
	return a.Role == "super_admin" || a.Role == "admin"
}

func (a *AdminUser) CanModerateContent() bool {
	return a.Role == "super_admin" || a.Role == "admin" || a.Role == "moderator"
}

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleModerator  = "moderator"
)
