package main

import (
	"fmt"
	"runtime"
	"strings"

	gl "github.com/go-gl/gl/v3.1/gles2"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

var vshader2 = `
  // an attribute will receive data from a buffer
  attribute vec4 a_position;

  // all shaders have a main function
  void main() {
    // gl_Position is a special variable a vertex shader
    // is responsible for setting
    gl_Position = a_position;
  }
`
var fshader2 = `
  void main() {
    // gl_FragColor is a special variable a fragment shader
    // is responsible for setting
    gl_FragColor = vec4(0, 1, 0, 1); // return redish-purple
  }
`

var vshader3 = `#version 300 es
precision highp float;

in vec4 a_position;
out vec4 v_color;

// all shaders have a main function
void main() {
	// gl_Position is a special variable a vertex shader
	// is responsible for setting
	gl_Position = a_position;
	v_color = gl_Position * 0.5 + 0.5;
}
`
var fshader3 = `#version 300 es
precision highp float;

out vec4 outColor;

in vec4 v_color;

void main() {
  outColor = v_color;
}
`

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// glfw.WindowHint(glfw.Resizable, glfw.False)
	// glfw.WindowHint(glfw.ContextVersionMajor, 4)
	// glfw.WindowHint(glfw.ContextVersionMinor, 1)
	// glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	// glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// gles3 := false
	gles3 := true

	glfw.WindowHint(glfw.ClientAPI, glfw.OpenGLESAPI)
	if gles3 {
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 0)
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 0)
	}

	W := 640
	H := 480
	window, err := glfw.CreateWindow(W, H, "gles2", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	fmt.Println("OpenGL vendor", gl.GoStr(gl.GetString(gl.VENDOR)))
	fmt.Println("OpenGL version", gl.GoStr(gl.GetString(gl.VERSION)))
	fmt.Println("OpenGL renderer", gl.GoStr(gl.GetString(gl.RENDERER)))
	fmt.Println("OpenGL shading language version", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
	// fmt.Println("OpenGL ext", gl.GoStr(gl.GetString(gl.EXTENSIONS)))

	gl.Viewport(0, 0, int32(W), int32(H))
	// gl.Viewport(0, 0, 200, 200)
	// gl.Viewport(0, 0, 1000, 1000)

	// Configure the vertex and fragment shaders
	vshader := vshader2
	fshader := fshader2
	if gles3 {
		vshader = vshader3
		fshader = fshader3
	}
	program, err := newProgram(vshader, fshader)
	if err != nil {
		panic(err)
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	fmt.Printf("vbo=%v\n", vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(triVertices)*4, gl.Ptr(triVertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("a_position\x00")))
	fmt.Printf("vertAttrib=%v\n", vertAttrib)
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	// gl.ClearColor(.2, 0., 0., 0.5)
	gl.ClearColor(0., 0., 0.2, 1.)

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(program)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		// Do OpenGL stuff.
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

var triVertices = []float32{
	-1, -1,
	1, 0,
	0, 1,
}
