package main

import (
//	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	_ "image/png"
	"log"
	"math"
	"math/rand"
)

const (
	WIDTH  = 640
	HEIGHT = 480
)

var img *ebiten.Image
var game *Game

func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("gopher.png")
	if err != nil {
		log.Fatal(err)
	}
	game = NewGame()
}

func NewGame() *Game {
	var g Game
	g.gophers = make([]*Gopher, 10)
	for i := range g.gophers {
		g.gophers[i] = NewGopher()
	}
	return &g
}

func NewGopher() *Gopher {
	var g Gopher
	g.control.p = 1.
	g.control.d = 0.7
	g.control.i = 0.
	g.pos = Vec{rand.Float64() * WIDTH, rand.Float64() * HEIGHT}
	g.PickGoal(Vec{float64(WIDTH/2), float64(HEIGHT/2)})
	return &g
}

type Vec struct {
	x float64
	y float64
}

func Len(v Vec) float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

func (v *Vec) Normalize(a Vec) {
	l := Len(a)
	v.x = a.x / l
	v.y = a.y / l
}

func (v *Vec) Mult(a Vec, b float64) {
	v.x = a.x * b
	v.y = a.y * b
}

func (v *Vec) Sub(a Vec, b Vec) {
	v.x = a.x - b.x
	v.y = a.y - b.y
}

func (v *Vec) Add(a Vec, b Vec) {
	v.x = a.x + b.x
	v.y = a.y + b.y
}

func (v *Vec) Clamp(a Vec, b float64) {
	m := math.Min(b, Len(a))
	if m == 0. {
		v.x = 0.
		v.y = 0.
		return
	}
	a.Normalize(a)
	v.Mult(a, m)
}

type Control struct {
	p  float64
	d  float64
	i  float64
	e  Vec
	de Vec
	ie Vec
	o  Vec
}

type Gopher struct {
	goal    Vec
	pos     Vec
	vel     Vec
	acc     Vec
	control Control
	theta   float64
}

type Game struct {
	gophers []*Gopher
}

func (g *Gopher) PickGoal(cog Vec) {
	// pick a new goal randomly distributed around a central point. The central
	// point is a weighted blend between the center of the screen and the center
	// of gravity of the flock
	var dist_from_center Vec
	center := Vec{float64(WIDTH/2), float64(HEIGHT/2)}
	dist_from_center.Sub(center, cog)
	d := Len(dist_from_center) / float64(HEIGHT/2)
	a := math.Min(d, 1.0)
	var b, c, target Vec
	b.Mult(center, a)
	c.Mult(cog, 1.0 - a)
	target.Add(b, c)

	g.goal.x = target.x + (-0.2 + rand.Float64()*0.4) * WIDTH
	g.goal.y = target.y + (-0.2 + rand.Float64()*0.4) * HEIGHT
	g.goal.x = math.Max(math.Min(g.goal.x, float64(WIDTH)), 0.)
	g.goal.y = math.Max(math.Min(g.goal.y, float64(HEIGHT)), 0.)
}

func (c *Control) Update(e Vec) {
	// a proportional, differential, integral control algorithm output is a
	// weighted sum of input, change in input, and accumulated input
	c.de.Sub(e, c.e)
	c.e = e
	c.ie.Add(e, c.ie)
	var p, d, i Vec
	p.Mult(e, c.p)
	d.Mult(c.de, c.d)
	i.Mult(c.ie, c.i)
	c.o.Add(p, d)
	c.o.Add(c.o, i)
}

func (g *Gopher) Update(cog Vec) error {
	var v Vec
	v.Add(g.vel, g.acc)
	g.vel.Clamp(v, 6.0)
	g.pos.Add(g.pos, g.vel)

	if rand.Int()%100 < 5 {
		g.PickGoal(cog)
	}
	var diff Vec
	diff.Sub(g.goal, g.pos)
	g.control.Update(diff)
	g.acc.Clamp(g.control.o, 1.0)
	return nil
}

func (g *Game) Update() error {
	// find the center of gravity of the gophers
	var cog Vec;
	for i:= range g.gophers {
		cog.Add(cog, g.gophers[i].pos);
	}
	cog.Mult(cog, 1./float64(len(g.gophers)))

	for i:= range g.gophers {
		g.gophers[i].Update(cog)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ib := img.Bounds()
	for _, gg := range g.gophers {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-float64(ib.Dx())/2., -float64(ib.Dy())/2.)
		op.GeoM.Scale(0.2, 0.2)
		op.GeoM.Rotate(gg.theta)
		op.GeoM.Translate(gg.pos.x, gg.pos.y)
		screen.DrawImage(img, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WIDTH, HEIGHT
}

func main() {
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
