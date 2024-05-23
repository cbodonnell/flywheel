package scenes

import (
	"image/color"

	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/cbodonnell/flywheel/client/ui"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type MenuScene struct {
	*BaseScene

	onLogin  func(email, password string) error
	ui       *ebitenui.UI
	email    string
	password string
	loginErr string
}

type MenuSceneOptions struct {
	// OnLogin is called when the start game button is pressed.
	OnLogin func(email, password string) error
}

var _ Scene = &MenuScene{}

func NewMenuScene(opts MenuSceneOptions) (Scene, error) {
	return &MenuScene{
		BaseScene: NewBaseScene(objects.NewBaseObject("menu-root", nil)),
		onLogin:   opts.OnLogin,
	}, nil
}

func (s *MenuScene) Init() error {
	s.renderUI()
	return s.BaseScene.Init()
}

func (s *MenuScene) renderUI() {
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

	emailTextInput := widget.NewTextInput(
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
		widget.TextInputOpts.Placeholder("Email"),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			s.email = args.InputText
		}),
	)
	emailTextInput.SetText(s.email)
	rootContainer.AddChild(emailTextInput)

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
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			s.password = args.InputText
		}),
	)
	passwordTextInput.SetText(s.password)
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

	if s.loginErr != "" {
		rootContainer.AddChild(widget.NewText(
			widget.TextOpts.Text(s.loginErr, fontFace, color.NRGBA{R: 255, G: 0, B: 0, A: 255}),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		))
		s.loginErr = ""
	}

	// auto focus the email text input
	emailTextInput.Focus(true)

	// register login handler with relevant widget events
	loginHandler := func(args interface{}) {
		email, password := emailTextInput.GetText(), passwordTextInput.GetText()
		if email == "" || password == "" {
			return
		}
		if err := s.onLogin(email, password); err != nil {
			log.Error("Failed to login: %v", err)
			if actionableErr, ok := err.(*ui.ActionableError); ok {
				s.loginErr = actionableErr.Message
			} else {
				s.loginErr = "Failed to login. Please try again."
			}
		}
		s.renderUI()
	}
	emailTextInput.SubmitEvent.AddHandler(loginHandler)
	passwordTextInput.SubmitEvent.AddHandler(loginHandler)
	button.ClickedEvent.AddHandler(loginHandler)

	ui := &ebitenui.UI{
		Container: rootContainer,
	}

	s.ui = ui
}

func (s *MenuScene) Update() error {
	s.ui.Update()
	return s.BaseScene.Update()
}

func (s *MenuScene) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
	s.BaseScene.Draw(screen)
}
