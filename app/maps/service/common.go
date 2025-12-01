package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

var availableInterfaces = map[enumsv1.MapSource][]enumsv1.MapInterfaceType{
	enumsv1.MapSource_MAP_SOURCE_AMAP: []enumsv1.MapInterfaceType{
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_CONVERT,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_TO_ADDR,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ADDR_TO_LOCATION,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_WALKING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_BICYCLING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_DRIVING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_E_BICYCLING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_TRANSIT,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_LOCATION_BY_IP,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTANCE_MEASURE,
	},
	enumsv1.MapSource_MAP_SOURCE_TENCENT: []enumsv1.MapInterfaceType{
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS_SEARCH,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS_LIST,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS_CHILDREN,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_CONVERT,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_TO_ADDR,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ADDR_TO_LOCATION,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_WALKING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_BICYCLING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_DRIVING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_E_BICYCLING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_TRANSIT,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_WALKING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_BICYCLING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_DRIVING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_LOCATION_BY_IP,
	},
	enumsv1.MapSource_MAP_SOURCE_TIAN_DI_TU: []enumsv1.MapInterfaceType{
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_DISTRICTS,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_TO_ADDR,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ADDR_TO_LOCATION,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_DRIVING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_TRANSIT,
	},
	enumsv1.MapSource_MAP_SOURCE_HUAWEI: []enumsv1.MapInterfaceType{
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_LOCATION_TO_ADDR,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ADDR_TO_LOCATION,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_WALKING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_BICYCLING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_ROUTE_PLAN_DRIVING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_WALKING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_BICYCLING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_MATRIX_PLAN_DRIVING,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_LOCATION_BY_IP,
		enumsv1.MapInterfaceType_MAP_INTERFACE_TYPE_GET_TIMEZONE_BY_LOCATION,
	},
}

func getAvailableInterfaces(source enumsv1.MapSource) ([]enumsv1.MapInterfaceType, error) {
	types, ok := availableInterfaces[source]
	if !ok {
		return nil, ecode.ErrMapsSourceNotSupported
	}
	return types, nil
}

func hasInterface(source enumsv1.MapSource, interfaceType enumsv1.MapInterfaceType) bool {
	interfaces, ok := availableInterfaces[source]
	if !ok {
		return false
	}
	for _, i := range interfaces {
		if i == interfaceType {
			return true
		}
	}
	return false
}

func isDBNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	return false
}
