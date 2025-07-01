package admin

import "errors"

var (
	ErrAdminNotFound      = errors.New("admin not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAdminAlreadyExists = errors.New("admin already exists")
	ErrUnauthorizedAccess = errors.New("unauthorized access")
	ErrInvalidRole        = errors.New("invalid admin role")
	ErrInsufficientRights = errors.New("insufficient admin rights")
)
