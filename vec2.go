package main

import (
	"math"
)

type Vec2 struct {
	x float64
	y float64
}

func (v Vec2) distanceTo(other Vec2) float64 {
	dx := v.x - other.x
	dy := v.y - other.y
	return math.Sqrt(dx*dx + dy*dy)
}

func (v Vec2) lerp(other Vec2, t float64) Vec2 {
	return Vec2{
		x: v.x + (other.x-v.x)*t,
		y: v.y + (other.y-v.y)*t,
	}
}

func (v Vec2) add(other Vec2) Vec2 {
	return Vec2{
		x: v.x + other.x,
		y: v.y + other.y,
	}
}

func (v Vec2) sub(other Vec2) Vec2 {
	return Vec2{
		x: v.x - other.x,
		y: v.y - other.y,
	}
}

func (v Vec2) mul(scalar float64) Vec2 {
	return Vec2{
		x: v.x * scalar,
		y: v.y * scalar,
	}
}

func (v Vec2) div(scalar float64) Vec2 {
	return Vec2{
		x: v.x / scalar,
		y: v.y / scalar,
	}
}

func (v Vec2) dot(other Vec2) float64 {
	return v.x*other.x + v.y*other.y
}

func (v Vec2) magnitude() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

func (v Vec2) rotateAround(origin Vec2, angle float64) Vec2 {
	sin := math.Sin(angle)
	cos := math.Cos(angle)

	// translate point back to origin:
	v.x -= origin.x
	v.y -= origin.y

	// rotate point
	xnew := v.x*cos - v.y*sin
	ynew := v.x*sin + v.y*cos

	// translate point back:
	v.x = xnew + origin.x
	v.y = ynew + origin.y
	return v
}

func (v Vec2) normalize() Vec2 {
	amag := 1 / v.magnitude()
	return Vec2{
		x: v.x * amag,
		y: v.y * amag,
	}
}

func (v Vec2) tangent() Vec2 {
	return Vec2{
		x: -v.y,
		y: v.x,
	}
}

func (v Vec2) clone() Vec2 {
	return Vec2{
		x: v.x,
		y: v.y,
	}
}

func (v Vec2) angle() float64 {
	return math.Atan2(v.y, -v.x)
}
