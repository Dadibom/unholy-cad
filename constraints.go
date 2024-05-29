package main

import (
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
	id int
	cornerPointId int
	linePoint1Id int
	linePoint2Id int
	angle float64
}

func (c *SketchConstraintCornerAngle) getBranches() int {
	return 2
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

func (c *SketchConstraintCornerAngle) apply(g *Game, branch int) bool {
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

	currentAngle := c.GetCurrentAngle(g)
	offset := c.angle - currentAngle

	pointToRotate := linePoint1
	if branch == 1 {
		pointToRotate = linePoint2
		offset = -offset
	}

	pointToRotate.position = pointToRotate.position.rotateAround(cornerPoint.position, offset * math.Pi / 180)

	// @TODO implement branches 2 and 3 which alters the length of either lines

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
		// @TODO
	}
}






type SketchConstraintLineLength struct {
	id int
	lineId int
	length float64
}

func isNearZero (v float64) bool {
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
	} else { // Move startPoint
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

	StrokeLine(screen, startPosition, startPosition.add(tangent.mul(offset + 5)), 1, col)
	StrokeLine(screen, endPosition, endPosition.add(tangent.mul(offset + 5)), 1, col)

	startPosition = startPosition.add(tangent)
	endPosition = endPosition.add(tangent)
	midPoint := startPosition.lerp(endPosition, 0.5).add(tangent.mul(offset))

	g.drawArrow(screen, midPoint, startPosition.add(tangent.mul(offset)).add(direction.mul(2.0)), col, camera)
	g.drawArrow(screen, midPoint, endPosition.add(tangent.mul(offset)).sub(direction.mul(2.0)), col, camera)
}

func (c *SketchConstraintLineLength) getId() int {
	return c.id
}

func (c *SketchConstraintLineLength) getBranches() int {
	return 2
}