/*
Provides a "Fly By" camera implementation on top of the G3N engine. Functionality:

 W       = Cam Forward (away from the viewer)
 A       = Strafe Left
 S       = Back (toward the viewer)
 D       = Strafe Right
 Q       = Strafe Up
 Z       = Strafe Down
 Mouse   = Look
 Keypad+ = Increase movement speed
 Keypad- = Decrease movement speed
 F12     = Unbind/bind keyboard/mouse from/to the sim window. Initially, controls are bound
 ESC     = Exit simulation

Thanks to: https://learnopengl.com/About for the vector math supporting mouse look functionality. Specifically:
https://github.com/JoeyDeVries/LearnOpenGL/blob/master/includes/learnopengl/camera.h

The referenced GitHub code is copyrighted by Joey de Vries (https://twitter.com/JoeyDeVriez) under
license: https://creativecommons.org/licenses/by-nc/4.0/
*/
package flycam
