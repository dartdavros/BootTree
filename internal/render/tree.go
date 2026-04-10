package render

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"boottree/internal/core/model"
)

type TreeRenderOptions struct {
	MaxDepth int
}

type treeNode struct {
	name     string
	isDir    bool
	children map[string]*treeNode
}

func RenderTree(snapshot model.TreeSnapshot, options TreeRenderOptions) string {
	rootName := filepath.Base(filepath.Clean(snapshot.Root))
	if rootName == "." || rootName == string(filepath.Separator) || rootName == "" {
		rootName = snapshot.Root
	}

	root := &treeNode{name: rootName, isDir: true, children: map[string]*treeNode{}}
	for _, entry := range snapshot.Entries {
		segments := splitPath(entry.Path)
		if len(segments) == 0 {
			continue
		}
		if options.MaxDepth > 0 && len(segments) > options.MaxDepth {
			continue
		}
		insertTreeEntry(root, segments, entry.IsDir)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s/\n", root.name)
	renderTreeChildren(&b, root, "")
	return strings.TrimRight(b.String(), "\n")
}

func insertTreeEntry(root *treeNode, segments []string, isDir bool) {
	current := root
	for index, segment := range segments {
		child, ok := current.children[segment]
		if !ok {
			child = &treeNode{
				name:     segment,
				isDir:    index < len(segments)-1 || isDir,
				children: map[string]*treeNode{},
			}
			current.children[segment] = child
		}
		if index == len(segments)-1 {
			child.isDir = isDir
		}
		current = child
	}
}

func renderTreeChildren(b *strings.Builder, node *treeNode, prefix string) {
	children := orderedChildren(node)
	for index, child := range children {
		branch := "├── "
		nextPrefix := prefix + "│   "
		if index == len(children)-1 {
			branch = "└── "
			nextPrefix = prefix + "    "
		}

		name := child.name
		if child.isDir {
			name += "/"
		}
		fmt.Fprintf(b, "%s%s%s\n", prefix, branch, name)
		if child.isDir && len(child.children) > 0 {
			renderTreeChildren(b, child, nextPrefix)
		}
	}
}

func orderedChildren(node *treeNode) []*treeNode {
	children := make([]*treeNode, 0, len(node.children))
	for _, child := range node.children {
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		if children[i].isDir != children[j].isDir {
			return children[i].isDir
		}
		return children[i].name < children[j].name
	})
	return children
}

func splitPath(relPath string) []string {
	cleaned := filepath.Clean(relPath)
	if cleaned == "." || cleaned == "" {
		return nil
	}
	return strings.Split(cleaned, string(filepath.Separator))
}
