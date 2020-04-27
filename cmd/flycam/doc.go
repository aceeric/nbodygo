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

Thanks to: https://github.com/JoeyDeVries/LearnOpenGL for the vector math supporting this functionality. According
to the attribution page there, no license / copyright presentation is required if content from that source is
used in the public domain.
*/
package flycam
