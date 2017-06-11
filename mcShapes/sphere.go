package mcshapes

import (
	"io"
	"math"
)

// Sphere is a hollow sphere defined by a center
// point and a radius with a given surface
type Sphere struct {
	surface string
	radius  int
	center  XYZ
}

// NewSphere creates a new sphere
func NewSphere(opts ...SphereOption) *Sphere {
	s := &Sphere{
		//default surface is "minecraft:glass"
		surface: "minecraft:glass",
		//default radius 30
		radius: 30,
		//default center to bring whole sphere on surface
		center: XYZ{Y: 30},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// SphereOption sets various options for NewSphere
type SphereOption func(*Sphere)

// WithRadius set the radius of the sphere
func WithRadius(r int) SphereOption {
	return func(s *Sphere) { s.radius = r }
}

// WithSphereSurface set the surface of the sphere
func WithSphereSurface(surface string) SphereOption {
	return func(s *Sphere) { s.surface = surface }
}

// WithCenter set the center point of the sphere
// note that the center should be at least radius
// for Y, otherwise sphere is below ground level
func WithCenter(c XYZ) SphereOption {
	return func(s *Sphere) { s.center = c }
}

// WriteShape satisfies ObjectWriter interface
func (s *Sphere) WriteShape(w io.Writer) error {
	var voxels []ObjectWriter
	for x := -s.radius; x <= s.radius; x++ {
		for y := -s.radius; y <= s.radius; y++ {
			for z := -s.radius; z <= s.radius; z++ {
				sqs := math.Pow(float64(x), 2) +
					math.Pow(float64(y), 2) +
					math.Pow(float64(z), 2)
				outline := math.Sqrt(sqs)
				if outline >= float64(s.radius-2) && outline <= float64(s.radius) {
					b := NewBox(
						At(XYZ{X: x + s.center.X, Y: y + s.center.Y, Z: z + s.center.Z}),
						WithSurface(s.surface))
					voxels = append(voxels, b)
				}
			}
		}
	}
	return WriteShapes(w, voxels)
}