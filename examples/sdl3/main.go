package main

import (
	"log/slog"
	"runtime/debug"
	"unsafe"

	"github.com/totallygamerjet/clay"
	"github.com/totallygamerjet/clay/examples/fonts"
	sl "github.com/totallygamerjet/clay/examples/shared-layouts"
	sdl3 "github.com/totallygamerjet/clay/renderers/sdl3"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

func handleClayError(errorText clay.ErrorData) {
	slog.Error(errorText.ErrorText.String(), "stacktrace", debug.Stack())
}

// TODO: CreateArenaWithCapacityAndMemory should take a slice of bytes

func main() {
	const (
		winWidth, winHeight = 640, 480
	)

	defer binsdl.Load().Unload()
	defer binttf.Load().Unload()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	if err := ttf.Init(); err != nil {
		panic(err)
	}

	var (
		window   *sdl.Window
		renderer *sdl.Renderer
		err      error
	)

	window, renderer, err = sdl.CreateWindowAndRenderer("SDL", winWidth, winHeight, sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	if err := window.SetResizable(true); err != nil {
		panic(err)
	}

	textEngine, err := ttf.CreateRendererTextEngine(renderer)
	if err != nil {
		panic(err)
	}

	stream, err := sdl.IOFromConstMem(fonts.RobotoRegularTTF)
	if err != nil {
		panic(err)
	}

	font, err := ttf.OpenFontIO(stream, false, 16)
	if err != nil {
		panic(err)
	}

	rendererData := &sdl3.RendererData{
		Renderer:   renderer,
		TextEngine: textEngine,
		Fonts: []*ttf.Font{
			font,
		},
	}

	// Initialize Clay
	totalMemorySize := clay.MinMemorySize()
	memory := make([]byte, totalMemorySize)
	arena := clay.CreateArenaWithCapacityAndMemory(uint64(totalMemorySize), unsafe.Pointer(unsafe.SliceData(memory)))
	clay.Initialize(arena, clay.Dimensions{Width: winWidth, Height: winHeight}, clay.ErrorHandler{ErrorHandlerFunction: handleClayError})
	clay.SetMeasureTextFunction(sdl3.MeasureText, unsafe.Pointer(&rendererData.Fonts))

	var demoData = sl.ClayVideoDemo_Initialize()

	sdl.RunLoop(func() error {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				return sdl.EndLoop
			case sdl.EVENT_WINDOW_RESIZED:
				e := event.WindowEvent()
				clay.SetLayoutDimensions(clay.Dimensions{
					Width:  float32(e.Data1),
					Height: float32(e.Data2),
				})
			case sdl.EVENT_MOUSE_WHEEL:
				e := event.MouseWheelEvent()
				scrollDelta := clay.Vector2{
					X: e.X,
					Y: e.Y,
				}
				clay.UpdateScrollContainers(true, scrollDelta, 0.01)
			}
		}
		state, x, y := sdl.GetMouseState()
		clay.SetPointerState(clay.Vector2{
			X: float32(x),
			Y: float32(y),
		}, state&sdl.BUTTON_LEFT != 0)

		renderCommands := sl.ClayVideoDemo_CreateLayout(&demoData)

		_ = renderer.SetDrawColor(0, 0, 0, 255)
		_ = renderer.Clear()

		sdl3.ClayRender(rendererData, renderCommands)

		renderer.Present()

		return nil
	})
}
