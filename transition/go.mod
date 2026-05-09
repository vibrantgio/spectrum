module github.com/vibrantgio/spectrum/transition

go 1.25.1

require (
	github.com/vibrantgio/prism/tokens v0.0.0
	github.com/vibrantgio/pulse/tween v0.0.0
)

replace (
	github.com/vibrantgio/prism/tokens => ../../prism/tokens
	github.com/vibrantgio/pulse/tween => ../../pulse/tween
)
