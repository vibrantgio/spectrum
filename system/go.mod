module github.com/vibrantgio/spectrum/system

go 1.25.1

require (
	github.com/reactivego/rx v0.2.2
	github.com/vibrantgio/prism/theme v0.0.0
	github.com/vibrantgio/prism/tokens v0.0.0
)

require github.com/reactivego/scheduler v0.1.2 // indirect

replace github.com/vibrantgio/prism/theme => ../../prism/theme

replace github.com/vibrantgio/prism/tokens => ../../prism/tokens
