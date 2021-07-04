package tracer

// Material represents a material that produces a scattered ray
type Material struct {
	Color        Color
	RefrIndex    float64
	Reflectivity float64
	Roughness    float64
	Emittance    float64
	Transparent  bool
	Lambert      bool
	Normal       bool
}

func NormalMaterial() Material {
	return Material{NewColor(0, 0, 0), 1, 0, 0, 0, false, false, true}
}

// DiffuseMaterial returns a diffuse material
func DiffuseMaterial(color Color) Material {
	return Material{color, 1, 0, 0, 0, false, false, false}
}

// LambertMaterial returns a lambertian material
func LambertMaterial(albedo Color) Material {
	return Material{albedo, 1, 0, 0, 0, false, true, false}
}

// MetalicMaterial returns a metalic (reflective) material
func MetalicMaterial(albedo Color, reflectivity, roughness float64) Material {
	return Material{albedo, 1, reflectivity, roughness, 0, false, false, false}
}

// DielectricMaterial returns a dielectric (refractive) material
func DielectricMaterial(index float64) Material {
	return Material{NewColor(0, 0, 0), index, 0, 0, 0, true, false, false}
}

func LightMaterial(intensity Color, emittance float64) Material {
	return Material{intensity, 1, 0, 0, emittance, false, false, false}
}
