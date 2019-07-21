package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	glutils "github.com/henrixapp/go-OGLL/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Instruction storing information
type Instruction struct {
	//storage for movement informations
	x1 float64
	//storage for angle information
	x2 float64
	// color information
	r, g, b, a float32
	// variables storing misc operations
	movement    bool
	rotate      bool
	stackpush   bool
	stackpop    bool
	changeColor bool
	colorClear  bool
}

// rune is a sign to be assigned

// CoordAndColor storing step
type CoordAndColor struct {
	x          float64
	y          float64
	r, g, b, a float32
}

//global vars

var symbolMap map[rune]Instruction
var replacementMap map[rune][]rune

//Function that executes the rules
func executeReplacement(term []rune, times int) []rune {
	if times <= 0 {
		return term
	}
	var result = make([]rune, 0)
	//left sided deriviation
	for _, v := range term {
		if replacementMap[v] == nil {
			result = append(result, v)
		} else {
			result = append(result, executeReplacement(replacementMap[v], times-1)...)
		}
	}

	return result
}

// Strip containing multiple positions
type Strip []CoordAndColor

func stripsToStars(strip Strip) {
	elements := len(strip)
	if elements > 0 {
		data := make([]float64, elements*2)
		color := make([]float32, elements*4)
		for i, v := range strip {
			data[i*2] = v.x
			data[i*2+1] = v.y
			color[i*4] = v.r
			color[i*4+1] = v.g
			color[i*4+2] = v.b
			color[i*4+3] = v.a
		}

		var vbo uint32
		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, elements*2*8, gl.Ptr(data), gl.STATIC_DRAW)
		//colors
		var colorVbo uint32
		gl.GenBuffers(1, &colorVbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, colorVbo)
		gl.BufferData(gl.ARRAY_BUFFER, elements*4*4, gl.Ptr(color), gl.STATIC_DRAW)
		var vao uint32
		gl.GenVertexArrays(1, &vao)
		gl.BindVertexArray(vao)
		gl.EnableVertexAttribArray(0)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.VertexAttribPointer(0, 2, gl.DOUBLE, false, 0, gl.PtrOffset(0))
		gl.BindBuffer(gl.ARRAY_BUFFER, colorVbo)
		gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(1)
		stripes = append(stripes, vao)
		stripesSizes = append(stripesSizes, int32(elements))
	}
}
func executeInstructions(term []rune) {

	var ZurZeit = CoordAndColor{0, 0, 0, 0, 0, 1.}
	angle := 0.0
	var Stack = make([]CoordAndColor, 0)
	var StackAngle = make([]float64, 0)

	Strips := make([]Strip, 0)
	Aktueller := make(Strip, 0)
	Aktueller = append(Aktueller, ZurZeit)
	for _, symbol := range term {
		a := symbolMap[symbol]
		// Priorization
		//1. Stack push, stack pop
		//2. rotate
		//3. movement
		if a.stackpush {
			Stack = append(Stack, ZurZeit)
			StackAngle = append(StackAngle, angle)
		}
		if a.stackpop {
			if len(Stack) == 0 {
				fmt.Println("Stack exception. Tried to pop instead of push.")
			} else {
				ZurZeit = Stack[len(Stack)-1]
				angle = StackAngle[len(Stack)-1]
				Stack = Stack[:len(Stack)-2]
				StackAngle = StackAngle[:len(StackAngle)-2]
				Strips = append(Strips, Aktueller)
				Aktueller = make(Strip, 0)
			}
		}
		if a.colorClear {
			ZurZeit.r = 0.0
			ZurZeit.g = 0.0
			ZurZeit.b = 0.0
			ZurZeit.a = 1.0
		}
		if a.changeColor {
			ZurZeit.r = float32(math.Mod(float64(ZurZeit.r+a.r), 1.000001))
			ZurZeit.g = float32(math.Mod(float64(ZurZeit.g+a.g), 1.000001))
			ZurZeit.b = float32(math.Mod(float64(ZurZeit.b+a.b), 1.000001))
			ZurZeit.a = float32(math.Mod(float64(ZurZeit.a+a.a), 1.000001))
		}
		if a.rotate {
			angle += a.x2
		}
		if a.movement {
			ZurZeit.x += a.x1 * math.Cos(angle*math.Pi/180.0)
			ZurZeit.y += a.x1 * math.Sin(angle*math.Pi/180.0)
			Aktueller = append(Aktueller, ZurZeit)
		}
	}
	Strips = append(Strips, Aktueller)
	var size = 0
	for _, strip := range Strips {
		size += len(strip)
		stripsToStars(strip)
	}
}

var replaced []rune

// Parser
func parse(command string, render chan bool) {
	var symbol rune
	if strings.Contains(command, "=") {
		// Parse for property
		operands := strings.Split(command, "=")
		symbol = []rune(operands[0])[0]
		inst := symbolMap[symbol]
		properties := strings.Split(operands[1], ",")
		for _, prop := range properties {
			var value = ""
			var property = prop
			if strings.Contains(prop, "(") {
				property = strings.Split(prop, "(")[0]
				value = strings.Split(strings.Split(prop, "(")[1], ")")[0]
			}
			fval, _ := strconv.ParseFloat(value, 32)
			switch property {
			case "pop":
				inst.stackpop = true
			case "push":
				inst.stackpush = true
			case "mov":
				inst.movement = true
				inst.x1, _ = strconv.ParseFloat(value, 64)
			case "rot":
				inst.rotate = true
				inst.x2, _ = strconv.ParseFloat(value, 64)
			case "red":
				inst.changeColor = true
				inst.r = float32(fval)
			case "green":
				inst.changeColor = true
				inst.g = float32(fval)
			case "blue":
				inst.changeColor = true
				inst.b = float32(fval)
			case "alpha":
				inst.changeColor = true
				inst.a = float32(fval)
			case "clear":
				inst.colorClear = true
			}
		}
		symbolMap[symbol] = inst
	} else if strings.Contains(command, "render(") {
		options := strings.Split(strings.Split(strings.Split(command, "(")[1], ")")[0], ",")
		recDepth, _ := strconv.ParseInt(options[1], 10, 32)
		replaced = executeReplacement([]rune(options[0]), int(recDepth))
		render <- true

		rendered := <-render
		if rendered {
			fmt.Println("Sucessfully rendered.")
		}
	} else {
		// Parse for replacement rule.
		symbol = []rune(command)[0]
		replacementMap[symbol] = []rune(command[3:])
	}
}

//input routine
func inputRoutine(render chan bool) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Simple OGLL Shell")
	fmt.Println("---------------------")
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		parse(text, render)
	}

}

// Sleeps for given waitingTime and returns than via channel
func pollRoutine(c chan bool, waitingTime int) {
	time.Sleep(time.Duration(waitingTime) * time.Millisecond)
	c <- true
}

var vertexShader = `#version 400
layout(location=0) in vec2 vp;
layout(location=1) in vec4 vc;
 uniform float YStretch;
 uniform float scale;
 uniform vec2 verschub;
 out vec4 coloer;
void main () {
  gl_Position = vec4 (scale*vec2((vp+verschub).x,(vp+verschub).y*YStretch), 0.0,1.0);
  coloer=vc;
}`

var fragmentShader = `#version 400
in vec4 coloer;
out vec4 frag_colour;
void main () {
  frag_colour = coloer;
}`

var glslScale, glslVerschub, glslYStretch int32
var scale, yStretch float32 = 1, 0
var verschub [2]float32

var window *glfw.Window
var program uint32

//setup function
func setupGL() {
	runtime.LockOSThread()
	var err error
	program, err = glutils.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)
	glslScale = gl.GetUniformLocation(program, gl.Str("scale\x00"))
	glslVerschub = gl.GetUniformLocation(program, gl.Str("verschub\x00"))
	glslYStretch = gl.GetUniformLocation(program, gl.Str("YStretch\x00"))
	updateFactors()
	gl.BindFragDataLocation(program, 0, gl.Str("frag_colour\x00"))
	runtime.UnlockOSThread()
}
func handleInput() {
	if glfw.Press == window.GetKey(glfw.KeyEscape) {
		window.SetShouldClose(true)
	}
	//scaling
	if glfw.Press == window.GetKey(glfw.KeyQ) {
		scale -= scale * 0.1
	}
	if glfw.Press == window.GetKey(glfw.KeyE) {
		scale += scale * 0.1
	}
	if glfw.Press == window.GetKey(glfw.KeyA) {
		verschub[0] += 0.01 / scale
	}
	if glfw.Press == window.GetKey(glfw.KeyD) {
		verschub[0] -= 0.01 / scale
	}
	if glfw.Press == window.GetKey(glfw.KeyW) {
		verschub[1] -= 0.01 / scale
	}
	if glfw.Press == window.GetKey(glfw.KeyS) {
		verschub[1] += 0.01 / scale
	}
	w, h := window.GetSize()
	yStretch = float32(w) / float32(h)
}
func updateFactors() {
	gl.Uniform1f(glslScale, scale)
	gl.Uniform1f(glslYStretch, yStretch)
	gl.Uniform2fv(glslVerschub, 1, &verschub[0])
}

var stripes []uint32
var stripesSizes []int32

//drawing function
func redraw() {
	//reset view
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	updateFactors()
	// draw
	for i, v := range stripes {
		gl.BindVertexArray(v)
		// Drawing
		gl.DrawArrays(gl.LINE_STRIP, 0, stripesSizes[i])
	}
}
func main() {
	//initialize global variables
	symbolMap = make(map[rune]Instruction)
	replacementMap = make(map[rune][]rune)
	var ch = make(chan bool)
	//setup Gl
	runtime.LockOSThread()
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 16)
	window, err = glfw.CreateWindow(640, 480, "OGLL", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	//init glow
	if err := gl.Init(); err != nil {
		panic(err)
	}
	setupGL()
	go inputRoutine(ch)
	var rendering = false
	waitingChan := make(chan bool)
	for !window.ShouldClose() {
		go pollRoutine(waitingChan, 50)
		select {
		case rendering = <-ch:
			if rendering {
				executeInstructions(replaced)
				//finished rendering.
				ch <- true
			}
		case <-waitingChan:
			handleInput()
			updateFactors()
			redraw()
			window.SwapBuffers()
			glfw.PollEvents()

		}
	}
}
