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

type AuthScene struct {
	*BaseScene

	onLogin         func(email, password string) error
	onRegister      func(email, password string) error
	isRegistering   bool
	ui              *ebitenui.UI
	emailTextInput  *widget.TextInput
	email           string
	password        string
	confirmPassword string
	errMsg          string
}

type AuthSceneOptions struct {
	OnLogin    func(email, password string) error
	OnRegister func(email, password string) error
}

var _ Scene = &AuthScene{}

func NewAuthScene(opts AuthSceneOptions) (Scene, error) {
	return &AuthScene{
		BaseScene:  NewBaseScene(objects.NewBaseObject("auth-root", nil)),
		onLogin:    opts.OnLogin,
		onRegister: opts.OnRegister,
		// isRegistering: true, // TODO: create a
	}, nil
}

func (s *AuthScene) Init() error {
	s.renderUI()
	s.emailTextInput.Focus(true)
	return s.BaseScene.Init()
}

func (s *AuthScene) renderUI() {
	buttonImage := &widget.ButtonImage{
		Idle:    image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255}),
		Hover:   image.NewNineSliceColor(color.NRGBA{R: 135, G: 135, B: 150, A: 255}),
		Pressed: image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255}),
	}

	linkButtonImage := &widget.ButtonImage{
		Idle:    image.NewNineSliceColor(color.NRGBA{R: 0, G: 0, B: 0, A: 0}),
		Hover:   image.NewNineSliceColor(color.NRGBA{R: 0, G: 0, B: 0, A: 0}),
		Pressed: image.NewNineSliceColor(color.NRGBA{R: 0, G: 0, B: 0, A: 0}),
	}

	normalFontFace := fonts.TTFNormalFont

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
		widget.TextInputOpts.Face(normalFontFace),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(normalFontFace, 2),
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
		widget.TextInputOpts.Face(normalFontFace),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(normalFontFace, 2),
		),
		widget.TextInputOpts.Placeholder("Password"),
		widget.TextInputOpts.Secure(true),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			s.password = args.InputText
		}),
	)
	passwordTextInput.SetText(s.password)
	rootContainer.AddChild(passwordTextInput)

	confirmPasswordTextInput := widget.NewTextInput(
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
		widget.TextInputOpts.Face(normalFontFace),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(normalFontFace, 2),
		),
		widget.TextInputOpts.Placeholder("Confirm Password"),
		widget.TextInputOpts.Secure(true),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			s.confirmPassword = args.InputText
		}),
	)
	confirmPasswordTextInput.SetText(s.confirmPassword)

	toggleButtonText := "New User?"
	if s.isRegistering {
		toggleButtonText = "Existing User?"
		rootContainer.AddChild(confirmPasswordTextInput)
	}

	toggleRegisterContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	toggleRegisterLink := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
			}),
		),
		widget.ButtonOpts.Image(linkButtonImage),
		widget.ButtonOpts.Text(toggleButtonText, normalFontFace, &widget.ButtonTextColor{
			Idle:     color.NRGBA{254, 255, 255, 255},
			Hover:    color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   5,
			Right:  5,
			Top:    5,
			Bottom: 5,
		}),
		widget.ButtonOpts.TextPosition(widget.TextPositionEnd, widget.TextPositionEnd),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			s.isRegistering = !s.isRegistering
			s.renderUI()
		}),
	)
	toggleRegisterContainer.AddChild(toggleRegisterLink)

	submitButton := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text("Submit", normalFontFace, &widget.ButtonTextColor{
			Idle:     color.NRGBA{254, 255, 255, 255},
			Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   15,
			Right:  15,
			Top:    5,
			Bottom: 5,
		}),
	)

	buttonContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Top:    0,
				Left:   0,
				Right:  0,
				Bottom: 0,
			}),
			widget.GridLayoutOpts.Spacing(20, 0),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
		)),
	)
	buttonContainer.AddChild(submitButton)
	buttonContainer.AddChild(toggleRegisterContainer)

	rootContainer.AddChild(buttonContainer)

	if s.errMsg != "" {
		rootContainer.AddChild(widget.NewText(
			widget.TextOpts.Text(s.errMsg, normalFontFace, color.NRGBA{R: 255, G: 0, B: 0, A: 255}),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionStart,
				}),
			),
		))
		s.errMsg = ""
	}

	submitHandler := func(args interface{}) {
		defer s.renderUI()
		if s.isRegistering {
			email, password, confirmPassword := emailTextInput.GetText(), passwordTextInput.GetText(), confirmPasswordTextInput.GetText()
			if email == "" {
				s.errMsg = "Email is required."
				return
			}
			if password == "" {
				s.errMsg = "Password is required."
				return
			}
			if confirmPassword == "" {
				s.errMsg = "Confirm password is required."
				return
			}
			if password != confirmPassword {
				s.errMsg = "Passwords do not match."
				return
			}
			if err := s.onRegister(email, password); err != nil {
				log.Error("Failed to register: %v", err)
				if actionableErr, ok := err.(*ui.ActionableError); ok {
					s.errMsg = actionableErr.Message
				} else {
					s.errMsg = "Failed to register. Please try again."
				}
			}
		} else {
			email, password := emailTextInput.GetText(), passwordTextInput.GetText()
			if email == "" {
				s.errMsg = "Email is required."
				return
			}
			if password == "" {
				s.errMsg = "Password is required."
				return
			}
			if err := s.onLogin(email, password); err != nil {
				log.Error("Failed to login: %v", err)
				if actionableErr, ok := err.(*ui.ActionableError); ok {
					s.errMsg = actionableErr.Message
				} else {
					s.errMsg = "Failed to login. Please try again."
				}
			}
		}
	}
	emailTextInput.SubmitEvent.AddHandler(submitHandler)
	passwordTextInput.SubmitEvent.AddHandler(submitHandler)
	confirmPasswordTextInput.SubmitEvent.AddHandler(submitHandler)
	submitButton.ClickedEvent.AddHandler(submitHandler)

	ui := &ebitenui.UI{
		Container: rootContainer,
	}

	s.ui = ui
	s.emailTextInput = emailTextInput
}

func (s *AuthScene) Update() error {
	s.ui.Update()
	return s.BaseScene.Update()
}

func (s *AuthScene) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
	s.BaseScene.Draw(screen)
}
