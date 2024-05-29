package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	gridSize     = 20
)

type Vec2 struct {
	x float64
	y float64
}

type Camera struct {
	position Vec2
	scale    float64
}

type Game struct {
	camera       Camera
	lastMousePos Vec2
	isDragging   bool
}

func (g *Game) Update() error {
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
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	startX := int((g.camera.position.x)/gridSize) * gridSize
	endX := int((g.camera.position.x+screenWidth/g.camera.scale)/gridSize)*gridSize + gridSize

	startY := int((g.camera.position.y)/gridSize) * gridSize
	endY := int((g.camera.position.y+screenHeight/g.camera.scale)/gridSize)*gridSize + gridSize

	// Draw lighter grid lines first
	for x := startX; x <= endX; x += gridSize {
		if x%200 != 0 {
			g.drawLine(screen, Vec2{float64(x), float64(startY)}, Vec2{float64(x), float64(endY)}, color.RGBA{0xee, 0xee, 0xee, 0xFF}, g.camera)
		}
	}

	for y := startY; y <= endY; y += gridSize {
		if y%200 != 0 {
			g.drawLine(screen, Vec2{float64(startX), float64(y)}, Vec2{float64(endX), float64(y)}, color.RGBA{0xee, 0xee, 0xee, 0xFF}, g.camera)
		}
	}

	// Draw darker grid lines on top
	for x := startX; x <= endX; x += gridSize {
		if x%200 == 0 {
			g.drawLine(screen, Vec2{float64(x), float64(startY)}, Vec2{float64(x), float64(endY)}, color.RGBA{0xbb, 0xbb, 0xbb, 0xFF}, g.camera)
		}
	}

	for y := startY; y <= endY; y += gridSize {
		if y%200 == 0 {
			g.drawLine(screen, Vec2{float64(startX), float64(y)}, Vec2{float64(endX), float64(y)}, color.RGBA{0xbb, 0xbb, 0xbb, 0xFF}, g.camera)
		}
	}
}

func (c *Camera) transformPoint(p Vec2) Vec2 {
	return Vec2{
		x: (p.x - c.position.x) * c.scale,
		y: (p.y - c.position.y) * c.scale,
	}
}

func (g *Game) drawLine(screen *ebiten.Image, p1, p2 Vec2, c color.Color, camera Camera) {
	p1 = camera.transformPoint(p1)
	p2 = camera.transformPoint(p2)
	vector.StrokeLine(screen, float32(p1.x), float32(p1.y), float32(p2.x), float32(p2.y), 1, c, true)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Unholy CAD")
	if err := ebiten.RunGame(&Game{
		camera: Camera{
			position: Vec2{0, 0},
			scale:    1,
		},
	}); err != nil {
		log.Fatal(err)
	}
}
