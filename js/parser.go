package js

import (
	"fmt"
	"strconv"
	"strings"
)

// parseError is returned when a syntax error is encountered.
type parseError struct {
	msg  string
	line int
}

func (e *parseError) Error() string {
	return fmt.Sprintf("SyntaxError at line %d: %s", e.line, e.msg)
}

// Parser converts JavaScript source into an AST.
type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
}

// NewParser creates a new parser for the given source.
func NewParser(src string) *Parser {
	p := &Parser{lexer: NewLexer(src)}
	p.cur = p.lexer.Next()
	p.peek = p.lexer.Next()
	return p
}

// ParseProgram parses a complete JavaScript program.
func (p *Parser) ParseProgram() (prog *Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*parseError); ok {
				err = pe
			} else {
				err = fmt.Errorf("parse error: %v", r)
			}
		}
	}()
	prog = &Program{}
	for p.cur.Type != tokEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Body = append(prog.Body, stmt)
		}
	}
	return
}

// ---- Token helpers ----

func (p *Parser) advance() Token {
	prev := p.cur
	p.cur = p.peek
	p.peek = p.lexer.Next()
	return prev
}

func (p *Parser) expect(t TokenType, val string) Token {
	tok := p.cur
	if tok.Type != t {
		panic(&parseError{fmt.Sprintf("expected %q, got %q", val, tok.Value), tok.Line})
	}
	return p.advance()
}

func (p *Parser) expectIdent(name string) {
	if p.cur.Type != tokIdent || p.cur.Value != name {
		panic(&parseError{fmt.Sprintf("expected %q, got %q", name, p.cur.Value), p.cur.Line})
	}
	p.advance()
}

func (p *Parser) check(t TokenType, val string) bool {
	return p.cur.Type == t && (val == "" || p.cur.Value == val)
}

// consumeSemicolon applies ASI rules and consumes a semicolon.
func (p *Parser) consumeSemicolon() {
	if p.cur.Type == tokSemicolon {
		p.advance()
		return
	}
	// ASI: the next token is on a new line, or we're at } or EOF
	if p.cur.AfterNL || p.cur.Type == tokRBrace || p.cur.Type == tokEOF {
		return
	}
}

// isKeyword checks if current token is a specific keyword.
func (p *Parser) isKeyword(kw string) bool {
	return p.cur.Type == tokIdent && p.cur.Value == kw
}

// ---- Statements ----

func (p *Parser) parseStatement() Node {
	// Empty statement
	if p.cur.Type == tokSemicolon {
		p.advance()
		return nil
	}

	if p.cur.Type == tokLBrace {
		return p.parseBlock()
	}

	if p.cur.Type == tokIdent {
		switch p.cur.Value {
		case "var", "let", "const":
			return p.parseVarDecl()
		case "function":
			return p.parseFunctionDecl()
		case "if":
			return p.parseIf()
		case "while":
			return p.parseWhile()
		case "do":
			return p.parseDoWhile()
		case "for":
			return p.parseFor()
		case "return":
			return p.parseReturn()
		case "break":
			p.advance()
			p.consumeSemicolon()
			return &BreakStatement{}
		case "continue":
			p.advance()
			p.consumeSemicolon()
			return &ContinueStatement{}
		case "throw":
			return p.parseThrow()
		case "try":
			return p.parseTryCatch()
		case "switch":
			return p.parseSwitch()
		case "class":
			return p.parseClass()
		case "async":
			// async function declaration or async expression
			if p.peek.Type == tokIdent && p.peek.Value == "function" {
				p.advance() // skip "async"
				return p.parseFunctionDecl()
			}
		}
	}

	// Expression statement
	expr := p.parseExpression()
	p.consumeSemicolon()
	return &ExpressionStatement{Expr: expr}
}

func (p *Parser) parseBlock() *BlockStatement {
	p.expect(tokLBrace, "{")
	block := &BlockStatement{}
	for p.cur.Type != tokRBrace && p.cur.Type != tokEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Body = append(block.Body, stmt)
		}
	}
	p.expect(tokRBrace, "}")
	return block
}

func (p *Parser) parseVarDecl() *VariableDeclaration {
	kind := p.cur.Value // "var", "let", "const"
	p.advance()
	decl := &VariableDeclaration{Kind: kind}
	for {
		id := p.parseIdentifier()
		var init Node
		if p.cur.Type == tokAssign {
			p.advance()
			init = p.parseAssignment()
		}
		decl.Declarators = append(decl.Declarators, VariableDeclarator{ID: id, Init: init})
		if p.cur.Type != tokComma {
			break
		}
		p.advance()
	}
	p.consumeSemicolon()
	return decl
}

func (p *Parser) parseFunctionDecl() *FunctionDeclaration {
	p.expectIdent("function")
	var id *Identifier
	if p.cur.Type == tokIdent && !keywords[p.cur.Value] {
		id = p.parseIdentifier()
	}
	params := p.parseParams()
	body := p.parseBlock()
	return &FunctionDeclaration{ID: id, Params: params, Body: body}
}

func (p *Parser) parseParams() []*Identifier {
	p.expect(tokLParen, "(")
	var params []*Identifier
	for p.cur.Type != tokRParen && p.cur.Type != tokEOF {
		// Handle rest parameter
		if p.cur.Type == tokDotDotDot {
			p.advance()
		}
		if p.cur.Type == tokIdent {
			params = append(params, p.parseIdentifier())
		}
		if p.cur.Type == tokAssign {
			// default parameter value - skip it for now
			p.advance()
			p.parseAssignment()
		}
		if p.cur.Type != tokComma {
			break
		}
		p.advance()
	}
	p.expect(tokRParen, ")")
	return params
}

func (p *Parser) parseIf() *IfStatement {
	p.expectIdent("if")
	p.expect(tokLParen, "(")
	test := p.parseExpression()
	p.expect(tokRParen, ")")
	consequent := p.parseStatement()
	var alternate Node
	if p.isKeyword("else") {
		p.advance()
		alternate = p.parseStatement()
	}
	return &IfStatement{Test: test, Consequent: consequent, Alternate: alternate}
}

func (p *Parser) parseWhile() *WhileStatement {
	p.expectIdent("while")
	p.expect(tokLParen, "(")
	test := p.parseExpression()
	p.expect(tokRParen, ")")
	body := p.parseStatement()
	return &WhileStatement{Test: test, Body: body}
}

func (p *Parser) parseDoWhile() *DoWhileStatement {
	p.expectIdent("do")
	body := p.parseStatement()
	p.expectIdent("while")
	p.expect(tokLParen, "(")
	test := p.parseExpression()
	p.expect(tokRParen, ")")
	p.consumeSemicolon()
	return &DoWhileStatement{Body: body, Test: test}
}

func (p *Parser) parseFor() Node {
	p.expectIdent("for")
	p.expect(tokLParen, "(")

	// Check for for...in / for...of
	// Possibilities: "for (var x in/of ...)" or "for (x in/of ...)"
	savedPos := p.cur

	var left Node
	if p.isKeyword("var") || p.isKeyword("let") || p.isKeyword("const") {
		kind := p.cur.Value
		p.advance()
		id := p.parseIdentifier()
		// Check if this is for...in or for...of
		if p.isKeyword("in") || p.isKeyword("of") {
			isOf := p.cur.Value == "of"
			p.advance()
			right := p.parseAssignment()
			p.expect(tokRParen, ")")
			body := p.parseStatement()
			decl := &VariableDeclaration{Kind: kind, Declarators: []VariableDeclarator{{ID: id}}}
			return &ForInStatement{Left: decl, Right: right, Body: body, Of: isOf}
		}
		// Regular for loop - reconstruct init
		var init Node
		if p.cur.Type == tokAssign {
			p.advance()
			init = p.parseAssignment()
		}
		decl := &VariableDeclaration{Kind: kind, Declarators: []VariableDeclarator{{ID: id, Init: init}}}
		// Additional declarators
		for p.cur.Type == tokComma {
			p.advance()
			id2 := p.parseIdentifier()
			var init2 Node
			if p.cur.Type == tokAssign {
				p.advance()
				init2 = p.parseAssignment()
			}
			decl.Declarators = append(decl.Declarators, VariableDeclarator{ID: id2, Init: init2})
		}
		left = decl
	} else if p.cur.Type == tokSemicolon {
		left = nil // no init
	} else {
		_ = savedPos
		expr := p.parseExpression()
		if p.isKeyword("in") || p.isKeyword("of") {
			isOf := p.cur.Value == "of"
			p.advance()
			right := p.parseAssignment()
			p.expect(tokRParen, ")")
			body := p.parseStatement()
			return &ForInStatement{Left: expr, Right: right, Body: body, Of: isOf}
		}
		left = &ExpressionStatement{Expr: expr}
	}

	p.expect(tokSemicolon, ";")
	var test Node
	if p.cur.Type != tokSemicolon {
		test = p.parseExpression()
	}
	p.expect(tokSemicolon, ";")
	var update Node
	if p.cur.Type != tokRParen {
		update = p.parseExpression()
	}
	p.expect(tokRParen, ")")
	body := p.parseStatement()
	return &ForStatement{Init: left, Test: test, Update: update, Body: body}
}

func (p *Parser) parseReturn() *ReturnStatement {
	p.expectIdent("return")
	// ASI: if next token is on a new line or ; or }, no argument
	var arg Node
	if !p.cur.AfterNL && p.cur.Type != tokSemicolon && p.cur.Type != tokRBrace && p.cur.Type != tokEOF {
		arg = p.parseExpression()
	}
	p.consumeSemicolon()
	return &ReturnStatement{Argument: arg}
}

func (p *Parser) parseThrow() *ThrowStatement {
	p.expectIdent("throw")
	arg := p.parseExpression()
	p.consumeSemicolon()
	return &ThrowStatement{Argument: arg}
}

func (p *Parser) parseTryCatch() *TryCatchStatement {
	p.expectIdent("try")
	block := p.parseBlock()
	var param *Identifier
	var handler *BlockStatement
	var finally *BlockStatement
	if p.isKeyword("catch") {
		p.advance()
		if p.cur.Type == tokLParen {
			p.advance()
			if p.cur.Type == tokIdent {
				param = p.parseIdentifier()
			}
			p.expect(tokRParen, ")")
		}
		handler = p.parseBlock()
	}
	if p.isKeyword("finally") {
		p.advance()
		finally = p.parseBlock()
	}
	return &TryCatchStatement{Block: block, Param: param, Handler: handler, Finally: finally}
}

func (p *Parser) parseSwitch() *SwitchStatement {
	p.expectIdent("switch")
	p.expect(tokLParen, "(")
	discriminant := p.parseExpression()
	p.expect(tokRParen, ")")
	p.expect(tokLBrace, "{")
	var cases []SwitchCase
	for p.cur.Type != tokRBrace && p.cur.Type != tokEOF {
		var test Node
		if p.isKeyword("case") {
			p.advance()
			test = p.parseExpression()
			p.expect(tokColon, ":")
		} else if p.isKeyword("default") {
			p.advance()
			p.expect(tokColon, ":")
		} else {
			p.advance() // skip unknown
			continue
		}
		var consequent []Node
		for p.cur.Type != tokRBrace && !p.isKeyword("case") && !p.isKeyword("default") && p.cur.Type != tokEOF {
			stmt := p.parseStatement()
			if stmt != nil {
				consequent = append(consequent, stmt)
			}
		}
		cases = append(cases, SwitchCase{Test: test, Consequent: consequent})
	}
	p.expect(tokRBrace, "}")
	return &SwitchStatement{Discriminant: discriminant, Cases: cases}
}

// parseClass parses a class declaration (simplified - treats methods as properties).
func (p *Parser) parseClass() Node {
	p.expectIdent("class")
	var name string
	if p.cur.Type == tokIdent && !keywords[p.cur.Value] {
		name = p.cur.Value
		p.advance()
	}
	var superClass Node
	if p.isKeyword("extends") {
		p.advance()
		superClass = p.parseLeftHandSide()
	}
	_ = superClass

	// Parse class body into a constructor function + prototype methods
	p.expect(tokLBrace, "{")
	var constructor *FunctionDeclaration
	methods := map[string]*FunctionExpression{}

	for p.cur.Type != tokRBrace && p.cur.Type != tokEOF {
		isStatic := false
		if p.isKeyword("static") {
			isStatic = true
			p.advance()
		}
		_ = isStatic

		// async/get/set modifiers
		for p.isKeyword("async") || p.isKeyword("get") || p.isKeyword("set") {
			p.advance()
		}

		// Method name
		var methodName string
		if p.cur.Type == tokIdent || p.cur.Type == tokString {
			methodName = p.cur.Value
			p.advance()
		} else if p.cur.Type == tokLBracket {
			// computed method name - skip
			p.advance()
			p.parseAssignment()
			p.expect(tokRBracket, "]")
			methodName = "__computed__"
		} else {
			p.advance()
			continue
		}

		// Field initializer (no parens) or method
		if p.cur.Type == tokAssign {
			p.advance()
			p.parseAssignment()
			p.consumeSemicolon()
			continue
		}
		if p.cur.Type == tokSemicolon || p.cur.AfterNL {
			p.consumeSemicolon()
			continue
		}

		params := p.parseParams()
		body := p.parseBlock()
		fn := &FunctionExpression{Params: params, Body: body}
		if methodName == "constructor" {
			constructor = &FunctionDeclaration{
				ID:     &Identifier{Name: name},
				Params: params,
				Body:   body,
			}
		} else {
			methods[methodName] = fn
		}
	}
	p.expect(tokRBrace, "}")

	if constructor == nil {
		constructor = &FunctionDeclaration{
			ID:     &Identifier{Name: name},
			Params: nil,
			Body:   &BlockStatement{},
		}
	}

	// Attach methods to prototype by generating:
	//   function Name(...) { ... }
	//   Name.prototype.method = function(...) { ... };
	stmts := []Node{constructor}
	for mName, mFn := range methods {
		assign := &ExpressionStatement{
			Expr: &AssignmentExpression{
				Op: "=",
				Left: &MemberExpression{
					Object: &MemberExpression{
						Object:   &Identifier{Name: name},
						Property: &Identifier{Name: "prototype"},
					},
					Property: &Identifier{Name: mName},
				},
				Right: mFn,
			},
		}
		stmts = append(stmts, assign)
	}

	// Wrap in a block if multiple statements
	if len(stmts) == 1 {
		return constructor
	}
	return &StatementList{Body: stmts}
}

// ---- Expressions ----

func (p *Parser) parseExpression() Node {
	expr := p.parseAssignment()
	if p.cur.Type == tokComma {
		exprs := []Node{expr}
		for p.cur.Type == tokComma {
			p.advance()
			exprs = append(exprs, p.parseAssignment())
		}
		return &SequenceExpression{Expressions: exprs}
	}
	return expr
}

func (p *Parser) parseAssignment() Node {
	left := p.parseConditional()

	switch p.cur.Type {
	case tokAssign, tokPlusAssign, tokMinusAssign, tokStarAssign, tokSlashAssign,
		tokPercentAssign, tokAmpAssign, tokPipeAssign, tokCaretAssign:
		op := p.cur.Value
		p.advance()
		right := p.parseAssignment()
		return &AssignmentExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseConditional() Node {
	expr := p.parseNullCoalescing()
	if p.cur.Type == tokQuestion {
		p.advance()
		consequent := p.parseAssignment()
		p.expect(tokColon, ":")
		alternate := p.parseAssignment()
		return &ConditionalExpression{Test: expr, Consequent: consequent, Alternate: alternate}
	}
	return expr
}

func (p *Parser) parseNullCoalescing() Node {
	left := p.parseOr()
	for p.cur.Type == tokNullCoal {
		op := p.cur.Value
		p.advance()
		right := p.parseOr()
		left = &LogicalExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseOr() Node {
	left := p.parseAnd()
	for p.cur.Type == tokOr {
		op := p.cur.Value
		p.advance()
		right := p.parseAnd()
		left = &LogicalExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() Node {
	left := p.parseBitwiseOr()
	for p.cur.Type == tokAnd {
		op := p.cur.Value
		p.advance()
		right := p.parseBitwiseOr()
		left = &LogicalExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseBitwiseOr() Node {
	left := p.parseBitwiseXor()
	for p.cur.Type == tokPipe {
		op := p.cur.Value
		p.advance()
		right := p.parseBitwiseXor()
		left = &BinaryExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseBitwiseXor() Node {
	left := p.parseBitwiseAnd()
	for p.cur.Type == tokCaret {
		op := p.cur.Value
		p.advance()
		right := p.parseBitwiseAnd()
		left = &BinaryExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseBitwiseAnd() Node {
	left := p.parseEquality()
	for p.cur.Type == tokAmp {
		op := p.cur.Value
		p.advance()
		right := p.parseEquality()
		left = &BinaryExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseEquality() Node {
	left := p.parseRelational()
	for p.cur.Type == tokEqEq || p.cur.Type == tokEqEqEq || p.cur.Type == tokNeq || p.cur.Type == tokNeqEq {
		op := p.cur.Value
		p.advance()
		right := p.parseRelational()
		left = &BinaryExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseRelational() Node {
	left := p.parseShift()
	for {
		if p.cur.Type == tokLt || p.cur.Type == tokGt || p.cur.Type == tokLte || p.cur.Type == tokGte {
			op := p.cur.Value
			p.advance()
			right := p.parseShift()
			left = &BinaryExpression{Op: op, Left: left, Right: right}
		} else if p.isKeyword("instanceof") || p.isKeyword("in") {
			op := p.cur.Value
			p.advance()
			right := p.parseShift()
			left = &BinaryExpression{Op: op, Left: left, Right: right}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseShift() Node {
	left := p.parseAdditive()
	for p.cur.Type == tokLShift || p.cur.Type == tokRShift {
		op := p.cur.Value
		p.advance()
		right := p.parseAdditive()
		left = &BinaryExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAdditive() Node {
	left := p.parseMultiplicative()
	for p.cur.Type == tokPlus || p.cur.Type == tokMinus {
		op := p.cur.Value
		p.advance()
		right := p.parseMultiplicative()
		left = &BinaryExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplicative() Node {
	left := p.parseExponentiation()
	for p.cur.Type == tokStar || p.cur.Type == tokSlash || p.cur.Type == tokPercent {
		op := p.cur.Value
		p.advance()
		right := p.parseExponentiation()
		left = &BinaryExpression{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseExponentiation() Node {
	base := p.parseUnary()
	if p.cur.Type == tokStarStar {
		p.advance()
		exp := p.parseExponentiation()
		return &BinaryExpression{Op: "**", Left: base, Right: exp}
	}
	return base
}

func (p *Parser) parseUnary() Node {
	switch p.cur.Type {
	case tokNot, tokMinus, tokPlus, tokTilde:
		op := p.cur.Value
		p.advance()
		arg := p.parseUnary()
		return &UnaryExpression{Op: op, Argument: arg, Prefix: true}
	case tokPlusPlus, tokMinusMinus:
		op := p.cur.Value
		p.advance()
		arg := p.parseLeftHandSide()
		return &UpdateExpression{Op: op, Argument: arg, Prefix: true}
	}
	if p.cur.Type == tokIdent {
		switch p.cur.Value {
		case "typeof":
			p.advance()
			arg := p.parseUnary()
			return &UnaryExpression{Op: "typeof", Argument: arg, Prefix: true}
		case "void":
			p.advance()
			arg := p.parseUnary()
			return &UnaryExpression{Op: "void", Argument: arg, Prefix: true}
		case "delete":
			p.advance()
			arg := p.parseUnary()
			return &UnaryExpression{Op: "delete", Argument: arg, Prefix: true}
		case "await":
			p.advance()
			arg := p.parseUnary()
			return arg // treat await as a no-op wrapper
		}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() Node {
	expr := p.parseCallExpr()
	if !p.cur.AfterNL {
		if p.cur.Type == tokPlusPlus {
			p.advance()
			return &UpdateExpression{Op: "++", Argument: expr, Prefix: false}
		}
		if p.cur.Type == tokMinusMinus {
			p.advance()
			return &UpdateExpression{Op: "--", Argument: expr, Prefix: false}
		}
	}
	return expr
}

func (p *Parser) parseCallExpr() Node {
	expr := p.parseNewExpr()
	for {
		if p.cur.Type == tokLParen {
			args := p.parseArguments()
			expr = &CallExpression{Callee: expr, Arguments: args}
		} else if p.cur.Type == tokDot {
			p.advance()
			prop := p.parseIdentifier()
			expr = &MemberExpression{Object: expr, Property: prop}
		} else if p.cur.Type == tokOptChain {
			p.advance()
			prop := p.parseIdentifier()
			expr = &MemberExpression{Object: expr, Property: prop}
		} else if p.cur.Type == tokLBracket {
			p.advance()
			prop := p.parseExpression()
			p.expect(tokRBracket, "]")
			expr = &MemberExpression{Object: expr, Property: prop, Computed: true}
		} else {
			break
		}
	}
	return expr
}

func (p *Parser) parseNewExpr() Node {
	if p.isKeyword("new") {
		p.advance()
		// Handle "new new Foo()" (rare but valid)
		callee := p.parseNewExpr()
		var args []Node
		if p.cur.Type == tokLParen {
			args = p.parseArguments()
		}
		return &NewExpression{Callee: callee, Arguments: args}
	}
	return p.parseLeftHandSide()
}

func (p *Parser) parseLeftHandSide() Node {
	return p.parsePrimary()
}

func (p *Parser) parseArguments() []Node {
	p.expect(tokLParen, "(")
	var args []Node
	for p.cur.Type != tokRParen && p.cur.Type != tokEOF {
		if p.cur.Type == tokDotDotDot {
			p.advance()
			arg := p.parseAssignment()
			args = append(args, &SpreadElement{Argument: arg})
		} else {
			args = append(args, p.parseAssignment())
		}
		if p.cur.Type != tokComma {
			break
		}
		p.advance()
	}
	p.expect(tokRParen, ")")
	return args
}

func (p *Parser) parsePrimary() Node {
	tok := p.cur

	switch tok.Type {
	case tokNumber:
		p.advance()
		var v float64
		if strings.HasPrefix(tok.Value, "0x") || strings.HasPrefix(tok.Value, "0X") {
			n, _ := strconv.ParseInt(tok.Value[2:], 16, 64)
			v = float64(n)
		} else {
			v, _ = strconv.ParseFloat(tok.Value, 64)
		}
		return &Literal{Value: v, Raw: tok.Value}

	case tokString:
		p.advance()
		return &Literal{Value: tok.Value, Raw: tok.Value}

	case tokTemplate:
		p.advance()
		return p.parseTemplateLiteral(tok.Value)

	case tokLParen:
		p.advance()
		// Check for arrow function: () => or (a, b) =>
		// Heuristic: parse contents, if we see => it's arrow function
		if p.cur.Type == tokRParen {
			p.advance()
			if p.cur.Type == tokFatArrow {
				p.advance()
				return p.parseArrowBody(nil)
			}
			// empty parens - probably an error, return undefined
			return &Literal{Value: nil}
		}
		// Try to detect arrow function by looking for =>
		// Simple approach: if first token inside is an ident followed by ) =>, treat as arrow
		if p.cur.Type == tokIdent && !keywords[p.cur.Value] && p.peek.Type == tokRParen {
			// Could be (x) => ...
			id := p.parseIdentifier()
			p.expect(tokRParen, ")")
			if p.cur.Type == tokFatArrow {
				p.advance()
				return p.parseArrowBody([]*Identifier{id})
			}
			// Otherwise it's just (x) - but we consumed the tokens already
			// Reconstruct as expression
			return id
		}
		expr := p.parseExpression()
		p.expect(tokRParen, ")")
		// Check for arrow after (a, b)
		if p.cur.Type == tokFatArrow {
			p.advance()
			// expr should be SequenceExpression of identifiers
			params := seqToParams(expr)
			return p.parseArrowBody(params)
		}
		return expr

	case tokLBrace:
		return p.parseObjectLiteral()

	case tokLBracket:
		return p.parseArrayLiteral()

	case tokIdent:
		switch tok.Value {
		case "true":
			p.advance()
			return &Literal{Value: true, Raw: "true"}
		case "false":
			p.advance()
			return &Literal{Value: false, Raw: "false"}
		case "null":
			p.advance()
			return &Literal{Value: nil, Raw: "null"}
		case "undefined":
			p.advance()
			return &Literal{Value: jsUndefined, Raw: "undefined"}
		case "this":
			p.advance()
			return &ThisExpression{}
		case "function":
			return p.parseFunctionExpression()
		case "async":
			// async () => ... or async function
			p.advance()
			if p.isKeyword("function") {
				return p.parseFunctionExpression()
			}
			// async arrow: async (params) => body or async x => body
			if p.cur.Type == tokLParen {
				p.advance()
				var params []*Identifier
				for p.cur.Type != tokRParen && p.cur.Type != tokEOF {
					if p.cur.Type == tokDotDotDot {
						p.advance()
					}
					if p.cur.Type == tokIdent {
						params = append(params, p.parseIdentifier())
					}
					if p.cur.Type != tokComma {
						break
					}
					p.advance()
				}
				p.expect(tokRParen, ")")
				p.expect(tokFatArrow, "=>")
				ids := make([]Node, len(params))
				for i, pa := range params {
					ids[i] = pa
				}
				return p.parseArrowBody(params)
			}
			if p.cur.Type == tokIdent {
				id := p.parseIdentifier()
				p.expect(tokFatArrow, "=>")
				return p.parseArrowBody([]*Identifier{id})
			}
			return &Identifier{Name: "async"}
		}
		id := p.parseIdentifier()
		// Arrow function: x => body
		if p.cur.Type == tokFatArrow {
			p.advance()
			return p.parseArrowBody([]*Identifier{id})
		}
		return id

	default:
		p.advance() // skip unknown token
		return &Literal{Value: nil}
	}
}

// jsUndefined is a sentinel for the `undefined` literal.
var jsUndefined = struct{}{}

func (p *Parser) parseIdentifier() *Identifier {
	tok := p.cur
	if tok.Type != tokIdent {
		panic(&parseError{fmt.Sprintf("expected identifier, got %q", tok.Value), tok.Line})
	}
	p.advance()
	return &Identifier{Name: tok.Value}
}

func (p *Parser) parseFunctionExpression() *FunctionExpression {
	if p.isKeyword("function") {
		p.advance()
	}
	var id *Identifier
	if p.cur.Type == tokIdent && !keywords[p.cur.Value] {
		id = p.parseIdentifier()
	}
	params := p.parseParams()
	body := p.parseBlock()
	return &FunctionExpression{ID: id, Params: params, Body: body}
}

func (p *Parser) parseArrowBody(params []*Identifier) *ArrowFunctionExpression {
	var ids []Node
	for _, param := range params {
		ids = append(ids, param)
	}
	var body Node
	if p.cur.Type == tokLBrace {
		body = p.parseBlock()
	} else {
		body = p.parseAssignment()
	}
	return &ArrowFunctionExpression{Params: ids, Body: body}
}

func (p *Parser) parseObjectLiteral() *ObjectExpression {
	p.expect(tokLBrace, "{")
	obj := &ObjectExpression{}
	for p.cur.Type != tokRBrace && p.cur.Type != tokEOF {
		// Spread: ...expr
		if p.cur.Type == tokDotDotDot {
			p.advance()
			p.parseAssignment() // skip spread for now
			if p.cur.Type == tokComma {
				p.advance()
			}
			continue
		}

		// Method shorthand, getter/setter, or computed key
		isComputed := false
		var key Node
		if p.cur.Type == tokLBracket {
			p.advance()
			key = p.parseAssignment()
			p.expect(tokRBracket, "]")
			isComputed = true
		} else if p.cur.Type == tokIdent || p.cur.Type == tokString || p.cur.Type == tokNumber {
			keyTok := p.cur
			p.advance()
			if keyTok.Type == tokNumber {
				v, _ := strconv.ParseFloat(keyTok.Value, 64)
				key = &Literal{Value: v, Raw: keyTok.Value}
			} else {
				key = &Identifier{Name: keyTok.Value}
			}
		} else {
			p.advance() // skip
			continue
		}
		_ = isComputed

		// Shorthand method: { foo() { } }
		if p.cur.Type == tokLParen {
			params := p.parseParams()
			body := p.parseBlock()
			fn := &FunctionExpression{Params: params, Body: body}
			obj.Properties = append(obj.Properties, Property{Key: key, Value: fn})
		} else if p.cur.Type == tokColon {
			p.advance()
			val := p.parseAssignment()
			obj.Properties = append(obj.Properties, Property{Key: key, Value: val})
		} else {
			// Shorthand property: { x } => { x: x }
			if ident, ok := key.(*Identifier); ok {
				obj.Properties = append(obj.Properties, Property{Key: key, Value: &Identifier{Name: ident.Name}, Shorthand: true})
			}
		}

		if p.cur.Type != tokComma {
			break
		}
		p.advance()
	}
	p.expect(tokRBrace, "}")
	return obj
}

func (p *Parser) parseArrayLiteral() *ArrayExpression {
	p.expect(tokLBracket, "[")
	arr := &ArrayExpression{}
	for p.cur.Type != tokRBracket && p.cur.Type != tokEOF {
		if p.cur.Type == tokComma {
			arr.Elements = append(arr.Elements, nil) // hole
			p.advance()
			continue
		}
		if p.cur.Type == tokDotDotDot {
			p.advance()
			elem := p.parseAssignment()
			arr.Elements = append(arr.Elements, &SpreadElement{Argument: elem})
		} else {
			arr.Elements = append(arr.Elements, p.parseAssignment())
		}
		if p.cur.Type != tokComma {
			break
		}
		p.advance()
	}
	p.expect(tokRBracket, "]")
	return arr
}

// parseTemplateLiteral splits raw template text into quasis and expressions.
// The raw value looks like: "hello ${name}, world" (without backticks).
func (p *Parser) parseTemplateLiteral(raw string) *TemplateLiteral {
	tl := &TemplateLiteral{}
	rest := raw
	for {
		idx := strings.Index(rest, "${")
		if idx < 0 {
			tl.Quasis = append(tl.Quasis, unescapeTemplate(rest))
			break
		}
		tl.Quasis = append(tl.Quasis, unescapeTemplate(rest[:idx]))
		rest = rest[idx+2:]
		// Find matching }
		depth := 1
		i := 0
		for i < len(rest) && depth > 0 {
			if rest[i] == '{' {
				depth++
			} else if rest[i] == '}' {
				depth--
			}
			if depth > 0 {
				i++
			}
		}
		exprSrc := rest[:i]
		rest = rest[i+1:]
		exprParser := NewParser(exprSrc)
		expr, err := exprParser.ParseProgram()
		if err == nil && len(expr.Body) > 0 {
			if es, ok := expr.Body[0].(*ExpressionStatement); ok {
				tl.Expressions = append(tl.Expressions, es.Expr)
			} else {
				tl.Expressions = append(tl.Expressions, &Literal{Value: ""})
			}
		} else {
			tl.Expressions = append(tl.Expressions, &Literal{Value: ""})
		}
	}
	return tl
}

func unescapeTemplate(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\`", "`")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// seqToParams extracts identifier params from a SequenceExpression or single expression.
func seqToParams(expr Node) []*Identifier {
	switch e := expr.(type) {
	case *SequenceExpression:
		var ids []*Identifier
		for _, ex := range e.Expressions {
			if id, ok := ex.(*Identifier); ok {
				ids = append(ids, id)
			}
		}
		return ids
	case *Identifier:
		return []*Identifier{e}
	}
	return nil
}
