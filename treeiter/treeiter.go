//go:build !solution

package treeiter

type Node interface {
	Left() *Node
	Right() *Node
}

func DoInOrder[Noda interface {
	Left() *Noda
	Right() *Noda
}](root *Noda, cb func(val *Noda)) {
	if root == nil {
		return
	}
	DoInOrder((*root).Left(), cb)
	cb(root)
	DoInOrder((*root).Right(), cb)

}
