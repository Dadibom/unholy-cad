package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type SketchConstraint interface {
	getId() int
	getBranches() int
	apply(g *Game, branch int) bool
	isSatisfied(g *Game) bool
}

type SketchConstraintCornerAngle struct {
	id            int
	cornerPointId int
	linePoint1Id  int
	linePoint2Id  int
	angle         float64
}

func (c *SketchConstraintCornerAngle) clone() SketchElement {
	return &SketchConstraintCornerAngle{
		id:            c.id,
		cornerPointId: c.cornerPointId,
		linePoint1Id:  c.linePoint1Id,
		linePoint2Id:  c.linePoint2Id,
		angle:         c.angle,
	}
}
func (c *SketchConstraintCornerAngle) GetCurrentAngle(g *Game) float64 {
	cornerPoint, err := getSketchElementByID[*SketchPoint](g, c.cornerPointId)
	if err != nil {
		log.Fatal(err)
	}

	linePoint1, err := getSketchElementByID[*SketchPoint](g, c.linePoint1Id)
	if err != nil {
		log.Fatal(err)
	}

	linePoint2, err := getSketchElementByID[*SketchPoint](g, c.linePoint2Id)
	if err != nil {
		log.Fatal(err)
	}

	v1 := cornerPoint.position.sub(linePoint1.position).normalize()
	v2 := cornerPoint.position.sub(linePoint2.position).normalize()

	// returns angle in degrees
	return math.Acos(v1.dot(v2)) * 180 / math.Pi
}

func (c *SketchConstraintCornerAngle) isSatisfied(g *Game) bool {
	return isNearZero(c.GetCurrentAngle(g) - c.angle)
}

func (c *SketchConstraintCornerAngle) getBranches() int {
	if c.angle == 90 { // @TODO support other angles
		return 6
	}
	return 2
}

func (cc *SketchConstraintCornerAngle) apply(g *Game, branch int) bool {
	cornerPoint, err := getSketchElementByID[*SketchPoint](g, cc.cornerPointId)
	if err != nil {
		log.Fatal(err)
	}

	linePoint1, err := getSketchElementByID[*SketchPoint](g, cc.linePoint1Id)
	if err != nil {
		log.Fatal(err)
	}

	linePoint2, err := getSketchElementByID[*SketchPoint](g, cc.linePoint2Id)
	if err != nil {
		log.Fatal(err)
	}

	currentAngle := cc.GetCurrentAngle(g)
	offset := cc.angle - currentAngle

	if branch == 0 || branch == 1 {
		// rotate one of the lines around the corner point
		pointToRotate := linePoint1
		if branch == 1 {
			pointToRotate = linePoint2
			offset = -offset
		}

		pointToRotate.position = pointToRotate.position.rotateAround(cornerPoint.position, -offset*math.Pi/180)
	} else if branch == 2 || branch == 3 { // branch 2 and 3, move corner along one of the lines to reach 90 degrees
		o1 := linePoint1.position.sub(cornerPoint.position)
		o2 := linePoint2.position.sub(cornerPoint.position)
		if branch == 2 {
			// move corner along line1
			mag := o2.dot(o1.normalize())
			if mag <= o1.magnitude() {
				//return false
			}
			cornerPoint.position = cornerPoint.position.add(o1.normalize().mul(mag))
		} else {
			// move corner along line2
			mag := o1.dot(o2.normalize())
			if mag <= o2.magnitude() {
				//return false
			}
			cornerPoint.position = cornerPoint.position.add(o2.normalize().mul(mag))
		}
	} else if branch == 4 || branch == 5 {
		// rotate the corner around one of the lines
		// A = corner, B = pivot, C = third point
		// rotate A around B
		A := cornerPoint.position
		B := linePoint1.position
		C := linePoint2.position
		if branch == 5 {
			B = linePoint2.position
			C = linePoint1.position
		}

		c := B.sub(A).magnitude()
		a := C.sub(B).magnitude()

		radians := cc.angle * math.Pi / 180
		
		// get the required length of the line between A and C to make the desired corner angle
		b := math.Sqrt(a*a + c*c - 2*a*c*math.Cos(radians))

		cornerPoint.position = C.sub(B).normalize().rotate(math.Pi/2).mul(b).add(B) // @TODO could make this negative too
	}

	return true
}

func (c *SketchConstraintCornerAngle) getId() int {
	return c.id
}

func (c *SketchConstraintCornerAngle) draw(g *Game, screen *ebiten.Image, camera Camera) {
	col := color.RGBA{0x11, 0x11, 0x11, 0xFF}
	if !c.isSatisfied(g) {
		col = color.RGBA{0xFF, 0x00, 0x00, 0xFF}
	}

	cornerPoint, err := getSketchElementByID[*SketchPoint](g, c.cornerPointId)
	if err != nil {
		log.Fatal(err)
	}

	linePoint1, err := getSketchElementByID[*SketchPoint](g, c.linePoint1Id)
	if err != nil {
		log.Fatal(err)
	}

	linePoint2, err := getSketchElementByID[*SketchPoint](g, c.linePoint2Id)
	if err != nil {
		log.Fatal(err)
	}

	if c.angle == 90 {
		offset := 15.0

		cp := camera.transformPoint(cornerPoint.position)
		point1 := camera.transformPoint(linePoint1.position)
		point2 := camera.transformPoint(linePoint2.position)

		o1 := point1.sub(camera.transformPoint(cornerPoint.position)).normalize().mul(offset)
		o2 := point2.sub(camera.transformPoint(cornerPoint.position)).normalize().mul(offset)

		StrokeLine(screen, cp.add(o1), cp.add(o1).add(o2), 1, col)
		StrokeLine(screen, cp.add(o2), cp.add(o1).add(o2), 1, col)
	} else {
		center := camera.transformPoint(cornerPoint.position)
		radius := 20.0
		angle1 := cornerPoint.position.sub(linePoint1.position).angle()
		angle2 := cornerPoint.position.sub(linePoint2.position).angle()
		StrokeArc(screen, center, radius, angle1, angle2, 1, col)

		midPointAngle := (angle1 + angle2) / 2

		mPoint := center.add(Vec2{math.Cos(midPointAngle), math.Sin(midPointAngle)}.mul(radius))

		DrawText(screen, fmt.Sprintf("%.0f°", c.angle), mPoint, col)
	}

}

type SketchConstraintLineLength struct {
	id     int
	lineId int
	length float64
}

func (c *SketchConstraintLineLength) clone() SketchElement {
	return &SketchConstraintLineLength{
		id:     c.id,
		lineId: c.lineId,
		length: c.length,
	}
}

func isNearZero(v float64) bool {
	return math.Abs(v) < 0.00001
}

func (c *SketchConstraintLineLength) isSatisfied(g *Game) bool {
	line, err := getSketchElementByID[*SketchLine](g, c.lineId)
	if err != nil {
		log.Fatal(err)
	}

	startPoint, err := getSketchElementByID[*SketchPoint](g, line.startId)
	if err != nil {
		log.Fatal(err)
	}
	endPoint, err := getSketchElementByID[*SketchPoint](g, line.endId)
	if err != nil {
		log.Fatal(err)
	}

	return isNearZero(startPoint.position.distanceTo(endPoint.position) - c.length)
}

func (c *SketchConstraintLineLength) apply(g *Game, branch int) bool {
	line, err := getSketchElementByID[*SketchLine](g, c.lineId)
	if err != nil {
		log.Fatal(err)
	}

	startPoint, err := getSketchElementByID[*SketchPoint](g, line.startId)
	if err != nil {
		log.Fatal(err)
	}
	endPoint, err := getSketchElementByID[*SketchPoint](g, line.endId)
	if err != nil {
		log.Fatal(err)
	}

	currentLength := startPoint.position.distanceTo(endPoint.position)
	t := c.length / currentLength

	if branch == 0 { // Move endPoint
		endPoint.position = startPoint.position.lerp(endPoint.position, t)
	} else if branch == 1 { // Move startPoint
		startPoint.position = endPoint.position.lerp(startPoint.position, t)
	}

	return true
}

func (c *SketchConstraintLineLength) draw(g *Game, screen *ebiten.Image, camera Camera) {
	col := color.RGBA{0x11, 0x11, 0x11, 0xFF}
	if !c.isSatisfied(g) {
		col = color.RGBA{0xFF, 0x00, 0x00, 0xFF}
	}

	line, err := getSketchElementByID[*SketchLine](g, c.lineId)
	if err != nil {
		log.Fatal(err)
	}

	startPoint, err := getSketchElementByID[*SketchPoint](g, line.startId)
	if err != nil {
		log.Fatal(err)
	}
	endPoint, err := getSketchElementByID[*SketchPoint](g, line.endId)
	if err != nil {
		log.Fatal(err)
	}

	startPosition := camera.transformPoint(startPoint.position)
	endPosition := camera.transformPoint(endPoint.position)

	direction := endPosition.sub(startPosition).normalize()
	tangent := direction.tangent()

	offset := 12.0

	StrokeLine(screen, startPosition, startPosition.add(tangent.mul(offset+5)), 1, col)
	StrokeLine(screen, endPosition, endPosition.add(tangent.mul(offset+5)), 1, col)

	startPosition = startPosition.add(tangent)
	endPosition = endPosition.add(tangent)
	midPoint := startPosition.lerp(endPosition, 0.5).add(tangent.mul(offset))

	g.drawArrow(screen, midPoint, startPosition.add(tangent.mul(offset)).add(direction.mul(2.0)), col, camera)
	g.drawArrow(screen, midPoint, endPosition.add(tangent.mul(offset)).sub(direction.mul(2.0)), col, camera)

	DrawText(screen, "L="+fmt.Sprintf("%.2f", c.length), midPoint.add(tangent.mul(5)), col)
}

func (c *SketchConstraintLineLength) getId() int {
	return c.id
}

func (c *SketchConstraintLineLength) getBranches() int {
	return 2
}
