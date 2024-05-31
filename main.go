package main

import (
	"errors"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	gridSize     = 1
)

type Camera struct {
	position Vec2
	scale    float64
}

type Game struct {
	camera       Camera
	lastMousePos Vec2
	isDragging   bool

	sketch Sketch
}

type SketchElement interface {
	getId() int
	draw(g *Game, screen *ebiten.Image, camera Camera)
	clone() SketchElement
}

type SketchLine struct {
	id      int
	startId int
	endId   int
}

func (l *SketchLine) clone() SketchElement {
	return &SketchLine{
		id:      l.id,
		startId: l.startId,
		endId:   l.endId,
	}
}

func (l *SketchLine) draw(g *Game, screen *ebiten.Image, camera Camera) {
	startPoint, err := getSketchElementByID[*SketchPoint](g, l.startId)
	if err != nil {
		log.Fatal(err)
	}
	endPoint, err := getSketchElementByID[*SketchPoint](g, l.endId)
	if err != nil {
		log.Fatal(err)
	}

	g.drawLineWithThickness(screen, endPoint.position, startPoint.position, color.RGBA{0x33, 0x99, 0xff, 0xFF}, camera, 2)
}

func (l *SketchLine) getId() int {
	return l.id
}

type SketchPoint struct {
	id       int
	position Vec2
}

func (p *SketchPoint) clone() SketchElement {
	return &SketchPoint{
		id:       p.id,
		position: p.position.clone(),
	}
}

func (p *SketchPoint) draw(g *Game, screen *ebiten.Image, camera Camera) {
	g.drawCircle(screen, p.position, float32(3), color.RGBA{0x33, 0x99, 0xff, 0xFF}, camera)
}

func (p *SketchPoint) getId() int {
	return p.id
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		g.sketch.attemptApplyConstraints(g)
	}
	// randomly move the position of the point on x key press
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		for _, element := range g.sketch.elements {
			if point, ok := element.(*SketchPoint); ok {
				point.position.x += rand.Float64()*2 - 1
				point.position.y += rand.Float64()*2 - 1
			}
		}
	}

	mouseX, mouseY := ebiten.CursorPosition()
	mouseVec := Vec2{float64(mouseX), float64(mouseY)}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.isDragging {
			delta := Vec2{
				x: mouseVec.x - g.lastMousePos.x,
				y: mouseVec.y - g.lastMousePos.y,
			}
			g.camera.position.x -= delta.x / g.camera.scale
			g.camera.position.y -= delta.y / g.camera.scale
		}
		g.isDragging = true
		g.lastMousePos = mouseVec
	} else {
		g.isDragging = false
	}

	// Zooming
	_, dy := ebiten.Wheel()
	if dy != 0 {
		g.zoom(mouseVec, dy)
	}

	return nil
}

func (g *Game) zoom(mousePos Vec2, scrollAmount float64) {
	previousScale := g.camera.scale
	g.camera.scale *= 1 + scrollAmount*0.1

	if g.camera.scale < 0.1 {
		g.camera.scale = 0.1
	}

	// Adjust the camera position to zoom around the cursor
	mouseWorldX := (mousePos.x / previousScale) + g.camera.position.x
	mouseWorldY := (mousePos.y / previousScale) + g.camera.position.y
	newMouseWorldX := (mousePos.x / g.camera.scale) + g.camera.position.x
	newMouseWorldY := (mousePos.y / g.camera.scale) + g.camera.position.y

	g.camera.position.x += mouseWorldX - newMouseWorldX
	g.camera.position.y += mouseWorldY - newMouseWorldY
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	g.drawGrid(screen)

	for _, element := range g.sketch.elements {
		element.draw(g, screen, g.camera)
	}
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	startX := int((g.camera.position.x)/gridSize)*gridSize - gridSize
	endX := int((g.camera.position.x+screenWidth/g.camera.scale)/gridSize)*gridSize + gridSize

	startY := int((g.camera.position.y)/gridSize)*gridSize - gridSize
	endY := int((g.camera.position.y+screenHeight/g.camera.scale)/gridSize)*gridSize + gridSize

	thickLineInterval := gridSize * 5

	// Draw lighter grid lines first
	for x := startX; x <= endX; x += gridSize {
		if x%thickLineInterval != 0 {
			g.drawLine(screen, Vec2{float64(x), float64(startY)}, Vec2{float64(x), float64(endY)}, color.RGBA{0xee, 0xee, 0xee, 0xFF}, g.camera)
		}
	}

	for y := startY; y <= endY; y += gridSize {
		if y%thickLineInterval != 0 {
			g.drawLine(screen, Vec2{float64(startX), float64(y)}, Vec2{float64(endX), float64(y)}, color.RGBA{0xee, 0xee, 0xee, 0xFF}, g.camera)
		}
	}

	// Draw darker grid lines on top
	for x := startX; x <= endX; x += gridSize {
		if x%thickLineInterval == 0 {
			g.drawLine(screen, Vec2{float64(x), float64(startY)}, Vec2{float64(x), float64(endY)}, color.RGBA{0xbb, 0xbb, 0xbb, 0xFF}, g.camera)
		}
	}

	for y := startY; y <= endY; y += gridSize {
		if y%thickLineInterval == 0 {
			g.drawLine(screen, Vec2{float64(startX), float64(y)}, Vec2{float64(endX), float64(y)}, color.RGBA{0xbb, 0xbb, 0xbb, 0xFF}, g.camera)
		}
	}
}

func (c *Camera) transformPoint(p Vec2) Vec2 {
	// round to avoid subpixel rendering
	return Vec2{
		x: math.Round((p.x - c.position.x) * c.scale),
		y: math.Round((p.y - c.position.y) * c.scale),
	}
}

func getSketchElementByID[T SketchElement](g *Game, id int) (T, error) {
	var zero T
	for _, element := range g.sketch.elements {
		if element.getId() == id {
			if specificElement, ok := element.(T); ok {
				return specificElement, nil
			}
		}
	}
	return zero, errors.New("element not found or type mismatch")
}

func (g *Game) drawLine(screen *ebiten.Image, p1, p2 Vec2, c color.Color, camera Camera) {
	g.drawLineWithThickness(screen, p1, p2, c, camera, 1)
}

func (g *Game) drawLineWithThickness(screen *ebiten.Image, p1, p2 Vec2, c color.Color, camera Camera, thickness float32) {
	p1 = camera.transformPoint(p1)
	p2 = camera.transformPoint(p2)
	vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), thickness, c, true)
}

func (g *Game) drawArrow(screen *ebiten.Image, p1, p2 Vec2, c color.Color, camera Camera) {
	direction := p2.sub(p1).normalize()
	tangent := direction.tangent().normalize()

	headWidth := 4.0
	headLength := 10.0

	// Draw the line
	vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), 1, c, true)

	arrowHeadBase := p2.sub(direction.mul(headLength))
	arrowHeadLeft := arrowHeadBase.add(tangent.mul(headWidth))
	arrowHeadRight := arrowHeadBase.sub(tangent.mul(headWidth))

	vector.StrokeLine(screen, float32(p2.x), float32(p2.y), float32(arrowHeadLeft.x), float32(arrowHeadLeft.y), 1, c, true)
	vector.StrokeLine(screen, float32(p2.x), float32(p2.y), float32(arrowHeadRight.x), float32(arrowHeadRight.y), 1, c, true)
}

func (g *Game) drawCircle(screen *ebiten.Image, p Vec2, radius float32, c color.Color, camera Camera) {
	p = camera.transformPoint(p)
	vector.StrokeCircle(screen, float32(p.x), float32(p.y), radius, 1, c, true)
}

func (g *Game) drawConstructionLine(screen *ebiten.Image, p1, p2 Vec2, c color.Color, camera Camera) {
	p1 = camera.transformPoint(p1)
	p2 = camera.transformPoint(p2)

	// Dashed line parameters
	dashLength := 10.0
	spaceLength := 10.0

	// Calculate the total distance between the points
	totalDistance := p1.distanceTo(p2)

	// Calculate the unit vector in the direction of the line
	unitVector := Vec2{
		x: (p2.x - p1.x) / totalDistance,
		y: (p2.y - p1.y) / totalDistance,
	}

	// Iterate over the total distance, drawing dashes and leaving spaces
	for distance := 0.0; distance < totalDistance; distance += dashLength + spaceLength {
		start := Vec2{
			x: p1.x + unitVector.x*distance,
			y: p1.y + unitVector.y*distance,
		}
		end := Vec2{
			x: p1.x + unitVector.x*math.Min(distance+dashLength, totalDistance),
			y: p1.y + unitVector.y*math.Min(distance+dashLength, totalDistance),
		}
		vector.StrokeLine(screen, float32(start.x), float32(start.y), float32(end.x), float32(end.y), 2, c, true)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func StrokeLine(screen *ebiten.Image, p1, p2 Vec2, thickness float32, c color.Color) {
	vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), thickness, c, true)
}

func StrokeArc(screen *ebiten.Image, p Vec2, radius, startAngle, endAngle float64, thickness float32, c color.Color) {
	p = Vec2{p.x, screenHeight - p.y}

	// draw arc using StrokeLine segments
	segments := 5

	for i := 0; i < segments; i++ {
		angle1 := startAngle + (endAngle-startAngle)*float64(i)/float64(segments)
		angle2 := startAngle + (endAngle-startAngle)*float64(i+1)/float64(segments)

		x1 := p.x + radius*math.Cos(angle1)
		y1 := p.y + radius*math.Sin(angle1)
		x2 := p.x + radius*math.Cos(angle2)
		y2 := p.y + radius*math.Sin(angle2)

		StrokeLine(screen, Vec2{x1, screenHeight - y1}, Vec2{x2, screenHeight - y2}, thickness, c)
	}
}

type Sketch struct {
	elements    []SketchElement
	constraints []SketchConstraint
}

func (s *Sketch) getClonedElements() []SketchElement {
	elements := make([]SketchElement, len(s.elements))
	for i, element := range s.elements {
		elements[i] = element.clone()
	}
	return elements
}

func (s *Sketch) getConstraints() []SketchConstraint {
	constraints := make([]SketchConstraint, 0)
	for _, element := range s.elements {
		if constraint, ok := element.(SketchConstraint); ok {
			constraints = append(constraints, constraint)
		}
	}
	return constraints
}

func (s *Sketch) attemptApplyConstraints(g *Game) {
	constraints := s.getConstraints()

	// shuffle the constraints
	/*rand.Shuffle(len(constraints), func(i, j int) {
		constraints[i], constraints[j] = constraints[j], constraints[i]
	})*/

	// int array to store the number of branches for each constraint
	branches := make([]int, len(constraints))
	currentBranches := make([]int, len(constraints))

	totalBrahchCombinations := 1

	allConstraintsSatisfied := true

	// get the number of branches for each constraint
	for i, constraint := range constraints {
		branches[i] = constraint.getBranches()
		totalBrahchCombinations *= branches[i]
		currentBranches[i] = 0
		if allConstraintsSatisfied && !constraint.isSatisfied(g) {
			allConstraintsSatisfied = false
		}
	}

	if allConstraintsSatisfied {
		log.Printf("Constraints already satisfied")
		return
	}

	attempts := 0

	log.Printf("Attempting to satisfy constraints with %d possible solutions", totalBrahchCombinations)

	for !allConstraintsSatisfied {
		attempts++
		// deep clone the sketch
		originalElements := s.getClonedElements()

		for i, constraint := range constraints {
			if constraint.isSatisfied(g) {
				continue
			}
			branch := currentBranches[i]
			//log.Printf("Applying constraint %d with branch %d", constraint.getId(), branch)
			constraint.apply(g, branch)
		}

		// check if any constraints are violated
		allConstraintsSatisfied = true
		for _, constraint := range constraints {
			if !constraint.isSatisfied(g) {
				// log error
				//log.Printf("Constraints could not satisfied, reverting")

				// revert to the previous state
				s.elements = originalElements
				allConstraintsSatisfied = false
				break
			}
		}
		if allConstraintsSatisfied {
			log.Printf("\u2713 Constraints satisfied after %d attempts", attempts)
			return
		}

		i := 0

		currentBranches[i]++
		for currentBranches[i] >= branches[i] {
			currentBranches[i] = 0
			i++
			if i >= len(currentBranches) {
				log.Printf("\u2717 No solution found after %d attempts", attempts)
				return
			}
			currentBranches[i]++
		}

	}
}

func main() {
	initFonts()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Unholy CAD")

	// square shape
	/*
		sketch := Sketch{
			elements: []SketchElement{
				&SketchPoint{position: Vec2{1, 1}, id: 0},
				&SketchPoint{position: Vec2{1, 10}, id: 1},
				&SketchPoint{position: Vec2{10, 10}, id: 2},
				&SketchPoint{position: Vec2{10, 1}, id: 3},
				&SketchPoint{position: Vec2{5, 5}, id: 4},

				&SketchLine{startId: 0, endId: 1, id: 5},
				&SketchLine{startId: 1, endId: 2, id: 6},
				&SketchLine{startId: 2, endId: 3, id: 7},
				&SketchLine{startId: 3, endId: 0, id: 8},

				&SketchConstraintLineLength{lineId: 5, length: 6, id: 9},
				//&SketchConstraintLineLength{lineId: 8, length: 8, id: 12},
				// &SketchConstraintCornerAngle{cornerPointId: 0, linePoint1Id: 3, linePoint2Id: 1, angle: 45, id: 10},
				&SketchConstraintCornerAngle{cornerPointId: 2, linePoint1Id: 1, linePoint2Id: 3, angle: 90, id: 10},
				&SketchConstraintCornerAngle{cornerPointId: 3, linePoint1Id: 2, linePoint2Id: 0, angle: 90, id: 11},
			},
		}
	*/

	// triangle shape
	sketch := Sketch{
		elements: []SketchElement{
			&SketchPoint{position: Vec2{1, 1}, id: 0},
			&SketchPoint{position: Vec2{1, 10}, id: 1},
			&SketchPoint{position: Vec2{10, 10}, id: 2},

			&SketchLine{startId: 0, endId: 1, id: 3},
			&SketchLine{startId: 1, endId: 2, id: 4},
			&SketchLine{startId: 2, endId: 0, id: 5},

			&SketchConstraintLineLength{lineId: 3, length: 6, id: 6},
			&SketchConstraintCornerAngle{cornerPointId: 0, linePoint1Id: 2, linePoint2Id: 1, angle: 60, id: 7},
			//&SketchConstraintCornerAngle{cornerPointId: 1, linePoint1Id: 0, linePoint2Id: 2, angle: 60, id: 8},
			//&SketchConstraintLineLength{lineId: 4, length: 6, id: 9},
		},
	}

	if err := ebiten.RunGame(&Game{
		camera: Camera{
			position: Vec2{0, 0},
			scale:    20,
		},
		sketch: sketch,
	}); err != nil {
		log.Fatal(err)
	}
}
