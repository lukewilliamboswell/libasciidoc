package site

import (
	"path/filepath"
	"sort"
	"strings"
)

// NavNode represents a page or directory in the navigation tree.
type NavNode struct {
	Title    string
	Href     string // root-relative URL, e.g. "/repo/guides/intro.html"
	Weight   int
	Children []*NavNode
	Active   bool
}

// buildNav constructs a navigation tree from a flat list of pages.
// basePath is the URL prefix (e.g. "/" or "/repo/").
func buildNav(pages []*page, basePath string) *NavNode {
	root := &NavNode{Title: "root"}
	// index: directory path -> nav node
	dirs := map[string]*NavNode{
		"": root,
	}

	// ensure all directories in the tree exist
	ensureDir := func(dirPath string) *NavNode {
		if n, ok := dirs[dirPath]; ok {
			return n
		}
		parts := strings.Split(dirPath, string(filepath.Separator))
		current := ""
		var parent *NavNode = root
		for _, part := range parts {
			if current == "" {
				current = part
			} else {
				current = current + string(filepath.Separator) + part
			}
			if n, ok := dirs[current]; ok {
				parent = n
				continue
			}
			n := &NavNode{
				Title: part,
				Href:  basePath + strings.ReplaceAll(current, string(filepath.Separator), "/") + "/",
			}
			parent.Children = append(parent.Children, n)
			dirs[current] = n
			parent = n
		}
		return parent
	}

	for _, p := range pages {
		dir := filepath.Dir(p.relPath)
		if dir == "." {
			dir = ""
		}
		if p.isIndex {
			// index.adoc represents its parent directory
			node := ensureDir(dir)
			node.Title = p.title
			if node.Title == "" {
				node.Title = filepath.Base(dir)
			}
			node.Href = basePath + strings.ReplaceAll(p.outputPath, string(filepath.Separator), "/")
			node.Weight = p.weight
		} else {
			parent := ensureDir(dir)
			node := &NavNode{
				Title:  p.title,
				Href:   basePath + strings.ReplaceAll(p.outputPath, string(filepath.Separator), "/"),
				Weight: p.weight,
			}
			if node.Title == "" {
				name := filepath.Base(p.relPath)
				node.Title = strings.TrimSuffix(name, filepath.Ext(name))
			}
			parent.Children = append(parent.Children, node)
		}
	}

	sortNav(root)
	return root
}

// sortNav recursively sorts children by weight ascending, then title alphabetically.
func sortNav(node *NavNode) {
	sort.SliceStable(node.Children, func(i, j int) bool {
		if node.Children[i].Weight != node.Children[j].Weight {
			return node.Children[i].Weight < node.Children[j].Weight
		}
		return node.Children[i].Title < node.Children[j].Title
	})
	for _, child := range node.Children {
		sortNav(child)
	}
}

// setActive marks the nav node matching currentPath as active, and returns
// true if any node in the subtree is active (so parents can be styled).
func setActive(node *NavNode, currentPath string) bool {
	active := false
	if node.Href == currentPath {
		node.Active = true
		active = true
	}
	for _, child := range node.Children {
		if setActive(child, currentPath) {
			active = true
		}
	}
	return active
}

// cloneNav deep-copies a NavNode tree so that Active can be set per-page.
func cloneNav(node *NavNode) *NavNode {
	c := &NavNode{
		Title:  node.Title,
		Href:   node.Href,
		Weight: node.Weight,
	}
	for _, child := range node.Children {
		c.Children = append(c.Children, cloneNav(child))
	}
	return c
}
