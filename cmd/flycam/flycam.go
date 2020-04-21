package flycam

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"nbodygo/internal/pkg/camera"
	"nbodygo/internal/pkg/core"
	"nbodygo/internal/pkg/math32"
	"nbodygo/internal/pkg/window"
)

type CamValue float32

//
// fly cam state
//
type FlyCam struct {
	// pointer to the glfw window managed by g3n
	glfwWindow *window.GlfwWindow
	// pointer to the camera managed by g3n
	cam *camera.Camera
	// current position and directional vectors
	position math32.Vector3
	front    math32.Vector3
	up       math32.Vector3
	right    math32.Vector3
	lookAt   math32.Vector3
	worldUp  math32.Vector3
	// true if keyboard/mouse bound to the sim window, else false
	captureInput bool
	// used to interpret mouse event stream
	lastMouseX float32
	lastMouseY float32
	// continuously updated in response to mouse/keyboard events
	yaw           CamValue
	pitch         CamValue
	movementSpeed CamValue
}

const (
	True                          = 1
	False                         = 0
	FlyCamId                      = 1234567
	EngageDisengage               = 2345678
	NoLastMousePos                = -1
	FrustrumFar                   = 10000 // TODO consider 400000
	DefaultYaw           CamValue = -90.0
	DefaultPitch         CamValue = 0.0
	DefaultMovementSpeed CamValue = 1
	MouseSensitivity     CamValue = 0.1
)

// singleton
var flyCam FlyCam

//
// Creates the fly cam
//
// args:
//   glfwWindow      managed by g3n
//   scene           managed by g3n
//   width,height    screen dimensions
//   initialPosition initial position of the camera
//
// returns:
//   pointer to fly cam
//
func NewFlyCam(glfwWindow *window.GlfwWindow, scene *core.Node, width int, height int, initialPosition math32.Vector3) *FlyCam {
	if flyCam.glfwWindow != nil {
		panic("Cannot call NewFlyCam twice")
	}
	if !glfw.RawMouseMotionSupported() {
		panic("FlyCam requires raw mouse motion which glfw says is not supported")
	}
	flyCam = FlyCam{
		glfwWindow,
		camera.New(1),
		initialPosition,
		*math32.NewVector3(0, 0, -1),
		*math32.NewVector3(0, 1, 0),
		math32.Vector3{},
		*math32.NewVector3(0, 0, 0),
		*math32.NewVector3(0, 1, 0),
		true,
		NoLastMousePos, NoLastMousePos,
		DefaultYaw,
		DefaultPitch,
		DefaultMovementSpeed,
	}
	flyCam.cam.SetPosition(flyCam.position.X, flyCam.position.Y, flyCam.position.Z)
	flyCam.cam.LookAt(flyCam.position.Clone().Add(&flyCam.front), &flyCam.up)
	flyCam.cam.SetAspect(float32(width) / float32(height))
	flyCam.cam.SetFar(FrustrumFar)
	scene.Add(flyCam.cam)
	engage(true)
	return &flyCam
}

//
// Attaches mouse/keyboard to the sim window
//
// args:
//   all True if F12 should be engaged (only pass true the first time)
//
func engage(all bool) {
	glfwWindow := flyCam.glfwWindow
	glfwWindow.SubscribeID(window.OnKeyUp, FlyCamId, handleEsc)
	if all {
		glfwWindow.SubscribeID(window.OnKeyUp, EngageDisengage, handleF12)
	}
	glfwWindow.SubscribeID(window.OnKeyRepeat, FlyCamId, handleKey)
	glfwWindow.SubscribeID(window.OnKeyDown, FlyCamId, handleKey)
	glfwWindow.SubscribeID(window.OnCursor, FlyCamId, handleMouseLook)
	w := glfwWindow.Window
	w.SetInputMode(glfw.RawMouseMotion, True)
	// allow the scroll to change indefinitely and wrap
	w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
}

//
// returns the lower level g3n camera wrapped by the fly cam
//
func (flyCam FlyCam) Cam() *camera.Camera {
	return flyCam.cam
}

//
// Disengages the mouse/keyboard from the sim window. Doesn't unsubscribe the F12 handler- that
// one is always active because even when the controls are disengaged we need
// the F12 handler to allow us to re-engage
//
func disengage() {
	glfwWindow := flyCam.glfwWindow
	glfwWindow.UnsubscribeID(window.OnKeyUp, FlyCamId)
	glfwWindow.UnsubscribeID(window.OnKeyRepeat, FlyCamId)
	glfwWindow.UnsubscribeID(window.OnKeyDown, FlyCamId)
	glfwWindow.UnsubscribeID(window.OnCursor, FlyCamId)
	w := glfwWindow.Window
	w.SetInputMode(glfw.RawMouseMotion, False)
	w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

//
// Keyboard event handler
//
func handleKey(event string, ev interface{}) {
	var key window.Key
	switch kev := ev.(type) {
	case window.KeyEvent:
	case *window.KeyEvent:
		key = kev.Key
	default:
		return
	}
	switch key {
	case window.KeyW:
		handleMvFwd()
	case window.KeyA:
		handleStrafeLft()
	case window.KeyS:
		handleMvBck()
	case window.KeyD:
		handleStrafeRt()
	case window.KeyQ:
		handleStrafeUp()
	case window.KeyZ:
		handleStrafeDn()
	case window.KeyKPAdd:
		handleKeypadPlus()
	case window.KeyKPSubtract:
		handleKeypadMinus()
	}
}

func updateView() {
	flyCam.cam.SetPositionVec(&flyCam.position)
	flyCam.cam.LookAt(flyCam.position.Clone().Add(&flyCam.front), &flyCam.up)
}

func handleMvFwd() {
	flyCam.position.Add(flyCam.front.Clone().MultiplyScalar(float32(flyCam.movementSpeed)))
	updateView()
}

func handleStrafeLft() {
	flyCam.position.Sub(flyCam.right.Clone().MultiplyScalar(float32(flyCam.movementSpeed)))
	updateView()
}

func handleMvBck() {
	flyCam.position.Sub(flyCam.front.Clone().MultiplyScalar(float32(flyCam.movementSpeed)))
	updateView()
}

func handleStrafeRt() {
	flyCam.position.Add(flyCam.right.Clone().MultiplyScalar(float32(flyCam.movementSpeed)))
	updateView()
}

func handleStrafeUp() {
	flyCam.position.Add(flyCam.up.Clone().MultiplyScalar(float32(flyCam.movementSpeed)))
	updateView()
}

func handleStrafeDn() {
	flyCam.position.Sub(flyCam.up.Clone().MultiplyScalar(float32(flyCam.movementSpeed)))
	updateView()
}

//
// Mouse event handler. Oddly, G3N engine *decreases* y position as the mouse goes *up*,
// so negate the value from G3N so that mouse up increases Y for consistency (X increases
// right, so...)
//
func handleMouseLook(event string, ev interface{}) {
	var xPos, yPos float32
	switch cev := ev.(type) {
	case window.CursorEvent:
		xPos = cev.Xpos
		yPos = -cev.Ypos
	case *window.CursorEvent:
		xPos = cev.Xpos
		yPos = -cev.Ypos
	default:
		return
	}
	if flyCam.lastMouseX != NoLastMousePos {
		deltaX := CamValue(xPos - flyCam.lastMouseX)
		deltaY := CamValue(yPos - flyCam.lastMouseY)
		flyCam.yaw += deltaX * MouseSensitivity
		flyCam.pitch += deltaY * MouseSensitivity
		updateCamVectors()
		flyCam.cam.LookAt(flyCam.position.Clone().Add(&flyCam.front), &flyCam.up)
	}
	flyCam.lastMouseX = xPos
	flyCam.lastMouseY = yPos
}

func handleKeypadPlus() {
	flyCam.movementSpeed += 1
}

func handleKeypadMinus() {
	if flyCam.movementSpeed > 2 {
		flyCam.movementSpeed -= 1
	}
}

//
// Engages/Disengages the controls from the window
//
func handleF12(event string, ev interface{}) {
	var key window.Key
	switch kev := ev.(type) {
	case window.KeyEvent:
		key = kev.Key
	case *window.KeyEvent:
		key = kev.Key
	default:
		return
	}
	if key == window.KeyF12 {
		if flyCam.captureInput {
			disengage()
		} else {
			engage(false)
		}
		flyCam.captureInput = !flyCam.captureInput
	}
}

//
// Communicates to the G3N engine to exit
//
func handleEsc(event string, ev interface{}) {
	var key window.Key
	switch kev := ev.(type) {
	case window.KeyEvent:
		key = kev.Key
	case *window.KeyEvent:
		key = kev.Key
	default:
		return
	}
	if key == window.KeyEscape {
		flyCam.glfwWindow.SetShouldClose(true)
	}
}

//
// Implements the fly camera. This function modeled after code
// from https://github.com/JoeyDeVries/LearnOpenGL
//
func updateCamVectors() {
	front := math32.Vector3{}
	front.X = math32.Cos(math32.DegToRad(float32(flyCam.yaw))) * math32.Cos(math32.DegToRad(float32(flyCam.pitch)))
	front.Y = math32.Sin(math32.DegToRad(float32(flyCam.pitch)))
	front.Z = math32.Sin(math32.DegToRad(float32(flyCam.yaw))) * math32.Cos(math32.DegToRad(float32(flyCam.pitch)))
	flyCam.front = *front.Normalize()
	flyCam.right = *flyCam.front.Clone().Cross(&flyCam.worldUp).Normalize()
	flyCam.up = *flyCam.right.Clone().Cross(&flyCam.front).Normalize()
}
