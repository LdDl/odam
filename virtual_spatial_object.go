package odam

// SPATIAL_OBJECT_TYPE Alias to int
type SPATIAL_OBJECT_TYPE uint32

const (
	// Virtual line. See 'VirtualLine' description
	SPATIAL_VIRTUAL_LINE = EVENT_TYPE(iota + 1)
	// Virtual polygon. See 'VirtualPolygon' description
	SPATIAL_VIRTUAL_POLYGON
)

// SpatialObjectInfo Information about spatial object (either virtual line or virtual polygon)
type SpatialObjectInfo struct {
	ObjectID   int
	ObjectType SPATIAL_OBJECT_TYPE
}
