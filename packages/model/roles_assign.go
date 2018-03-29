package model

import (
	"fmt"
	"time"
)

type RolesAssign struct {
	prefix          int64
	Id              int64
	RoleID          int64
	RoleType        int64
	RoleName        string
	MemberId        int64
	MemberName      string
	MemberAvatar    []byte
	AppointedById   int64
	AppointedByName string
	DateStart       time.Time
	DateEnd         time.Time
	Delete          bool
}

// SetTablePrefix is setting table prefix
func (r *RolesAssign) SetTablePrefix(prefix int64) *RolesAssign {
	if prefix == 0 {
		prefix = 1
	}
	r.prefix = prefix
	return r
}

// TableName returns name of table
func (r RolesAssign) TableName() string {
	if r.prefix == 0 {
		r.prefix = 1
	}
	return fmt.Sprintf("%d_roles_assign", r.prefix)
}

func (r *RolesAssign) GetActiveMemberRoles(memberID int64) ([]RolesAssign, error) {
	roles := new([]RolesAssign)
	err := DBConn.Table(r.TableName()).Where("member_id = ? AND delete = ?", memberID, 0).Find(&roles).Error
	return *roles, err
}
