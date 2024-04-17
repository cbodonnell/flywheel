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

	GetID() string
	GetChildren() []GameObject
	GetChild(id string) GameObject
	AddChild(id string, child GameObject) error
	RemoveChild(id string) error
}

type IndexedObjectList struct {
	objects []GameObject
	idxID   map[string]int
}

func NewIndexedObjectList() *IndexedObjectList {
	return &IndexedObjectList{
		objects: make([]GameObject, 0),
		idxID:   make(map[string]int),
	}
}

func (l *IndexedObjectList) Add(id string, object GameObject) {
	l.objects = append(l.objects, object)
	l.idxID[id] = len(l.objects) - 1
}

func (l *IndexedObjectList) Remove(id string) {
	idx, ok := l.idxID[id]
	if !ok {
		return
	}
	l.objects = append(l.objects[:idx], l.objects[idx+1:]...)
	for i := idx; i < len(l.objects); i++ {
		l.idxID[l.objects[i].GetID()] = i
	}
	delete(l.idxID, id)
}

func (l *IndexedObjectList) Get(id string) GameObject {
	idx, ok := l.idxID[id]
	if !ok {
		return nil
	}
	return l.objects[idx]
}

func (l *IndexedObjectList) GetAll() []GameObject {
	return l.objects
}

// BaseObject is a base implementation of GameObject.
// All game objects should embed this struct to inherit its methods.
type BaseObject struct {
	id       string
	children *IndexedObjectList
}

var _ GameObject = &BaseObject{}

func NewBaseObject(id string) *BaseObject {
	return &BaseObject{
		id:       id,
		children: NewIndexedObjectList(),
	}
}

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

func (o *BaseObject) GetID() string {
	return o.id
}

func (o *BaseObject) GetChildren() []GameObject {
	return o.children.GetAll()
}

func (o *BaseObject) GetChild(id string) GameObject {
	return o.children.Get(id)
}

func (o *BaseObject) AddChild(id string, child GameObject) error {
	if _, ok := o.children.idxID[id]; ok {
		return fmt.Errorf("child object with id already exists")
	}
	if err := InitTree(child); err != nil {
		return fmt.Errorf("failed to initialize child object tree: %v", err)
	}
	o.children.Add(id, child)
	return nil
}

func (o *BaseObject) RemoveChild(id string) error {
	child := o.children.Get(id)
	if child == nil {
		return fmt.Errorf("child object with id does not exist")
	}
	if err := DestroyTree(child); err != nil {
		return fmt.Errorf("failed to destroy child object tree: %v", err)
	}
	o.children.Remove(id)
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
