package main

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	mplusFaceSource *text.GoTextFaceSource
	mplusNormalFace *text.GoTextFace
	mplusBigFace    *text.GoTextFace
)

func initFonts() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s

	mplusNormalFace = &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   14,
	}
	mplusBigFace = &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   18,
	}
}

func DrawText(dst *ebiten.Image, str string, pos Vec2, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(pos.x, pos.y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, str, mplusNormalFace, op)
}
