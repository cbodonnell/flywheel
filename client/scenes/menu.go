package scenes

import (
	"image/color"

	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type MenuScene struct {
	BaseScene

	ui *ebitenui.UI
}

type MenuSceneOptions struct {
	// OnLogin is called when the start game button is pressed.
	OnLogin func(username, password string)
}

func NewMenuScene(opts MenuSceneOptions) (Scene, error) {
	return &MenuScene{
		BaseScene{
			Root: objects.NewBaseObject("menu-root"),
		},
		initUI(InitUIOptions(opts)),
	}, nil
}

type InitUIOptions struct {
	// OnLogin is called when the start game button is pressed.
	OnLogin func(username, password string)
}

func initUI(opts InitUIOptions) *ebitenui.UI {
	buttonImage := &widget.ButtonImage{
		Idle:    image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255}),
		Hover:   image.NewNineSliceColor(color.NRGBA{R: 135, G: 135, B: 150, A: 255}),
		Pressed: image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255}),
	}

	fontFace := fonts.TTFNormalFont

	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    150,
				Left:   120,
				Right:  120,
				Bottom: 90,
			}))),
	)

	usernameTextInput := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
		widget.TextInputOpts.MobileInputMode("text"),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
			Disabled: image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
		}),
		widget.TextInputOpts.Face(fontFace),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(fontFace, 2),
		),
		widget.TextInputOpts.Placeholder("Username"),
	)
	rootContainer.AddChild(usernameTextInput)

	passwordTextInput := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
		widget.TextInputOpts.MobileInputMode("text"),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
			Disabled: image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
		}),
		widget.TextInputOpts.Face(fontFace),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(fontFace, 2),
		),
		widget.TextInputOpts.Placeholder("Password"),
		widget.TextInputOpts.Secure(true),
	)
	rootContainer.AddChild(passwordTextInput)

	button := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text("Login", fontFace, &widget.ButtonTextColor{
			Idle:     color.NRGBA{254, 255, 255, 255},
			Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),
	)
	rootContainer.AddChild(button)

	// auto focus the username text input
	usernameTextInput.Focus(true)

	// register login handler with relevant widget events
	loginHandler := func(args interface{}) {
		username, password := usernameTextInput.GetText(), passwordTextInput.GetText()
		if username == "" || password == "" {
			return
		}
		opts.OnLogin(username, password)
	}
	usernameTextInput.SubmitEvent.AddHandler(loginHandler)
	passwordTextInput.SubmitEvent.AddHandler(loginHandler)
	button.ClickedEvent.AddHandler(loginHandler)

	ui := &ebitenui.UI{
		Container: rootContainer,
	}

	return ui
}

func (s *MenuScene) Update() error {
	s.ui.Update()
	return nil
}

func (s *MenuScene) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}
