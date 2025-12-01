package service

import (
	"github.com/byteflowing/base/app/maps/dal/model"
	"github.com/byteflowing/base/pkg/utils/trans"
	"github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func accountModelToMapInfo(m *model.MapAccount) *mapsv1.MapInfo {
	info := &mapsv1.MapInfo{
		Id:        m.ID,
		Name:      m.Name,
		Source:    enumsv1.MapSource(m.MapSource),
		Type:      enumsv1.MapType(m.MapType),
		OwnerType: enumsv1.MapOwnerType(m.OwnerType),
		OwnerId:   m.ObjectID,
		Comment:   trans.Deref(m.Comment),
		Status:    enumsv1.MapStatus(m.Status),
		Key:       m.Key,
	}
	if m.CreatedAt != nil {
		info.CreatedAt = timestamppb.New(*m.CreatedAt)
	}
	if m.UpdatedAt != nil {
		info.UpdatedAt = timestamppb.New(*m.UpdatedAt)
	}
	if m.DeletedAt.Valid {
		info.DeletedAt = timestamppb.New(m.DeletedAt.Time)
	}
	return info
}

func interfaceModelToInfo(m *model.MapInterface, accountM *model.MapAccount) *mapsv1.MapInterfaceInfo {
	info := &mapsv1.MapInterfaceInfo{
		Id:            m.ID,
		MapId:         m.MapID,
		Source:        enumsv1.MapSource(accountM.MapSource),
		MapType:       enumsv1.MapType(accountM.MapType),
		InterfaceType: enumsv1.MapInterfaceType(m.InterfaceType),
		OwnerType:     enumsv1.MapOwnerType(accountM.OwnerType),
		ObjectId:      accountM.ObjectID,
		SecondLimit:   trans.Deref(m.SecondLimit),
		DailyLimit:    trans.Deref(m.DailyLimit),
	}
	if m.CreatedAt != nil {
		info.CreatedAt = timestamppb.New(*m.CreatedAt)
	}
	if m.UpdatedAt != nil {
		info.UpdatedAt = timestamppb.New(*m.UpdatedAt)
	}
	if m.DeletedAt.Valid {
		info.DeletedAt = timestamppb.New(m.DeletedAt.Time)
	}
	return info
}
