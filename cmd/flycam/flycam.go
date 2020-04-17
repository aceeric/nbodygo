package flycam

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"nbodygo/internal/pkg/camera"
	"nbodygo/internal/pkg/core"
	"nbodygo/internal/pkg/math32"
	"nbodygo/internal/pkg/window"
)

type CamValue float32

const (
	DefaultYaw           CamValue = -90.0
	DefaultPitch         CamValue = 0.0
	DefaultMovementSpeed CamValue = 1
	MouseSensitivity     CamValue = 0.1
)

type FlyCam struct {
	glfwWindow    *window.GlfwWindow
	cam           *camera.Camera
	position      math32.Vector3
	front         math32.Vector3
	up            math32.Vector3
	right         math32.Vector3
	lookAt        math32.Vector3
	worldUp       math32.Vector3
	captureInput  bool
	lastMouseX    float32
	lastMouseY    float32
	yaw           CamValue
	pitch         CamValue
	movementSpeed CamValue
}

const (
	True             = 1
	False            = 0
	FlyCamId         = 1234567
	EngageDisengage  = 2345678
	NoLastMousePos   = -1
	FrustrumFar      = 10000 // TODO consider 400000
)

// singleton
var flyCam FlyCam

func NewFlyCam(glfwWindow *window.GlfwWindow, scene *core.Node, width int, height int) *FlyCam {
	if flyCam.glfwWindow != nil {
		panic("Cannot call NewFlyCam twice")
	}
	if !glfw.RawMouseMotionSupported() {
		panic("FlyCam requires raw mouse motion which glfw says is not supported")
	}
	flyCam = FlyCam{
		glfwWindow,
		camera.New(1),
		*math32.NewVector3(10, 10, 100),
		*math32.NewVector3(0, 0, -1),
		*math32.NewVector3(0, 1, 0),
		math32.Vector3{},
		*math32.NewVector3(0, 0, 0),
		*math32.NewVector3(0, 1, 0),
		true,
		NoLastMousePos,NoLastMousePos,
		DefaultYaw,
		DefaultPitch,
		DefaultMovementSpeed,
	}
	flyCam.cam.SetPosition(flyCam.position.X, flyCam.position.Y, flyCam.position.Z)
	flyCam.cam.LookAt(flyCam.position.Clone().Add(&flyCam.front), &flyCam.up)
	flyCam.cam.SetAspect(float32(width) / float32(height))
	flyCam.cam.SetFar(FrustrumFar)
	scene.Add(flyCam.cam)
	engageCtrls(true)
	return &flyCam
}

func engageCtrls(all bool) {
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

// returns the lower level g3n camera wrapped by the fly cam
func (flyCam FlyCam) Cam() *camera.Camera {
	return flyCam.cam
}

// don't unsubscribe handleF12 - it's always active because even when the controls are disengaged we need
// the F12 handler to allow us to re-engage
func disengageCtrls() {
	glfwWindow := flyCam.glfwWindow
	glfwWindow.UnsubscribeID(window.OnKeyUp, FlyCamId)
	glfwWindow.UnsubscribeID(window.OnKeyRepeat, FlyCamId)
	glfwWindow.UnsubscribeID(window.OnKeyDown, FlyCamId)
	glfwWindow.UnsubscribeID(window.OnCursor, FlyCamId)
	w := glfwWindow.Window
	w.SetInputMode(glfw.RawMouseMotion, False)
	w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

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
	case window.KeyW: handleMvFwd()
	case window.KeyA: handleStrafeLft()
	case window.KeyS: handleMvBck()
	case window.KeyD: handleStrafeRt()
	case window.KeyQ: handleStrafeUp()
	case window.KeyZ: handleStrafeDn()
	case window.KeyKPAdd: handleKeypadPlus()
	case window.KeyKPSubtract: handleKeypadMinus()
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

// oddly, G3N engine *decreases* y position as the mouse goes *up*, so negate the value from G3N
// so that mouse up increases Y for consistency
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
	if flyCam.movementSpeed > 2 {flyCam.movementSpeed -= 1}
}

// engage/disengage the controls from the window
func handleF12(event string, ev interface{}) {
	var key window.Key
	switch kev := ev.(type) {
	case window.KeyEvent:  key = kev.Key
	case *window.KeyEvent: key = kev.Key
	default:               return
	}
	if key == window.KeyF12 {
		if flyCam.captureInput {
			disengageCtrls()
		} else {
			engageCtrls(false)
		}
		flyCam.captureInput = !flyCam.captureInput
	}
}

func handleEsc(event string, ev interface{}) {
	var key window.Key
	switch kev := ev.(type) {
	case window.KeyEvent:  key = kev.Key
	case *window.KeyEvent: key = kev.Key
	default:               return
	}
	if key == window.KeyEscape {
		flyCam.glfwWindow.SetShouldClose(true)
	}
}

func updateCamVectors() {
	front := math32.Vector3{}
	front.X = math32.Cos(math32.DegToRad(float32(flyCam.yaw))) * math32.Cos(math32.DegToRad(float32(flyCam.pitch)))
	front.Y = math32.Sin(math32.DegToRad(float32(flyCam.pitch)))
	front.Z = math32.Sin(math32.DegToRad(float32(flyCam.yaw))) * math32.Cos(math32.DegToRad(float32(flyCam.pitch)))
	flyCam.front = *front.Normalize()
	flyCam.right = *flyCam.front.Clone().Cross(&flyCam.worldUp).Normalize()
	flyCam.up = *flyCam.right.Clone().Cross(&flyCam.front).Normalize()
}
