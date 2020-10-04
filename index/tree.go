/*
Copyright © 2020 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package index

import (
	pb "github.com/ItalyPaleAle/prvt/index/proto"
)

// IndexTreeNode is a node in the tree
type IndexTreeNode struct {
	Name     string
	File     *pb.IndexElement
	Children []*IndexTreeNode
}

// Find returns the child node with the given name
func (n *IndexTreeNode) Find(name string) *IndexTreeNode {
	if name == "" || n == nil || n.Children == nil || len(n.Children) < 1 {
		return nil
	}

	for _, el := range n.Children {
		if el.Name == name {
			return el
		}
	}

	return nil
}

// Add a new child node
// file can be empty if adding a folder
func (n *IndexTreeNode) Add(name string, file *pb.IndexElement) *IndexTreeNode {
	add := &IndexTreeNode{
		Children: make([]*IndexTreeNode, 0),
		Name:     name,
		File:     file,
	}
	if n.Children == nil {
		n.Children = []*IndexTreeNode{add}
	} else {
		n.Children = append(n.Children, add)
	}
	return add
}
