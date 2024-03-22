package objects

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// GameObject is the highest level interface for game related types.
// It provides methods for initializing, destroying, updating, and drawing game objects.
// It also provides methods for managing child game objects.
// All game objects should implement this interface.
// The BaseObject struct provides a base implementation of this interface.
type GameObject interface {
	Lifecycle

	// Child management methods
	GetChildren() map[string]GameObject
	GetChild(id string) (GameObject, error)
	AddChild(id string, child GameObject) error
	RemoveChild(id string) error
}

// BaseObject is a base implementation of GameObject.
// All game objects should embed this struct to inherit its methods.
type BaseObject struct {
	Children map[string]GameObject
}

var _ GameObject = &BaseObject{}

func (o *BaseObject) Init() error {
	return nil
}

func (o *BaseObject) Destroy() error {
	return nil
}

func (o *BaseObject) Update() error {
	return nil
}

func (o *BaseObject) Draw(screen *ebiten.Image) {}

func (o *BaseObject) GetChildren() map[string]GameObject {
	return o.Children
}

func (o *BaseObject) GetChild(id string) (GameObject, error) {
	if _, ok := o.Children[id]; !ok {
		return nil, fmt.Errorf("child object not found")
	}
	return o.Children[id], nil
}

func (o *BaseObject) AddChild(id string, child GameObject) error {
	if o.Children == nil {
		o.Children = make(map[string]GameObject)
	}
	if _, ok := o.Children[id]; ok {
		return fmt.Errorf("child object already exists")
	}
	if err := InitTree(child); err != nil {
		return fmt.Errorf("failed to initialize child object tree: %v", err)
	}
	o.Children[id] = child
	return nil
}

func (o *BaseObject) RemoveChild(id string) error {
	if _, ok := o.Children[id]; !ok {
		return fmt.Errorf("child object not found")
	}
	if err := DestroyTree(o.Children[id]); err != nil {
		return fmt.Errorf("failed to destroy child object tree: %v", err)
	}
	delete(o.Children, id)
	return nil
}

func InitTree(object GameObject) error {
	if err := object.Init(); err != nil {
		return fmt.Errorf("failed to initialize object: %v", err)
	}
	for _, child := range object.GetChildren() {
		if err := InitTree(child); err != nil {
			return fmt.Errorf("failed to initialize child object: %v", err)
		}
	}
	return nil
}

func DestroyTree(object GameObject) error {
	for _, child := range object.GetChildren() {
		if err := DestroyTree(child); err != nil {
			return fmt.Errorf("failed to destroy child object: %v", err)
		}
	}
	if err := object.Destroy(); err != nil {
		return fmt.Errorf("failed to destroy object: %v", err)
	}
	return nil
}

func UpdateTree(object GameObject) error {
	if err := object.Update(); err != nil {
		return fmt.Errorf("failed to update object: %v", err)
	}
	for _, child := range object.GetChildren() {
		if err := UpdateTree(child); err != nil {
			return fmt.Errorf("failed to update child object: %v", err)
		}
	}
	return nil
}

func DrawTree(object GameObject, screen *ebiten.Image) {
	object.Draw(screen)
	for _, child := range object.GetChildren() {
		DrawTree(child, screen)
	}
}
