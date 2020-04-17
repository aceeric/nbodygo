/*
Provides a "Fly By" camera implementation on top of the G3N engine. Functionality:
	W       = Cam Forward (away from the viewer)
	A       = Strafe Left
	S       = Back (toward the viewer)
	D       = Strafe Right
	Q       = Strafe Up
	Z       = Strafe Down
	Mouse   = Look
	keypad+ = Increase movement speed
	keypad- = Decrease movement speed
	F12     = Unbind/bind keyboard from/to the sim window. Initially, controls are bound
	ESC     = Exit simulation

Thanks to: https://learnopengl.com/code_viewer_gh.php?code=includes/learnopengl/camera.h for much of this
functionality.

TODO LICENSE https://creativecommons.org/licenses/by-nc/4.0/
*/
package flycam

