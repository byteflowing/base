package tianditu

import (
	"encoding/xml"

	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

type RoutePlanResp struct {
	XMLName     xml.Name    `xml:"result"`
	Orig        string      `xml:"orig,attr"`
	Mid         string      `xml:"mid,attr"`
	Dest        string      `xml:"dest,attr"`
	Parameters  *Parameters `xml:"parameters"`
	Routes      *Routes     `xml:"routes"`
	Simple      *Simple     `xml:"simple"`
	Distance    float64     `xml:"distance"`
	Duration    float64     `xml:"duration"`
	RouteLatLon string      `xml:"routelatlon"`
	MapInfo     *MapInfo    `xml:"mapinfo"`
}

type Parameters struct {
	Orig    string `xml:"orig"`
	Dest    string `xml:"dest"`
	Mid     string `xml:"mid"`
	Key     string `xml:"key"`
	Width   int    `xml:"width"`
	Height  int    `xml:"height"`
	Style   int    `xml:"style"`
	Version string `xml:"version"`
	Sort    string `xml:"sort"`
}

type Routes struct {
	Count int         `xml:"count,attr"`
	Time  float64     `xml:"time,attr"`
	Items []RouteItem `xml:"item"`
}

type RouteItem struct {
	ID             int    `xml:"id,attr"`
	StrGuide       string `xml:"strguide"`
	Signage        string `xml:"signage"`
	StreetName     string `xml:"streetName"`
	NextStreetName string `xml:"nextStreetName"`
	TollStatus     int    `xml:"tollStatus"`
	TurnLatLon     string `xml:"turnlatlon"`
}

type Simple struct {
	Items []SimpleItem `xml:"item"`
}

type SimpleItem struct {
	ID             int     `xml:"id,attr"`
	StrGuide       string  `xml:"strguide"`
	StreetNames    string  `xml:"streetNames"`
	LastStreetName string  `xml:"lastStreetName"`
	LinkStreetName string  `xml:"linkStreetName"`
	Signage        string  `xml:"signage"`
	TollStatus     int     `xml:"tollStatus"`
	TurnLatLon     string  `xml:"turnlatlon"`
	StreetLatLon   string  `xml:"streetLatLon"`
	StreetDistance float64 `xml:"streetDistance"`
	SegmentNumber  string  `xml:"segmentNumber"`
}

type MapInfo struct {
	Center string `xml:"center"`
	Scale  int    `xml:"scale"`
}

func routePlanRespToProtoResp(v *RoutePlanResp) *mapsv1.TianDiTuRoutePlanResp {
	resp := &mapsv1.TianDiTuRoutePlanResp{
		Orig:        v.Orig,
		Mid:         v.Mid,
		Dest:        v.Dest,
		Distance:    v.Distance,
		Duration:    v.Duration,
		Routelatlon: v.RouteLatLon,
	}
	if v.Parameters != nil {
		resp.Parameters = &mapsv1.TianDiTuParameters{
			Orig:    v.Parameters.Orig,
			Dest:    v.Parameters.Dest,
			Mid:     v.Parameters.Mid,
			Key:     v.Parameters.Key,
			Width:   int32(v.Parameters.Width),
			Height:  int32(v.Parameters.Height),
			Style:   int32(v.Parameters.Style),
			Version: v.Parameters.Version,
			Sort:    v.Parameters.Sort,
		}
	}
	if v.Routes != nil {
		items := make([]*mapsv1.TianDiTuRouteItem, 0, len(v.Routes.Items))
		for _, item := range v.Routes.Items {
			items = append(items, &mapsv1.TianDiTuRouteItem{
				Id:             int32(item.ID),
				StrGuide:       item.StrGuide,
				Signage:        item.Signage,
				StreetName:     item.StreetName,
				NextStreetName: item.NextStreetName,
				TollStatus:     int32(item.TollStatus),
				TurnLatlon:     item.TurnLatLon,
			})
		}
		resp.Routes = &mapsv1.TianDiTuRoutes{
			Count: int32(v.Routes.Count),
			Time:  v.Routes.Time,
			Items: items,
		}
	}
	if v.Simple != nil {
		simpleItems := make([]*mapsv1.TianDiTuSimpleItem, 0, len(v.Simple.Items))
		for _, item := range v.Simple.Items {
			simpleItems = append(simpleItems, &mapsv1.TianDiTuSimpleItem{
				Id:             int32(item.ID),
				StrGuide:       item.StrGuide,
				StreetNames:    item.StreetNames,
				LastStreetName: item.LastStreetName,
				LinkStreetName: item.LinkStreetName,
				Signage:        item.Signage,
				TollStatus:     int32(item.TollStatus),
				TurnLatlon:     item.TurnLatLon,
				StreetLatlon:   item.StreetLatLon,
				StreetDistance: item.StreetDistance,
				SegmentNumber:  item.SegmentNumber,
			})
		}
		resp.Simple = &mapsv1.TianDiTuSimple{
			Items: simpleItems,
		}
	}
	if v.MapInfo != nil {
		resp.Mapinfo = &mapsv1.TianDiTuMapInfo{
			Center: v.MapInfo.Center,
			Scale:  int32(v.MapInfo.Scale),
		}
	}
	return resp
}
