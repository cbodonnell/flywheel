package scenes

import (
	"fmt"
	"image"
	"image/color"

	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/cbodonnell/flywheel/client/ui"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories/models"
	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type CharacterSelectionScene struct {
	*BaseScene

	ui                  *ebitenui.UI
	fetchCharacters     func() ([]*models.Character, error)
	createCharacter     func(name string) (*models.Character, error)
	deleteCharacter     func(characterID int32) error
	onSelectCharacter   func(characterID int32) error
	characters          []*models.Character
	isDeletingCharacter bool
	deletingCharacterID int32
	selectCharacterErr  string
}

type CharacterSelectionSceneOpts struct {
	// FetchCharacters is a function that fetches the characters that the user can select from.
	FetchCharacters func() ([]*models.Character, error)
	// CreateCharacter is a function that creates a new character.
	CreateCharacter func(name string) (*models.Character, error)
	// DeleteCharacter is a function that deletes a character.
	DeleteCharacter func(characterID int32) error
	// OnSelectCharacter is a callback that is called when a character is selected.
	OnSelectCharacter func(characterID int32) error
}

var _ Scene = &CharacterSelectionScene{}

func NewCharacterSelectionScene(opts CharacterSelectionSceneOpts) (Scene, error) {
	return &CharacterSelectionScene{
		BaseScene:         NewBaseScene(objects.NewBaseObject("character-selection-root", nil)),
		fetchCharacters:   opts.FetchCharacters,
		createCharacter:   opts.CreateCharacter,
		deleteCharacter:   opts.DeleteCharacter,
		onSelectCharacter: opts.OnSelectCharacter,
		characters:        make([]*models.Character, 0),
	}, nil
}

func (s *CharacterSelectionScene) Init() error {
	characters, err := s.fetchCharacters()
	if err != nil {
		log.Error("Failed to fetch characters: %v", err)
		if actionableErr, ok := err.(*ui.ActionableError); ok {
			s.selectCharacterErr = actionableErr.Message
		} else {
			s.selectCharacterErr = "Failed to fetch characters"
		}
	} else {
		s.characters = characters
	}

	s.renderUI()

	return s.BaseScene.Init()
}

func (s *CharacterSelectionScene) renderUI() {
	neutralButtonImage := &widget.ButtonImage{
		Idle:    eimage.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255}),
		Hover:   eimage.NewNineSliceColor(color.NRGBA{R: 135, G: 135, B: 150, A: 255}),
		Pressed: eimage.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255}),
	}
	positiveButtonImage := &widget.ButtonImage{
		Idle:    eimage.NewNineSliceColor(color.NRGBA{R: 80, G: 170, B: 80, A: 255}), // Green color
		Hover:   eimage.NewNineSliceColor(color.NRGBA{R: 65, G: 135, B: 65, A: 255}), // Darker green for hover
		Pressed: eimage.NewNineSliceColor(color.NRGBA{R: 50, G: 100, B: 50, A: 255}), // Even darker green for pressed
	}
	negativeButtonImage := &widget.ButtonImage{
		Idle:    eimage.NewNineSliceColor(color.NRGBA{R: 170, G: 80, B: 80, A: 255}), // Red color
		Hover:   eimage.NewNineSliceColor(color.NRGBA{R: 135, G: 65, B: 65, A: 255}), // Darker red for hover
		Pressed: eimage.NewNineSliceColor(color.NRGBA{R: 100, G: 50, B: 50, A: 255}), // Even darker red for pressed
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

	for _, character := range s.characters {
		buttonContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(20),
			)),
		)

		button := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
			widget.ButtonOpts.Image(neutralButtonImage),
			widget.ButtonOpts.Text(character.Name, fontFace, &widget.ButtonTextColor{
				Idle:     color.NRGBA{254, 255, 255, 255},
				Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			}),
			widget.ButtonOpts.TextPadding(widget.Insets{
				Left:   15,
				Right:  15,
				Top:    5,
				Bottom: 5,
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				if err := s.onSelectCharacter(character.ID); err != nil {
					log.Error("Failed to select character: %v", err)
					if actionableErr, ok := err.(*ui.ActionableError); ok {
						s.selectCharacterErr = actionableErr.Message
					} else {
						s.selectCharacterErr = "Failed to select character"
					}
				}
				s.renderUI()
			}),
		)
		buttonContainer.AddChild(button)

		deleteButton := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionEnd,
					VerticalPosition:   widget.AnchorLayoutPositionStart,
				}),
			),
			widget.ButtonOpts.Image(negativeButtonImage),
			widget.ButtonOpts.Text("Delete", fontFace, &widget.ButtonTextColor{
				Idle:     color.NRGBA{254, 255, 255, 255},
				Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			}),
			widget.ButtonOpts.TextPadding(widget.Insets{
				Left:   15,
				Right:  15,
				Top:    5,
				Bottom: 5,
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				s.isDeletingCharacter = true
				s.deletingCharacterID = character.ID

				s.renderUI()
			}),
		)
		buttonContainer.AddChild(deleteButton)

		rootContainer.AddChild(buttonContainer)
	}

	if len(s.characters) < 3 {
		nameTextInput := widget.NewTextInput(
			widget.TextInputOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
					Stretch:  true,
				}),
			),
			widget.TextInputOpts.MobileInputMode("text"),
			widget.TextInputOpts.Image(&widget.TextInputImage{
				Idle:     eimage.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
				Disabled: eimage.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
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
			widget.TextInputOpts.Placeholder("New Character"),
		)
		rootContainer.AddChild(nameTextInput)

		createButton := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
			widget.ButtonOpts.Image(positiveButtonImage),
			widget.ButtonOpts.Text("Create", fontFace, &widget.ButtonTextColor{
				Idle:     color.NRGBA{254, 255, 255, 255},
				Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			}),
			widget.ButtonOpts.TextPadding(widget.Insets{
				Left:   15,
				Right:  15,
				Top:    5,
				Bottom: 5,
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				name := nameTextInput.GetText()
				if name == "" {
					return
				}

				character, err := s.createCharacter(name)
				if err != nil {
					log.Error("Failed to create character: %v", err)
					if actionableErr, ok := err.(*ui.ActionableError); ok {
						s.selectCharacterErr = actionableErr.Message
					} else {
						s.selectCharacterErr = "Failed to create character"
					}
				} else {
					s.characters = append(s.characters, character)
				}
				s.renderUI()
			}),
		)

		rootContainer.AddChild(createButton)
	}

	if s.selectCharacterErr != "" {
		rootContainer.AddChild(widget.NewText(
			widget.TextOpts.Text(s.selectCharacterErr, fontFace, color.NRGBA{R: 255, G: 0, B: 0, A: 255}),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		))
		s.selectCharacterErr = ""
	}

	ebitenUI := &ebitenui.UI{
		Container: rootContainer,
	}

	if s.isDeletingCharacter {
		windowContainer := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{100, 100, 100, 255})),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)

		deleteConfirmationContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(20),
				widget.RowLayoutOpts.Padding(widget.Insets{
					Top:    36,
					Left:   24,
					Right:  24,
					Bottom: 72,
				}),
			)),
		)

		deleteConfirmationText := widget.NewText(
			widget.TextOpts.Text("Are you sure?", fontFace, color.NRGBA{254, 255, 255, 255}),
		)
		deleteConfirmationContainer.AddChild(deleteConfirmationText)

		deleteConfirmationYesButton := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
			widget.ButtonOpts.Image(negativeButtonImage),
			widget.ButtonOpts.Text("Yes", fontFace, &widget.ButtonTextColor{
				Idle:     color.NRGBA{254, 255, 255, 255},
				Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			}),
			widget.ButtonOpts.TextPadding(widget.Insets{
				Left:   15,
				Right:  15,
				Top:    5,
				Bottom: 5,
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				err := s.deleteCharacter(s.deletingCharacterID)
				if err != nil {
					log.Error("Failed to delete character: %v", err)
					if actionableErr, ok := err.(*ui.ActionableError); ok {
						s.selectCharacterErr = actionableErr.Message
					} else {
						s.selectCharacterErr = "Failed to delete character"
					}
				} else {
					for i, c := range s.characters {
						if c.ID == s.deletingCharacterID {
							s.characters = append(s.characters[:i], s.characters[i+1:]...)
							break
						}
					}
				}

				s.isDeletingCharacter = false
				s.renderUI()
			}),
		)
		deleteConfirmationContainer.AddChild(deleteConfirmationYesButton)

		deleteConfirmationNoButton := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
			widget.ButtonOpts.Image(neutralButtonImage),
			widget.ButtonOpts.Text("No", fontFace, &widget.ButtonTextColor{
				Idle:     color.NRGBA{254, 255, 255, 255},
				Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			}),
			widget.ButtonOpts.TextPadding(widget.Insets{
				Left:   15,
				Right:  15,
				Top:    5,
				Bottom: 5,
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				s.isDeletingCharacter = false
				s.renderUI()
			}),
		)
		deleteConfirmationContainer.AddChild(deleteConfirmationNoButton)

		windowContainer.AddChild(deleteConfirmationContainer)

		deletingCharacterName := ""
		for _, character := range s.characters {
			if character.ID == s.deletingCharacterID {
				deletingCharacterName = character.Name
				break
			}
		}

		titleContainer := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{150, 150, 150, 255})),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		titleContainer.AddChild(widget.NewText(
			widget.TextOpts.Text(fmt.Sprintf("Delete %s", deletingCharacterName), fontFace, color.NRGBA{254, 255, 255, 255}),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			})),
		))

		window := widget.NewWindow(
			widget.WindowOpts.Contents(windowContainer),
			widget.WindowOpts.TitleBar(titleContainer, 48),
			widget.WindowOpts.Modal(),
			widget.WindowOpts.CloseMode(widget.CLICK_OUT),
			widget.WindowOpts.ClosedHandler(func(args *widget.WindowClosedEventArgs) {
				s.isDeletingCharacter = false
				s.renderUI()
			}),
		)

		x, y := window.Contents.PreferredSize()
		r := image.Rect(0, 0, x, y)
		r = r.Add(image.Point{X: 135, Y: 140})
		window.SetLocation(r)
		ebitenUI.AddWindow(window)
	}

	s.ui = ebitenUI
}

func (s *CharacterSelectionScene) Update() error {
	s.ui.Update()
	return s.BaseScene.Update()
}

func (s *CharacterSelectionScene) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
	s.BaseScene.Draw(screen)
}
