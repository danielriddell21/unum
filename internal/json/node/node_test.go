package node

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func leaf(kind Kind, raw string) *Node {
	return &Node{Kind: kind, Raw: raw, Index: -1}
}

func objectNode(children ...*Node) *Node {
	n := &Node{Kind: KindObject, Index: -1}
	for _, c := range children {
		c.Parent = n
		n.Children = append(n.Children, c)
	}
	return n
}

func arrayNode(children ...*Node) *Node {
	n := &Node{Kind: KindArray, Index: -1}
	for i, c := range children {
		c.Parent = n
		c.Index = i
		n.Children = append(n.Children, c)
	}
	return n
}

func keyedLeaf(key string, kind Kind, raw string) *Node {
	return &Node{Kind: kind, Raw: raw, Key: key, Index: -1}
}

func TestKind_String(t *testing.T) {
	tests := []struct {
		k    Kind
		want string
	}{
		{KindNull, "null"},
		{KindBool, "bool"},
		{KindNumber, "number"},
		{KindString, "string"},
		{KindArray, "array"},
		{KindObject, "object"},
		{Kind(99), "unknown"},
	}
	for _, tc := range tests {
		if got := tc.k.String(); got != tc.want {
			t.Errorf("Kind(%d).String()=%q, want %q", tc.k, got, tc.want)
		}
	}
}

func TestAnnotate_HitAndMiss(t *testing.T) {
	n := &Node{}
	key := AnnotationKey{Lens: "test", Name: "val"}
	n.Annotate(Annotation{Key: key, Value: 42})

	v, ok := n.GetAnnotation("test", "val")
	if !ok || v != 42 {
		t.Errorf("GetAnnotation hit: got (%v, %v), want (42, true)", v, ok)
	}

	_, ok2 := n.GetAnnotation("test", "missing")
	if ok2 {
		t.Error("GetAnnotation should return false for unknown key")
	}
}

func TestAnnotate_MultipleAnnotations(t *testing.T) {
	n := &Node{}
	n.Annotate(Annotation{Key: AnnotationKey{"lens", "a"}, Value: 1})
	n.Annotate(Annotation{Key: AnnotationKey{"lens", "b"}, Value: 2})
	n.Annotate(Annotation{Key: AnnotationKey{"other", "x"}, Value: 3})

	anns := n.GetAnnotations("lens")
	if len(anns) != 2 {
		t.Fatalf("GetAnnotations: got %d, want 2", len(anns))
	}
}

func TestAnnotate_FiltersByLens(t *testing.T) {
	n := &Node{}
	n.Annotate(Annotation{Key: AnnotationKey{"lens1", "k"}, Value: "a"})
	n.Annotate(Annotation{Key: AnnotationKey{"lens2", "k"}, Value: "b"})

	got := n.GetAnnotations("lens2")
	if len(got) != 1 || got[0].Value != "b" {
		t.Errorf("GetAnnotations filter failed: %+v", got)
	}
}

func TestAnnotate_Concurrent(t *testing.T) {
	n := &Node{}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			n.Annotate(Annotation{Key: AnnotationKey{"lens", "k"}, Value: i})
		}(i)
	}
	wg.Wait()
	anns := n.GetAnnotations("lens")
	if len(anns) != 50 {
		t.Errorf("concurrent Annotate: got %d annotations, want 50", len(anns))
	}
}

func TestWalk_PreOrder(t *testing.T) {
	root := objectNode(
		keyedLeaf("a", KindNumber, "1"),
		keyedLeaf("b", KindNumber, "2"),
	)
	var visited []string
	root.Walk(func(n *Node) bool {
		visited = append(visited, n.Key)
		return true
	})
	// root key is "" then "a", "b"
	if len(visited) != 3 {
		t.Fatalf("Walk visited %d nodes, want 3", len(visited))
	}
	if visited[1] != "a" || visited[2] != "b" {
		t.Errorf("Walk order: %v", visited)
	}
}

func TestWalk_EarlyReturn(t *testing.T) {
	child := keyedLeaf("child", KindNumber, "1")
	root := objectNode(child)
	count := 0
	root.Walk(func(n *Node) bool {
		count++
		return false // skip subtree immediately
	})
	if count != 1 {
		t.Errorf("Walk early-return: visited %d, want 1", count)
	}
}

func TestWalkPost_ChildrenBeforeParent(t *testing.T) {
	child1 := keyedLeaf("c1", KindNumber, "1")
	child2 := keyedLeaf("c2", KindNumber, "2")
	root := objectNode(child1, child2)

	var order []string
	root.WalkPost(func(n *Node) {
		order = append(order, n.Key)
	})
	// children before root
	if order[len(order)-1] != "" {
		t.Errorf("WalkPost: root should be last, got order %v", order)
	}
	if order[0] != "c1" || order[1] != "c2" {
		t.Errorf("WalkPost order: %v", order)
	}
}

func TestPath_Root(t *testing.T) {
	n := &Node{}
	if got := n.Path(); got != "." {
		t.Errorf("Path(root)=%q, want \".\"", got)
	}
}

func TestPath_ObjectKey(t *testing.T) {
	root := objectNode(keyedLeaf("name", KindString, `"alice"`))
	child := root.Children[0]
	if got := child.Path(); got != ".name" {
		t.Errorf("Path(key)=%q, want \".name\"", got)
	}
}

func TestPath_ArrayElement(t *testing.T) {
	child := leaf(KindNumber, "1")
	arr := arrayNode(child)
	// Array element paths include a leading dot from the root segment.
	if got := arr.Children[0].Path(); got != ".[0]" {
		t.Errorf("Path(arr[0])=%q, want \".[0]\"", got)
	}
}

func TestPath_Nested(t *testing.T) {
	inner := keyedLeaf("age", KindNumber, "30")
	innerObj := objectNode(inner)
	innerObj.Key = "user"
	innerObj.Index = -1
	_ = objectNode(innerObj)

	if got := inner.Path(); got != ".user.age" {
		t.Errorf("Path(nested)=%q, want \".user.age\"", got)
	}
}

func TestCountNodes(t *testing.T) {
	root := objectNode(
		keyedLeaf("a", KindNumber, "1"),
		arrayNode(leaf(KindNull, "null"), leaf(KindBool, "true")),
	)
	// root + a + array + null + true = 5
	if got := root.CountNodes(); got != 5 {
		t.Errorf("CountNodes=%d, want 5", got)
	}
}

func TestToAny_Null(t *testing.T) {
	n := leaf(KindNull, "null")
	if n.ToAny() != nil {
		t.Error("ToAny(null) should return nil")
	}
}

func TestToAny_Bool(t *testing.T) {
	tr := leaf(KindBool, "true")
	if tr.ToAny() != true {
		t.Error("ToAny(true) should return true")
	}
	fa := leaf(KindBool, "false")
	if fa.ToAny() != false {
		t.Error("ToAny(false) should return false")
	}
}

func TestToAny_Number(t *testing.T) {
	n := leaf(KindNumber, "3.14")
	v := n.ToAny()
	num, ok := v.(json.Number)
	if !ok {
		t.Fatalf("ToAny(number): got %T, want json.Number", v)
	}
	if num.String() != "3.14" {
		t.Errorf("ToAny(number): %q, want \"3.14\"", num.String())
	}
}

func TestToAny_String(t *testing.T) {
	n := leaf(KindString, `"hello"`)
	v := n.ToAny()
	s, ok := v.(string)
	if !ok || s != "hello" {
		t.Errorf("ToAny(string): got %v (%T), want \"hello\"", v, v)
	}
}

func TestToAny_Array(t *testing.T) {
	root := arrayNode(leaf(KindNumber, "1"), leaf(KindNumber, "2"))
	v := root.ToAny()
	arr, ok := v.([]any)
	if !ok || len(arr) != 2 {
		t.Errorf("ToAny(array): got %T len=%d, want []any len=2", v, len(arr))
	}
}

func TestToAny_Object(t *testing.T) {
	root := objectNode(keyedLeaf("k", KindNumber, "7"))
	v := root.ToAny()
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("ToAny(object): got %T, want map[string]any", v)
	}
	if m["k"] == nil {
		t.Error("ToAny(object): missing key \"k\"")
	}
}

func TestMarshalJSON_Primitives(t *testing.T) {
	tests := []struct {
		n    *Node
		want string
	}{
		{leaf(KindNull, "null"), "null"},
		{leaf(KindBool, "true"), "true"},
		{leaf(KindBool, "false"), "false"},
		{leaf(KindNumber, "42"), "42"},
		{leaf(KindString, `"hi"`), `"hi"`},
	}
	for _, tc := range tests {
		b, err := tc.n.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJSON(%q): %v", tc.want, err)
			continue
		}
		if string(b) != tc.want {
			t.Errorf("MarshalJSON: got %q, want %q", b, tc.want)
		}
	}
}

func TestMarshalJSON_EmptyArray(t *testing.T) {
	n := &Node{Kind: KindArray}
	b, err := n.MarshalJSON()
	if err != nil || string(b) != "[]" {
		t.Errorf("MarshalJSON(empty array): %q %v", b, err)
	}
}

func TestMarshalJSON_EmptyObject(t *testing.T) {
	n := &Node{Kind: KindObject}
	b, err := n.MarshalJSON()
	if err != nil || string(b) != "{}" {
		t.Errorf("MarshalJSON(empty object): %q %v", b, err)
	}
}

func TestMarshalJSON_Array(t *testing.T) {
	n := arrayNode(leaf(KindNumber, "1"), leaf(KindNumber, "2"))
	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "[1,2]" {
		t.Errorf("MarshalJSON(array): %q", b)
	}
}

func TestMarshalJSON_Object(t *testing.T) {
	n := objectNode(
		keyedLeaf("x", KindNumber, "1"),
		keyedLeaf("y", KindBool, "true"),
	)
	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	// key order is preserved
	if !strings.Contains(string(b), `"x":1`) || !strings.Contains(string(b), `"y":true`) {
		t.Errorf("MarshalJSON(object): %q", b)
	}
}

func TestMarshalJSON_Nested(t *testing.T) {
	inner := arrayNode(leaf(KindNull, "null"))
	inner.Key = "arr"
	inner.Index = -1
	root := objectNode(inner)
	b, err := root.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"arr":[null]}` {
		t.Errorf("MarshalJSON(nested): %q", b)
	}
}
