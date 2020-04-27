package flycam

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"nbodygo/internal/pkg/camera"
	"nbodygo/internal/pkg/core"
	"nbodygo/internal/pkg/math32"
	"nbodygo/internal/pkg/window"
)

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
	// true if f12 has been bound, else false (by default)
	f12Bound bool
	// used to interpret mouse event stream
	lastMouseX float32
	lastMouseY float32
	// continuously updated in response to mouse/keyboard events
	yaw           float32
	pitch         float32
	movementSpeed float32
}

const (
	True                 = 1
	False                = 0
	DefaultEvId          = 1234567
	EngageDisengageEvId  = 2345678
	NoLastMousePos       = -1
	FrustrumFar          = 10000 // TODO consider 400000 as in the Java version
	DefaultYaw           = -90.0
	DefaultPitch         = 0.0
	DefaultMovementSpeed = 1
	MouseSensitivity     = 0.1
)

// singleton
var flyCam FlyCam

//
// Creates the fly cam singleton
//
// args:
//   glfwWindow      managed by g3n
//   scene           managed by g3n
//   width,height    sim window dimensions
//   initialPosition initial position of the camera
//
// returns:
//   pointer to fly cam
//
func NewFlyCam(glfwWindow *window.GlfwWindow, scene *core.Node, width, height int, initialPosition math32.Vector3) *FlyCam {
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
		false,
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
	engage()
	return &flyCam
}

//
// Attaches mouse/keyboard to the sim window. Only attaches F12 once
//
func engage() {
	glfwWindow := flyCam.glfwWindow
	if !flyCam.f12Bound {
		flyCam.f12Bound = true
		glfwWindow.SubscribeID(window.OnKeyUp, EngageDisengageEvId, handleF12)
	} else {
		flyCam.lastMouseX, flyCam.lastMouseY = NoLastMousePos, NoLastMousePos
	}
	glfwWindow.SubscribeID(window.OnKeyUp, DefaultEvId, handleEsc)
	glfwWindow.SubscribeID(window.OnKeyRepeat, DefaultEvId, handleKey)
	glfwWindow.SubscribeID(window.OnKeyDown, DefaultEvId, handleKey)
	glfwWindow.SubscribeID(window.OnCursor, DefaultEvId, handleMouseLook)
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
	glfwWindow.UnsubscribeID(window.OnKeyUp, DefaultEvId)
	glfwWindow.UnsubscribeID(window.OnKeyRepeat, DefaultEvId)
	glfwWindow.UnsubscribeID(window.OnKeyDown, DefaultEvId)
	glfwWindow.UnsubscribeID(window.OnCursor, DefaultEvId)
	w := glfwWindow.Window
	w.SetInputMode(glfw.RawMouseMotion, False)
	w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

//
// Keyboard event handler
//
func handleKey(_ string, ev interface{}) {
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
		flyCam.position.Add(flyCam.front.Clone().MultiplyScalar(flyCam.movementSpeed))
	case window.KeyA:
		flyCam.position.Sub(flyCam.right.Clone().MultiplyScalar(flyCam.movementSpeed))
	case window.KeyS:
		flyCam.position.Sub(flyCam.front.Clone().MultiplyScalar(flyCam.movementSpeed))
	case window.KeyD:
		flyCam.position.Add(flyCam.right.Clone().MultiplyScalar(flyCam.movementSpeed))
	case window.KeyQ:
		flyCam.position.Add(flyCam.up.Clone().MultiplyScalar(flyCam.movementSpeed))
	case window.KeyZ:
		flyCam.position.Sub(flyCam.up.Clone().MultiplyScalar(flyCam.movementSpeed))
	case window.KeyKPAdd:
		flyCam.movementSpeed += 1
		return
	case window.KeyKPSubtract:
		if flyCam.movementSpeed > 2 {
			flyCam.movementSpeed -= 1
		}
		return
	default:
		return
	}
	flyCam.cam.SetPositionVec(&flyCam.position)
	flyCam.cam.LookAt(flyCam.position.Clone().Add(&flyCam.front), &flyCam.up)
}

//
// Mouse event handler. Oddly, G3N engine *decreases* y position as the mouse goes *up*,
// so negate the value from G3N so that mouse up increases Y for consistency (X increases
// right, so...)
//
func handleMouseLook(_ string, ev interface{}) {
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
		deltaX := xPos - flyCam.lastMouseX
		deltaY := yPos - flyCam.lastMouseY
		flyCam.yaw += deltaX * MouseSensitivity
		flyCam.pitch += deltaY * MouseSensitivity
		updateCamVectors()
		flyCam.cam.LookAt(flyCam.position.Clone().Add(&flyCam.front), &flyCam.up)
	}
	flyCam.lastMouseX = xPos
	flyCam.lastMouseY = yPos
}

//
// Updates the fly camera vectors based on mouse look
//
func updateCamVectors() {
	front := math32.Vector3{}
	front.X = math32.Cos(math32.DegToRad(flyCam.yaw)) * math32.Cos(math32.DegToRad(flyCam.pitch))
	front.Y = math32.Sin(math32.DegToRad(flyCam.pitch))
	front.Z = math32.Sin(math32.DegToRad(flyCam.yaw)) * math32.Cos(math32.DegToRad(flyCam.pitch))
	flyCam.front = *front.Normalize()
	flyCam.right = *flyCam.front.Clone().Cross(&flyCam.worldUp).Normalize()
	flyCam.up = *flyCam.right.Clone().Cross(&flyCam.front).Normalize()
}

//
// Engages/Disengages the controls from the window
//
func handleF12(_ string, ev interface{}) {
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
			engage()
		}
		flyCam.captureInput = !flyCam.captureInput
	}
}

//
// Communicates to the G3N engine to exit
//
func handleEsc(_ string, ev interface{}) {
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
