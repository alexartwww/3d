package main

import (
	"math"
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Point3D struct {
	X, Y, Z float64
}

var (
	width        = 800
	height       = 600
	perspective  = true
	e            = 300.0
	ox, oy, oz   float64
	scale        = 1.0
	lastX, lastY float64
	isDragging   bool
	sinTable     [360]float64
	cosTable     [360]float64
)

func init() {
	runtime.LockOSThread()

	for i := 0; i < 360; i++ {
		rad := float64(i) * math.Pi / 180
		sinTable[i] = math.Sin(rad)
		cosTable[i] = math.Cos(rad)
	}
}

func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(width, height, "3D Cube with Mouse Control", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Viewport(0, 0, int32(width), int32(height))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(width), float64(height), 0, -1, 1)
	gl.MatrixMode(gl.MODELVIEW)

	// Настройка обработчиков
	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(mouseMoveCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)
	window.SetScrollCallback(mouseScrollCallback)

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.LoadIdentity()
		gl.Translated(float64(width)/2, float64(height)/2, 0)
		gl.Scaled(scale, scale, scale)
		gl.Translated(-float64(width)/2, -float64(height)/2, 0)

		drawCube()
		drawAxes()

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func transform(pt Point3D) (x2d, y2d float64) {
	if perspective {
		thit := 1.0 / (1.0 - pt.X/e)
		x2d = float64(width)/2 + pt.Y*thit
		y2d = float64(height)/2 - pt.Z*thit
	} else {
		x2d = float64(width)/2 + pt.Y
		y2d = float64(height)/2 - pt.Z
	}
	return
}

func rotate(ax, ay, az float64, p *Point3D) {
	s1 := sinTable[int(ax)%360]
	c1 := cosTable[int(ax)%360]
	s2 := sinTable[int(ay)%360]
	c2 := cosTable[int(ay)%360]
	s3 := sinTable[int(az)%360]
	c3 := cosTable[int(az)%360]

	x := p.X
	y := p.Y
	z := p.Z

	p.X = x*(c2*c3) + y*(c2*s3) - z*s2
	p.Y = x*(s1*s2*c3-c1*s3) + y*(s1*s2*s3+c1*c3) + z*(s1*c2)
	p.Z = x*(c1*s2*c3+s1*s3) + y*(c1*s2*s3-s1*c3) + z*(c1*c2)
}

func drawLine3D(p1, p2 Point3D, color []float32) {
	x1, y1 := transform(p1)
	x2, y2 := transform(p2)

	gl.Color3fv(&color[0])
	gl.Begin(gl.LINES)
	gl.Vertex2d(x1, y1)
	gl.Vertex2d(x2, y2)
	gl.End()
}

func drawCube() {
	size := 100.0
	cubePoints := []Point3D{
		{-size, -size, size}, {size, -size, size},
		{size, -size, size}, {size, size, size},
		{size, size, size}, {-size, size, size},
		{-size, size, size}, {-size, -size, size},

		{-size, -size, -size}, {size, -size, -size},
		{size, -size, -size}, {size, size, -size},
		{size, size, -size}, {-size, size, -size},
		{-size, size, -size}, {-size, -size, -size},

		{size, size, -size}, {size, size, size},
		{-size, size, -size}, {-size, size, size},
		{size, -size, -size}, {size, -size, size},
		{-size, -size, -size}, {-size, -size, size},
	}

	for i := 0; i < len(cubePoints); i += 2 {
		p1 := cubePoints[i]
		p2 := cubePoints[i+1]

		rotate(ox, oy, oz, &p1)
		rotate(ox, oy, oz, &p2)

		drawLine3D(p1, p2, []float32{1.0, 1.0, 1.0}) // Белый цвет для куба
	}
}

func drawAxes() {
	axes := []struct {
		p1, p2 Point3D
		color  []float32
	}{
		{Point3D{150, 0, 0}, Point3D{0, 0, 0}, []float32{1.0, 1.0, 0.0}}, // OX (желтый)
		{Point3D{0, 150, 0}, Point3D{0, 0, 0}, []float32{1.0, 1.0, 0.0}}, // OY (желтый)
		{Point3D{0, 0, 150}, Point3D{0, 0, 0}, []float32{1.0, 1.0, 0.0}}, // OZ (желтый)
	}

	for _, axis := range axes {
		p1 := axis.p1
		p2 := axis.p2

		rotate(ox, oy, oz, &p1)
		rotate(ox, oy, oz, &p2)

		drawLine3D(p1, p2, axis.color)
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && key == glfw.KeyEscape {
		window.SetShouldClose(true)
	}
}

func mouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft {
		isDragging = (action == glfw.Press)
		lastX, lastY = window.GetCursorPos() // Правильный способ получения позиции
	}
}

func mouseMoveCallback(window *glfw.Window, xpos, ypos float64) {
	if isDragging {
		dx := xpos - lastX
		dy := ypos - lastY

		// Чувствительность вращения
		oz -= dx * 0.5
		oy -= dy * 0.5

		lastX, lastY = xpos, ypos

		// Нормализация углов
		ox = math.Mod(ox, 360)
		oy = math.Mod(oy, 360)
		if ox < 0 {
			ox += 360
		}
		if oy < 0 {
			oy += 360
		}
		if oz < 0 {
			oz += 360
		}
		//if ox > 359 {
		//	ox -= 359
		//}
		//if oy > 359 {
		//	oy -= 359
		//}
		//if oz > 359 {
		//	oz -= 359
		//}
	}
}

func mouseScrollCallback(window *glfw.Window, xoff, yoff float64) {
	// Масштабирование колесиком мыши
	scale += yoff * 0.1
	if scale < 0.1 {
		scale = 0.1
	}
	if scale > 3.0 {
		scale = 3.0
	}
}
