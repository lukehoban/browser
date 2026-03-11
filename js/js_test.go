package js

import (
	"strings"
	"testing"

	"github.com/lukehoban/browser/dom"
)

// eval is a test helper that executes JS and returns the global environment.
func evalScript(t *testing.T, src string) *Interpreter {
	t.Helper()
	interp := NewInterpreter()
	parser := NewParser(src)
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if err := interp.Eval(prog); err != nil {
		t.Fatalf("eval error: %v", err)
	}
	return interp
}

func getVar(interp *Interpreter, name string) *Value {
	v, _ := interp.global.Get(name)
	if v == nil {
		return Undefined
	}
	return v
}

// ---- Literals and basic arithmetic ----

func TestLiterals(t *testing.T) {
	interp := evalScript(t, `
		var a = 42;
		var b = 3.14;
		var c = "hello";
		var d = true;
		var e = null;
		var f = undefined;
	`)
	if v := getVar(interp, "a"); v.ToNumber() != 42 {
		t.Errorf("a = %v, want 42", v)
	}
	if v := getVar(interp, "b"); v.ToNumber() != 3.14 {
		t.Errorf("b = %v, want 3.14", v)
	}
	if v := getVar(interp, "c"); v.ToString() != "hello" {
		t.Errorf("c = %v, want hello", v)
	}
	if v := getVar(interp, "d"); !v.ToBoolean() {
		t.Errorf("d should be true")
	}
	if v := getVar(interp, "e"); !v.IsNull() {
		t.Errorf("e should be null")
	}
	if v := getVar(interp, "f"); !v.IsUndefined() {
		t.Errorf("f should be undefined")
	}
}

func TestArithmetic(t *testing.T) {
	interp := evalScript(t, `
		var a = 10 + 3;
		var b = 10 - 3;
		var c = 10 * 3;
		var d = 10 / 4;
		var e = 10 % 3;
		var f = 2 ** 8;
	`)
	tests := []struct{ name string; want float64 }{
		{"a", 13}, {"b", 7}, {"c", 30}, {"d", 2.5}, {"e", 1}, {"f", 256},
	}
	for _, tt := range tests {
		v := getVar(interp, tt.name)
		if v.ToNumber() != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, v.ToNumber(), tt.want)
		}
	}
}

func TestStringConcatenation(t *testing.T) {
	interp := evalScript(t, `var s = "hello" + " " + "world";`)
	if v := getVar(interp, "s"); v.ToString() != "hello world" {
		t.Errorf("s = %q, want %q", v.ToString(), "hello world")
	}
}

// ---- Variables ----

func TestLetConst(t *testing.T) {
	interp := evalScript(t, `
		let x = 10;
		const y = 20;
		var z = x + y;
	`)
	if v := getVar(interp, "z"); v.ToNumber() != 30 {
		t.Errorf("z = %v, want 30", v.ToNumber())
	}
}

// ---- Control flow ----

func TestIfElse(t *testing.T) {
	interp := evalScript(t, `
		var result;
		if (10 > 5) {
			result = "yes";
		} else {
			result = "no";
		}
	`)
	if v := getVar(interp, "result"); v.ToString() != "yes" {
		t.Errorf("result = %q, want yes", v.ToString())
	}
}

func TestWhileLoop(t *testing.T) {
	interp := evalScript(t, `
		var i = 0;
		var sum = 0;
		while (i < 5) {
			sum += i;
			i++;
		}
	`)
	if v := getVar(interp, "sum"); v.ToNumber() != 10 {
		t.Errorf("sum = %v, want 10", v.ToNumber())
	}
}

func TestForLoop(t *testing.T) {
	interp := evalScript(t, `
		var sum = 0;
		for (var i = 0; i < 5; i++) {
			sum += i;
		}
	`)
	if v := getVar(interp, "sum"); v.ToNumber() != 10 {
		t.Errorf("sum = %v, want 10", v.ToNumber())
	}
}

func TestForOfArray(t *testing.T) {
	interp := evalScript(t, `
		var result = [];
		var arr = [1, 2, 3];
		for (var x of arr) {
			result.push(x * 2);
		}
	`)
	v := getVar(interp, "result")
	if v.typ != TypeObject || !v.objVal.isArray {
		t.Fatal("result should be an array")
	}
	elems := v.objVal.ArrayElements()
	if len(elems) != 3 {
		t.Fatalf("result length = %d, want 3", len(elems))
	}
	want := []float64{2, 4, 6}
	for i, e := range elems {
		if e.ToNumber() != want[i] {
			t.Errorf("result[%d] = %v, want %v", i, e.ToNumber(), want[i])
		}
	}
}

func TestBreakContinue(t *testing.T) {
	interp := evalScript(t, `
		var sum = 0;
		for (var i = 0; i < 10; i++) {
			if (i === 5) break;
			if (i % 2 === 0) continue;
			sum += i;
		}
	`)
	// i=1, i=3 are odd and < 5; sum = 1+3 = 4
	if v := getVar(interp, "sum"); v.ToNumber() != 4 {
		t.Errorf("sum = %v, want 4", v.ToNumber())
	}
}

// ---- Functions ----

func TestFunctionDeclaration(t *testing.T) {
	interp := evalScript(t, `
		function add(a, b) {
			return a + b;
		}
		var result = add(3, 4);
	`)
	if v := getVar(interp, "result"); v.ToNumber() != 7 {
		t.Errorf("result = %v, want 7", v.ToNumber())
	}
}

func TestFunctionExpression(t *testing.T) {
	interp := evalScript(t, `
		var mul = function(a, b) { return a * b; };
		var result = mul(4, 5);
	`)
	if v := getVar(interp, "result"); v.ToNumber() != 20 {
		t.Errorf("result = %v, want 20", v.ToNumber())
	}
}

func TestArrowFunction(t *testing.T) {
	interp := evalScript(t, `
		var double = x => x * 2;
		var result = double(7);
	`)
	if v := getVar(interp, "result"); v.ToNumber() != 14 {
		t.Errorf("result = %v, want 14", v.ToNumber())
	}
}

func TestClosure(t *testing.T) {
	interp := evalScript(t, `
		function makeCounter() {
			var count = 0;
			return function() {
				count++;
				return count;
			};
		}
		var counter = makeCounter();
		var a = counter();
		var b = counter();
		var c = counter();
	`)
	tests := []struct{ name string; want float64 }{
		{"a", 1}, {"b", 2}, {"c", 3},
	}
	for _, tt := range tests {
		if v := getVar(interp, tt.name); v.ToNumber() != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, v.ToNumber(), tt.want)
		}
	}
}

func TestRecursion(t *testing.T) {
	interp := evalScript(t, `
		function factorial(n) {
			if (n <= 1) return 1;
			return n * factorial(n - 1);
		}
		var result = factorial(5);
	`)
	if v := getVar(interp, "result"); v.ToNumber() != 120 {
		t.Errorf("result = %v, want 120", v.ToNumber())
	}
}

// ---- Objects and arrays ----

func TestObjectLiteral(t *testing.T) {
	interp := evalScript(t, `
		var obj = { name: "Alice", age: 30 };
		var name = obj.name;
		var age = obj["age"];
	`)
	if v := getVar(interp, "name"); v.ToString() != "Alice" {
		t.Errorf("name = %q, want Alice", v.ToString())
	}
	if v := getVar(interp, "age"); v.ToNumber() != 30 {
		t.Errorf("age = %v, want 30", v.ToNumber())
	}
}

func TestArrayLiteral(t *testing.T) {
	interp := evalScript(t, `
		var arr = [1, 2, 3];
		var first = arr[0];
		var len = arr.length;
		arr.push(4);
		var len2 = arr.length;
	`)
	if v := getVar(interp, "first"); v.ToNumber() != 1 {
		t.Errorf("first = %v, want 1", v.ToNumber())
	}
	if v := getVar(interp, "len"); v.ToNumber() != 3 {
		t.Errorf("len = %v, want 3", v.ToNumber())
	}
	if v := getVar(interp, "len2"); v.ToNumber() != 4 {
		t.Errorf("len2 = %v, want 4", v.ToNumber())
	}
}

func TestArrayMethods(t *testing.T) {
	interp := evalScript(t, `
		var arr = [3, 1, 4, 1, 5, 9, 2, 6];
		var mapped = arr.map(x => x * 2);
		var filtered = arr.filter(x => x > 3);
		var reduced = arr.reduce((acc, x) => acc + x, 0);
		var joined = arr.join("-");
		var found = arr.find(x => x > 4);
		var some = arr.some(x => x > 8);
		var every = arr.every(x => x > 0);
	`)
	if v := getVar(interp, "reduced"); v.ToNumber() != 31 {
		t.Errorf("reduced = %v, want 31", v.ToNumber())
	}
	if v := getVar(interp, "found"); v.ToNumber() != 5 {
		t.Errorf("found = %v, want 5", v.ToNumber())
	}
	if v := getVar(interp, "some"); !v.ToBoolean() {
		t.Errorf("some should be true")
	}
	if v := getVar(interp, "every"); !v.ToBoolean() {
		t.Errorf("every should be true")
	}
	if v := getVar(interp, "joined"); v.ToString() != "3-1-4-1-5-9-2-6" {
		t.Errorf("joined = %q", v.ToString())
	}
}

// ---- String methods ----

func TestStringMethods(t *testing.T) {
	interp := evalScript(t, `
		var s = "Hello, World!";
		var upper = s.toUpperCase();
		var lower = s.toLowerCase();
		var idx = s.indexOf("World");
		var sub = s.slice(7, 12);
		var split = s.split(", ");
		var trimmed = "  hello  ".trim();
		var starts = s.startsWith("Hello");
		var ends = s.endsWith("!");
		var includes = s.includes("World");
		var replaced = s.replace("World", "JS");
	`)
	if v := getVar(interp, "upper"); v.ToString() != "HELLO, WORLD!" {
		t.Errorf("upper = %q", v.ToString())
	}
	if v := getVar(interp, "lower"); v.ToString() != "hello, world!" {
		t.Errorf("lower = %q", v.ToString())
	}
	if v := getVar(interp, "idx"); v.ToNumber() != 7 {
		t.Errorf("idx = %v, want 7", v.ToNumber())
	}
	if v := getVar(interp, "sub"); v.ToString() != "World" {
		t.Errorf("sub = %q, want World", v.ToString())
	}
	if v := getVar(interp, "trimmed"); v.ToString() != "hello" {
		t.Errorf("trimmed = %q, want hello", v.ToString())
	}
	if v := getVar(interp, "starts"); !v.ToBoolean() {
		t.Errorf("starts should be true")
	}
	if v := getVar(interp, "ends"); !v.ToBoolean() {
		t.Errorf("ends should be true")
	}
	if v := getVar(interp, "includes"); !v.ToBoolean() {
		t.Errorf("includes should be true")
	}
	if v := getVar(interp, "replaced"); v.ToString() != "Hello, JS!" {
		t.Errorf("replaced = %q, want 'Hello, JS!'", v.ToString())
	}
}

// ---- Template literals ----

func TestTemplateLiterals(t *testing.T) {
	interp := evalScript(t, `
		var name = "World";
		var greeting = ` + "`" + `Hello, ${name}!` + "`" + `;
		var calc = ` + "`" + `2 + 2 = ${2 + 2}` + "`" + `;
	`)
	if v := getVar(interp, "greeting"); v.ToString() != "Hello, World!" {
		t.Errorf("greeting = %q, want 'Hello, World!'", v.ToString())
	}
	if v := getVar(interp, "calc"); v.ToString() != "2 + 2 = 4" {
		t.Errorf("calc = %q, want '2 + 2 = 4'", v.ToString())
	}
}

// ---- Math ----

func TestMath(t *testing.T) {
	interp := evalScript(t, `
		var a = Math.abs(-5);
		var b = Math.max(1, 2, 3);
		var c = Math.min(1, 2, 3);
		var d = Math.floor(3.9);
		var e = Math.ceil(3.1);
		var f = Math.round(3.5);
		var g = Math.sqrt(16);
		var h = Math.pow(2, 10);
	`)
	tests := []struct{ name string; want float64 }{
		{"a", 5}, {"b", 3}, {"c", 1}, {"d", 3}, {"e", 4}, {"f", 4}, {"g", 4}, {"h", 1024},
	}
	for _, tt := range tests {
		if v := getVar(interp, tt.name); v.ToNumber() != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, v.ToNumber(), tt.want)
		}
	}
}

// ---- JSON ----

func TestJSON(t *testing.T) {
	interp := evalScript(t, `
		var obj = { a: 1, b: "hello" };
		var serialized = JSON.stringify(obj);
		var parsed = JSON.parse('{"x":42}');
		var x = parsed.x;
	`)
	if v := getVar(interp, "x"); v.ToNumber() != 42 {
		t.Errorf("x = %v, want 42", v.ToNumber())
	}
}

// ---- Try/catch ----

func TestTryCatch(t *testing.T) {
	interp := evalScript(t, `
		var result;
		try {
			throw new Error("test error");
		} catch (e) {
			result = e.message;
		}
	`)
	if v := getVar(interp, "result"); v.ToString() != "test error" {
		t.Errorf("result = %q, want 'test error'", v.ToString())
	}
}

// ---- Classes ----

func TestClass(t *testing.T) {
	interp := evalScript(t, `
		class Animal {
			constructor(name) {
				this.name = name;
			}
			speak() {
				return this.name + " makes a noise.";
			}
		}
		var a = new Animal("Dog");
		var result = a.speak();
		var name = a.name;
	`)
	if v := getVar(interp, "name"); v.ToString() != "Dog" {
		t.Errorf("name = %q, want Dog", v.ToString())
	}
	if v := getVar(interp, "result"); v.ToString() != "Dog makes a noise." {
		t.Errorf("result = %q, want 'Dog makes a noise.'", v.ToString())
	}
}

// ---- Comparison ----

func TestComparison(t *testing.T) {
	interp := evalScript(t, `
		var a = (1 == "1");
		var b = (1 === "1");
		var c = (1 === 1);
		var d = (null == undefined);
		var e = (null === undefined);
	`)
	if v := getVar(interp, "a"); !v.ToBoolean() {
		t.Error("1 == '1' should be true")
	}
	if v := getVar(interp, "b"); v.ToBoolean() {
		t.Error("1 === '1' should be false")
	}
	if v := getVar(interp, "c"); !v.ToBoolean() {
		t.Error("1 === 1 should be true")
	}
	if v := getVar(interp, "d"); !v.ToBoolean() {
		t.Error("null == undefined should be true")
	}
	if v := getVar(interp, "e"); v.ToBoolean() {
		t.Error("null === undefined should be false")
	}
}

// ---- DOM manipulation ----

func makeTestDoc(htmlStr string) *dom.Node {
	doc := dom.NewDocument()
	html := dom.NewElement("html")
	body := dom.NewElement("body")
	body.AppendChild(dom.NewText(htmlStr))
	html.AppendChild(body)
	doc.AppendChild(html)
	return doc
}

func TestDOMGetElementById(t *testing.T) {
	doc := dom.NewDocument()
	htmlEl := dom.NewElement("html")
	body := dom.NewElement("body")
	div := dom.NewElement("div")
	div.SetAttribute("id", "myDiv")
	div.AppendChild(dom.NewText("original"))
	body.AppendChild(div)
	htmlEl.AppendChild(body)
	doc.AppendChild(htmlEl)

	Execute(doc)
	// Nothing to execute yet; just verify it doesn't crash

	interp := NewInterpreter()
	interp.SetupDOM(doc)

	if err := func() error {
		parser := NewParser(`
			var el = document.getElementById("myDiv");
			var found = (el !== null);
		`)
		prog, err := parser.ParseProgram()
		if err != nil {
			return err
		}
		return interp.Eval(prog)
	}(); err != nil {
		t.Fatalf("error: %v", err)
	}

	if v := getVar(interp, "found"); !v.ToBoolean() {
		t.Error("getElementById should find the element")
	}
}

func TestDOMTextContent(t *testing.T) {
	doc := dom.NewDocument()
	htmlEl := dom.NewElement("html")
	body := dom.NewElement("body")
	p := dom.NewElement("p")
	p.SetAttribute("id", "para")
	p.AppendChild(dom.NewText("original"))
	body.AppendChild(p)
	htmlEl.AppendChild(body)
	doc.AppendChild(htmlEl)

	interp := NewInterpreter()
	interp.SetupDOM(doc)

	src := `document.getElementById("para").textContent = "updated";`
	parser := NewParser(src)
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if err := interp.Eval(prog); err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Verify the DOM was updated
	if len(p.Children) == 0 || p.Children[0].Data != "updated" {
		t.Errorf("p.textContent should be 'updated', got %q", getTextContent(p))
	}
}

func TestDOMStyleChange(t *testing.T) {
	doc := dom.NewDocument()
	htmlEl := dom.NewElement("html")
	body := dom.NewElement("body")
	div := dom.NewElement("div")
	div.SetAttribute("id", "box")
	body.AppendChild(div)
	htmlEl.AppendChild(body)
	doc.AppendChild(htmlEl)

	interp := NewInterpreter()
	interp.SetupDOM(doc)

	src := `document.getElementById("box").style.backgroundColor = "red";`
	parser := NewParser(src)
	prog, _ := parser.ParseProgram()
	if err := interp.Eval(prog); err != nil {
		t.Fatalf("eval error: %v", err)
	}

	style := div.GetAttribute("style")
	if !strings.Contains(style, "red") {
		t.Errorf("expected style to contain 'red', got %q", style)
	}
}

func TestDOMCreateElement(t *testing.T) {
	doc := dom.NewDocument()
	htmlEl := dom.NewElement("html")
	body := dom.NewElement("body")
	ul := dom.NewElement("ul")
	ul.SetAttribute("id", "list")
	body.AppendChild(ul)
	htmlEl.AppendChild(body)
	doc.AppendChild(htmlEl)

	interp := NewInterpreter()
	interp.SetupDOM(doc)

	src := `
		var list = document.getElementById("list");
		for (var i = 0; i < 3; i++) {
			var li = document.createElement("li");
			li.textContent = "Item " + i;
			list.appendChild(li);
		}
	`
	parser := NewParser(src)
	prog, _ := parser.ParseProgram()
	if err := interp.Eval(prog); err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if len(ul.Children) != 3 {
		t.Errorf("ul should have 3 children, has %d", len(ul.Children))
	}
	if len(ul.Children) > 0 && getTextContent(ul.Children[0]) != "Item 0" {
		t.Errorf("first li should be 'Item 0', got %q", getTextContent(ul.Children[0]))
	}
}

func TestDOMClassList(t *testing.T) {
	doc := dom.NewDocument()
	htmlEl := dom.NewElement("html")
	body := dom.NewElement("body")
	div := dom.NewElement("div")
	div.SetAttribute("id", "box")
	div.SetAttribute("class", "initial")
	body.AppendChild(div)
	htmlEl.AppendChild(body)
	doc.AppendChild(htmlEl)

	interp := NewInterpreter()
	interp.SetupDOM(doc)

	src := `
		var el = document.getElementById("box");
		el.classList.add("active");
		el.classList.add("highlight");
		el.classList.remove("initial");
		var hasActive = el.classList.contains("active");
		var hasInitial = el.classList.contains("initial");
	`
	parser := NewParser(src)
	prog, _ := parser.ParseProgram()
	if err := interp.Eval(prog); err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if v := getVar(interp, "hasActive"); !v.ToBoolean() {
		t.Error("should have 'active' class")
	}
	if v := getVar(interp, "hasInitial"); v.ToBoolean() {
		t.Error("should not have 'initial' class")
	}

	cls := div.GetAttribute("class")
	if !strings.Contains(cls, "active") {
		t.Errorf("class attr should contain 'active', got %q", cls)
	}
	if strings.Contains(cls, "initial") {
		t.Errorf("class attr should not contain 'initial', got %q", cls)
	}
}

func TestDOMQuerySelector(t *testing.T) {
	doc := dom.NewDocument()
	htmlEl := dom.NewElement("html")
	body := dom.NewElement("body")

	div1 := dom.NewElement("div")
	div1.SetAttribute("class", "item")
	div2 := dom.NewElement("div")
	div2.SetAttribute("class", "item active")
	div3 := dom.NewElement("div")
	div3.SetAttribute("id", "special")

	body.AppendChild(div1)
	body.AppendChild(div2)
	body.AppendChild(div3)
	htmlEl.AppendChild(body)
	doc.AppendChild(htmlEl)

	interp := NewInterpreter()
	interp.SetupDOM(doc)

	src := `
		var items = document.querySelectorAll(".item");
		var itemCount = items.length;
		var active = document.querySelector(".active");
		var hasActive = active !== null;
		var special = document.getElementById("special");
		var hasSpecial = special !== null;
	`
	parser := NewParser(src)
	prog, _ := parser.ParseProgram()
	if err := interp.Eval(prog); err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if v := getVar(interp, "itemCount"); v.ToNumber() != 2 {
		t.Errorf("itemCount = %v, want 2", v.ToNumber())
	}
	if v := getVar(interp, "hasActive"); !v.ToBoolean() {
		t.Error("should find active element")
	}
	if v := getVar(interp, "hasSpecial"); !v.ToBoolean() {
		t.Error("should find special element by id")
	}
}

// ---- Execute integration test ----

func TestExecuteModifiesDOM(t *testing.T) {
	doc := dom.NewDocument()
	htmlEl := dom.NewElement("html")
	body := dom.NewElement("body")

	heading := dom.NewElement("h1")
	heading.SetAttribute("id", "title")
	heading.AppendChild(dom.NewText("Original Title"))

	script := dom.NewElement("script")
	script.AppendChild(dom.NewText(`
		document.getElementById("title").textContent = "Modified by JS";
	`))

	body.AppendChild(heading)
	body.AppendChild(script)
	htmlEl.AppendChild(body)
	doc.AppendChild(htmlEl)

	Execute(doc)

	text := getTextContent(heading)
	if text != "Modified by JS" {
		t.Errorf("heading text = %q, want 'Modified by JS'", text)
	}
}

func TestExecuteMap(t *testing.T) {
	interp := evalScript(t, `
		var m = new Map();
		m.set("a", 1);
		m.set("b", 2);
		var a = m.get("a");
		var has = m.has("b");
		var size = m.size;
		m.delete("a");
		var size2 = m.size;
	`)
	if v := getVar(interp, "a"); v.ToNumber() != 1 {
		t.Errorf("a = %v, want 1", v.ToNumber())
	}
	if v := getVar(interp, "has"); !v.ToBoolean() {
		t.Error("has should be true")
	}
	if v := getVar(interp, "size"); v.ToNumber() != 2 {
		t.Errorf("size = %v, want 2", v.ToNumber())
	}
	if v := getVar(interp, "size2"); v.ToNumber() != 1 {
		t.Errorf("size2 = %v, want 1", v.ToNumber())
	}
}

func TestExecuteSet(t *testing.T) {
	interp := evalScript(t, `
		var s = new Set();
		s.add(1);
		s.add(2);
		s.add(1);  // duplicate
		var size = s.size;
		var has1 = s.has(1);
		var has3 = s.has(3);
	`)
	if v := getVar(interp, "size"); v.ToNumber() != 2 {
		t.Errorf("size = %v, want 2", v.ToNumber())
	}
	if v := getVar(interp, "has1"); !v.ToBoolean() {
		t.Error("has1 should be true")
	}
	if v := getVar(interp, "has3"); v.ToBoolean() {
		t.Error("has3 should be false")
	}
}

func TestSwitchStatement(t *testing.T) {
	interp := evalScript(t, `
		var x = 2;
		var result;
		switch (x) {
			case 1:
				result = "one";
				break;
			case 2:
				result = "two";
				break;
			case 3:
				result = "three";
				break;
			default:
				result = "other";
		}
	`)
	if v := getVar(interp, "result"); v.ToString() != "two" {
		t.Errorf("result = %q, want 'two'", v.ToString())
	}
}

func TestTypeOf(t *testing.T) {
	interp := evalScript(t, `
		var a = typeof 42;
		var b = typeof "hello";
		var c = typeof true;
		var d = typeof undefined;
		var e = typeof null;
		var f = typeof {};
		var g = typeof function(){};
	`)
	want := map[string]string{
		"a": "number", "b": "string", "c": "boolean",
		"d": "undefined", "e": "object", "f": "object", "g": "function",
	}
	for name, expected := range want {
		if v := getVar(interp, name); v.ToString() != expected {
			t.Errorf("typeof %s = %q, want %q", name, v.ToString(), expected)
		}
	}
}

func TestNullishCoalescing(t *testing.T) {
	interp := evalScript(t, `
		var a = null ?? "default";
		var b = undefined ?? "fallback";
		var c = 0 ?? "not this";
		var d = "" ?? "not this either";
	`)
	if v := getVar(interp, "a"); v.ToString() != "default" {
		t.Errorf("a = %q, want 'default'", v.ToString())
	}
	if v := getVar(interp, "b"); v.ToString() != "fallback" {
		t.Errorf("b = %q, want 'fallback'", v.ToString())
	}
	if v := getVar(interp, "c"); v.ToNumber() != 0 {
		t.Errorf("c = %v, want 0 (not nullish)", v)
	}
	if v := getVar(interp, "d"); v.ToString() != "" {
		t.Errorf("d = %q, want '' (not nullish)", v.ToString())
	}
}
