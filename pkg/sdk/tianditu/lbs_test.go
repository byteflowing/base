package tianditu

import (
	"context"
	"testing"

	"log"

	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

func TestRoutePlan(t *testing.T) {
	req := &mapsv1.TianDiTuRoutePlanReq{
		Key: "d784e22735af8d66c7e5ef78396c1ccd",
		Orig: &mapsv1.TianDiTuLocation{
			Lat: 39.92277,
			Lng: 116.35506,
		},
		Dest: &mapsv1.TianDiTuLocation{
			Lat: 39.90854,
			Lng: 116.39751,
		},
		Style: 0,
	}
	s := NewMapService()
	ctx := context.Background()
	resp, err := s.RoutePlan(ctx, req)
	log.Println(resp, err)
}
