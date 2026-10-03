package service

import (
	"strings"

	"peminjaman-buku/app/model"
	"peminjaman-buku/helper"
)

func CanAccessUser(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	role string,
	perms *helper.PermissionSet,
) map[string]string {
	fields := map[string]string{}

	if !perms.IsKnownRole(role) {
		fields["role"] = "role tidak dikenal, pilih salah satu dari: " + strings.Join(perms.KnownRoles(), ", ")
	}
	if current.UserID == targetID {
		fields["role"] = "tidak boleh mengubah role diri sendiri"
	}

	return fields
}