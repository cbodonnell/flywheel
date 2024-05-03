package objects

import (
	"fmt"
)

// SortedZIndexObject is a GameObject that maintains a sorted list of child objects by z-index.
type SortedZIndexObject struct {
	*BaseObject

	// sorted is a list of child objects sorted by z-index.
	sorted []GameObject
}

var _ GameObject = &SortedZIndexObject{}

func NewSortedZIndexObject(id string) *SortedZIndexObject {
	return &SortedZIndexObject{
		BaseObject: NewBaseObject(id, nil),
		sorted:     make([]GameObject, 0),
	}
}

func (o *SortedZIndexObject) AddChild(id string, child GameObject) error {
	if _, ok := o.children.idxIDObjects[id]; ok {
		return fmt.Errorf("child object with id already exists")
	}
	if err := InitTree(child); err != nil {
		return fmt.Errorf("failed to initialize child object tree: %v", err)
	}
	o.children.Add(id, child)
	child.SetParent(o)
	for i, obj := range o.sorted {
		if obj.GetZIndex() > child.GetZIndex() {
			o.sorted = append(o.sorted[:i], append([]GameObject{child}, o.sorted[i:]...)...)
			return nil
		}
	}
	o.sorted = append(o.sorted, child)
	return nil
}

func (o *SortedZIndexObject) RemoveChild(id string) error {
	child := o.children.Get(id)
	if child == nil {
		return fmt.Errorf("child object with id does not exist")
	}
	if err := DestroyTree(child); err != nil {
		return fmt.Errorf("failed to destroy child object tree: %v", err)
	}
	o.children.Remove(id)
	child.SetParent(nil)
	for i, obj := range o.sorted {
		if obj.GetID() == id {
			o.sorted = append(o.sorted[:i], o.sorted[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("child not found in sorted list")
}

func (o *SortedZIndexObject) GetChildren() []GameObject {
	return o.sorted
}
